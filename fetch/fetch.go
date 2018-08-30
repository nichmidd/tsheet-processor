package fetch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/nichmidd/tsheet-processor/db"
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
func TSheetPages(bearertok string, url string, jobs *db.JobResults) (bool, error) {
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
			co := db.Contractor{ID: result.SuppData.Users[ts.UserID].ID, FirstName: result.SuppData.Users[ts.UserID].FirstName, LastName: result.SuppData.Users[ts.UserID].LastName}
			if jobs.Contractors == nil {
				jobs.Contractors = make(map[int]db.Contractor)
			}
			jobs.Contractors[co.ID] = co

			cl := db.Client{ID: result.SuppData.JobCodes[ts.JobCode].ID, Name: result.SuppData.JobCodes[ts.JobCode].Name}
			if jobs.Clients == nil {
				jobs.Clients = make(map[int]db.Client)
			}
			jobs.Clients[cl.ID] = cl

			timeRounding := 15 * time.Minute
			rawStart, _ := time.Parse(time.RFC3339, ts.Start)
			//fmt.Fprintf(os.Stdout, "%s\n", rawStart)
			start := rawStart.Round(timeRounding)
			//fmt.Fprintf(os.Stdout, "%s\n", start)
			rawEnd, _ := time.Parse(time.RFC3339, ts.End)
			//fmt.Fprintf(os.Stdout, "%s\n", rawEnd)
			end := rawEnd.Round(timeRounding)
			//fmt.Fprintf(os.Stdout, "%s\n", end)
			date, _ := time.Parse("2006-01-02", ts.Date)
			dur := (math.Round((float64(ts.Duration)/60/60)*100) / 100)
			jo := db.Job{ID: ts.ID, UserID: ts.UserID, ClientID: ts.JobCode, Start: start, End: end, Duration: dur, Date: date}
			if jobs.Jobs == nil {
				jobs.Jobs = make(map[int]db.Job)
			}
			jobs.Jobs[ts.ID] = jo
		}
	}
	return result.More, nil
}
