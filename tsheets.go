package main

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/nichmidd/tsheet-processor/db"
	"github.com/nichmidd/tsheet-processor/fetch"
)

const rooturl = "https://rest.tsheets.com/api/v1/timesheets?"

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Not enough arguments\nYou need start date: 2001-01-01 and end date: 2001-01-08")
	}

	var more = true
	var page = 1
	var querystartdate = os.Args[1]
	var queryenddate = os.Args[2]
	var jobs db.JobResults

	for more {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "%s&start_date=%s&end_date=%s&page=%d", rooturl, querystartdate, queryenddate, page)
		var url = buf.String()
		res, err := fetch.TSheetPages(url, &jobs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		page += 1
		more = res
	}

	_, err := db.PushToDB(&jobs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
