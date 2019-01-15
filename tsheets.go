package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	//"github.com/nichmidd/tsheet-processor/db"
	//"github.com/nichmidd/tsheet-processor/fetch"
)

const rooturl = "https://rest.tsheets.com/api/v1/timesheets?"

//Configuration : structure for parseing config.json file
type Configuration struct {
	Development struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Database string `json:"database"`
		Host     string `json:"host"`
		Dialect  string `json:"dialect"`
		Bearer   string `json:"bearer"`
	}
	Test struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Database string `json:"database"`
		Host     string `json:"host"`
		Dialect  string `json:"dialect"`
		Bearer   string `json:"bearer"`
	}
	Production struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Database string `json:"database"`
		Host     string `json:"host"`
		Dialect  string `json:"dialect"`
		Bearer   string `json:"bearer"`
	}
}

func main() {

	var querystartdate string
	var queryenddate string

	var b bytes.Buffer
	oneDay := time.Hour * -24
	t := time.Now().Add(oneDay)
	fmt.Fprintf(&b, t.Format("2006-01-02"))
	today := b.String()

	c := flag.String("c", "./config.json", "specify configuration file")
	flagstartdate := flag.String("start", "today", "start day of query: 2006-01-02")
	flagenddate := flag.String("end", "today", "end day of query: 2006-01-02")
	flag.Parse()
	if *flagstartdate == "today" {
		querystartdate = today
	} else {
		querystartdate = *flagstartdate
	}
	if *flagenddate == "today" {
		queryenddate = today
	} else {
		queryenddate = *flagenddate
	}
	configFile, err := os.Open(*c)
	defer configFile.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	decoder := json.NewDecoder(configFile)
	Config := Configuration{}
	err = decoder.Decode(&Config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var more = true
	var page = 1
	var jobs JobResults

	for more {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "%s&start_date=%s&end_date=%s&page=%d", rooturl, querystartdate, queryenddate, page)
		var url = buf.String()
		res, err := TSheetPages(Config.Production.Bearer, url, &jobs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		page += 1
		more = res
	}
	_, err = PushToDB(Config.Production.Username, Config.Production.Password, Config.Production.Host, &jobs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
