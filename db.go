package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"time"

	// blind importing due to nature of package
	_ "github.com/go-sql-driver/mysql"
)

//JobResults : 1
type JobResults struct {
	Jobs        map[int]Job
	Contractors map[int]Contractor
	Clients     map[int]Client
}

//Job : 1.1
type Job struct {
	ID       int
	UserID   int
	ClientID int
	Start    time.Time
	End      time.Time
	Duration float64
	Date     time.Time
}

//Client : 1.2
type Client struct {
	ID   int
	Name string
}

//Contractor : 1.3
type Contractor struct {
	ID        int
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

//PushToDB : writes stuff to a database
func PushToDB(dbuser string, dbpass string, dbhost string, dbName string, req *JobResults, debug bool) (bool, error) {
	var dsn bytes.Buffer
	fmt.Fprintf(&dsn, "%s:%s@tcp(%s)/%s", dbuser, dbpass, dbhost, dbName)

	db, err := sql.Open("mysql", dsn.String())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to DB: %v\n", err)
		os.Exit(1)
	}

	defer db.Close()

	err = db.Ping()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	clientinsstmt, err := db.Prepare("insert clients set id=?,name=?")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Prepare SQL failed: %v\n", err)
		return false, err
	}

	rows, err := db.Query("select id from clients")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query failed: %v\n", err)
		return false, err
	}

	currentclients := make(map[int]int)

	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Scan failed: %v\n", err)
			return false, err
		}
		currentclients[id] = id
	}

	for _, cl := range req.Clients {
		if cl.ID != currentclients[cl.ID] {
			//debug
			if debug {
				fmt.Fprintf(os.Stdout, "Adding new client: %d\t%s\n", cl.ID, cl.Name)
			}
			_, err := clientinsstmt.Exec(cl.ID, cl.Name)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Insert failed: %v\n", err)
				return false, err
			}
		}
	}

	contractorinsstmt, err := db.Prepare("insert contractors set id=?,fn=?,ln=?")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Prepare SQL failed: %v\n", err)
		return false, err
	}

	rows, err = db.Query("select id from contractors")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query failed: %v\n", err)
		return false, err
	}

	currentcontractors := make(map[int]int)
	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Scan failed: %v\n", err)
			return false, err
		}
		currentcontractors[id] = id
	}

	for _, co := range req.Contractors {
		if co.ID != currentcontractors[co.ID] {
			//debug
			if debug {
				fmt.Fprintf(os.Stdout, "Adding new contractor: %d\t%s %s\n", co.ID, co.FirstName, co.LastName)
			}
			_, err = contractorinsstmt.Exec(co.ID, co.FirstName, co.LastName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Insert failed: %v\n", err)
				return false, err
			}
		}
	}

	tsheetstmt, err := db.Prepare("insert timesheets set id=?,day=?,start=?,end=?,client=?,contractor=?,duration=?")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Prepare SQL failed: %v\n", err)
		return false, err
	}

	for _, jo := range req.Jobs {
		//debug
		if debug {
			fmt.Fprintf(os.Stdout, "Inserting:\nDate: %s\tID: %d\tStart: %s\tEnd: %s\tClient: %d\tDur: %f\n", jo.Date, jo.ID, jo.Start, jo.End, jo.ClientID, jo.Duration)
		}
		_, err = tsheetstmt.Exec(jo.ID, jo.Date, jo.Start, jo.End, jo.ClientID, jo.UserID, jo.Duration)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Insert failed: %v\n", err)
			return false, err
		}
	}

	return true, nil
}
