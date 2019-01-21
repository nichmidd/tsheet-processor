package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
	//"github.com/nichmidd/tsheet-processor/db"
)

//TimesheetResults : 1
type TimesheetResults struct {
	Results  map[string]Timesheets
	More     bool
	SuppData *SupData `json:"supplemental_data"`
}

//Timesheets : 2
type Timesheets map[string]TSheets

//SupData : 3
type SupData struct {
	JobCodes map[int]Jobcodes
	Users    map[int]Users
}

//TSheets : 4
type TSheets struct {
	ID       int
	UserID   int `json:"user_id"`
	JobCode  int `json:"jobcode_id"`
	Start    string
	End      string
	Duration int
	Date     string
}

//Jobcodes : 5
type Jobcodes struct {
	ID   int
	Name string
}

//Users : 6
type Users struct {
	ID        int
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string
}

// TSheetPages : does the actual fetching
func TSheetPages(bearertok string, url string, jobs *JobResults, debug bool) (bool, error) {
	// debug
	if debug {
		fmt.Fprintf(os.Stdout, "Fetching with URL: %s\n", url)
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	var authtok bytes.Buffer
	fmt.Fprintf(&authtok, "Bearer %s", bearertok)
	req.Header.Add("Authorization", authtok.String())
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "get failed: %v\n", err)
		return false, err
	}
	defer resp.Body.Close()

	var result TimesheetResults
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Fprintf(os.Stderr, "Decode failed: %v\n", err)
		return false, err
	}

	for _, item := range result.Results {
		for _, ts := range item {
			if result.SuppData.JobCodes[ts.JobCode].Name == "Lunch Break" {
				continue
			}
			co := Contractor{ID: result.SuppData.Users[ts.UserID].ID, FirstName: result.SuppData.Users[ts.UserID].FirstName, LastName: result.SuppData.Users[ts.UserID].LastName}
			if jobs.Contractors == nil {
				jobs.Contractors = make(map[int]Contractor)
			}
			jobs.Contractors[co.ID] = co

			cl := Client{ID: result.SuppData.JobCodes[ts.JobCode].ID, Name: result.SuppData.JobCodes[ts.JobCode].Name}
			if jobs.Clients == nil {
				jobs.Clients = make(map[int]Client)
			}
			jobs.Clients[cl.ID] = cl

			// FIX ME
			// this really needs to be its own function (or set of functions)
			//

			//calculate rounding - we round to 15min increments
			timeRounding := 15 * time.Minute
			//fetch the raw start time and format it
			rawStart, _ := time.Parse(time.RFC3339, ts.Start)
			//round the start time to nearest 15min
			start := rawStart.Round(timeRounding)
			//fetch the raw end time and format it
			rawEnd, _ := time.Parse(time.RFC3339, ts.End)
			//round the end time to nearest 15min
			end := rawEnd.Round(timeRounding)
			//if start == end then add 15min to end - this is our minimum charge time
			if start == end {
				end = end.Add(timeRounding)
			}
			//format the start day
			date, _ := time.Parse("2006-01-02", ts.Date)
			//calculate job duration from rounded start and end times
			var roundedDuration = end.Sub(start)
			//convert duration to decimal value
			dur := float64(float64(int64((float64(roundedDuration.Seconds())/60/60)*4)) / 4)
			//create map and post it
			jo := Job{ID: ts.ID, UserID: ts.UserID, ClientID: ts.JobCode, Start: start, End: end, Duration: dur, Date: date}
			//debug
			if debug {
				fmt.Fprintf(os.Stdout, "Date: %s\tID: %d\tStart: %s\tEnd: %s\tDur: %f\n", ts.Date, ts.ID, start, end, dur)
			}

			if jobs.Jobs == nil {
				jobs.Jobs = make(map[int]Job)
			}
			jobs.Jobs[ts.ID] = jo
		}
	}
	return result.More, nil
}
