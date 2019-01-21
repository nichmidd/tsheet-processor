package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"
)

// FIX ME
// this should be another part of the config file to allow testing/dev to not hit tsheets prod API
//
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
	var debug bool

	var bufStartDate bytes.Buffer
	var bufEndDate bytes.Buffer
	//set today to 'yesterday' - no point in getting 'todays' data
	oneDay := time.Hour * -24
	//now get yesterday 6 weeks ago
	sixWeeks := time.Hour * -1008
	t := time.Now().Add(oneDay)
	tSixWeeks := t.Add(sixWeeks)
	fmt.Fprintf(&bufEndDate, t.Format("2006-01-02"))
	today := bufEndDate.String()
	fmt.Fprintf(&bufStartDate, tSixWeeks.Format("2006-01-02"))
	todaySixWeeks := bufStartDate.String()

	c := flag.String("c", "./config/config.json", "specify configuration file")
	flagstartdate := flag.String("start", "today", "start day of query: 2006-01-02")
	flagenddate := flag.String("end", "today", "end day of query: 2006-01-02")
	flagDebug := flag.Bool("d", false, "enable debug")
	flag.Parse()
	if *flagstartdate == "today" {
		querystartdate = todaySixWeeks
	} else {
		querystartdate = *flagstartdate
	}
	if *flagenddate == "today" {
		queryenddate = today
	} else {
		queryenddate = *flagenddate
	}
	if *flagDebug {
		debug = true
	}

	// debug output
	if debug {
		fmt.Fprintf(os.Stdout, "StartDate: %s\tEndDate: %s\n", querystartdate, queryenddate)
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

	// support Dev,Test & Prod environments
	type EnvironmentConfig struct {
		Username string
		Password string
		Database string
		Host     string
		Dialect  string
		Bearer   string
	}
	EnvConfig := EnvironmentConfig{}
	environment := os.Getenv("BUILD_ENVIRONMENT")
	switch {
	case environment == "DEV":
		EnvConfig.Username = Config.Development.Username
		EnvConfig.Password = Config.Development.Password
		EnvConfig.Host = Config.Development.Host
		EnvConfig.Database = Config.Development.Database
		EnvConfig.Bearer = Config.Development.Bearer
	case environment == "TEST":
		EnvConfig.Username = Config.Test.Username
		EnvConfig.Password = Config.Test.Password
		EnvConfig.Host = Config.Test.Host
		EnvConfig.Database = Config.Test.Database
		EnvConfig.Bearer = Config.Test.Bearer
	default:
		EnvConfig.Username = Config.Production.Username
		EnvConfig.Password = Config.Production.Password
		EnvConfig.Host = Config.Production.Host
		EnvConfig.Database = Config.Production.Database
		EnvConfig.Bearer = Config.Production.Bearer
	}
	//debug
	if debug {
		fmt.Fprintf(os.Stdout, "Running in %s environment\n", environment)
	}
	envBearerTok := os.Getenv("BEARERTOK")
	if len(envBearerTok) > 0 {
		//debug
		if debug {
			fmt.Fprintf(os.Stdout, "Found Bearer Token as Environment Variable\n")
		}
		EnvConfig.Bearer = envBearerTok
	}

	var more = true
	var page = 1
	var jobs JobResults

	for more {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "%s&start_date=%s&end_date=%s&page=%d", rooturl, querystartdate, queryenddate, page)
		var url = buf.String()
		res, err := TSheetPages(EnvConfig.Bearer, url, &jobs, debug)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		page += 1
		more = res
	}
	_, err = PushToDB(EnvConfig.Username, EnvConfig.Password, EnvConfig.Host, EnvConfig.Database, &jobs, debug)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
