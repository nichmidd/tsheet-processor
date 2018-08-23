package main

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/nichmidd/tsheet-processor/db"
	"github.com/nichmidd/tsheet-processor/fetch"
)

const rooturl = "https://rest.tsheets.com/api/v1/timesheets?"

func main() {

	var querystartdate string
	var queryenddate string

	if len(os.Args) == 3 {
		querystartdate = os.Args[1]
		queryenddate = os.Args[2]
	} else {
		var b bytes.Buffer
		oneDay := time.Hour * -24
		t := time.Now().Add(oneDay)
		fmt.Fprintf(&b, t.Format("2006-01-02"))
		querystartdate = b.String()
		queryenddate = b.String()
		fmt.Printf("querystartdate: %s\nqueryenddate: %s\n", querystartdate, queryenddate)
	}

	var more = true
	var page = 1
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
