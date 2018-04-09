package main

import (
  "fmt"
//  "github.com/PuerkitoBio/goquery"
  "log"
  "net/http"
//  "strings"
  "time"
  "database/sql"
)


const (
    maxDepth    = 3  // how many links to follow from a root
    snooze      = 250 * time.Millisecond
)

// print the table
func displayTable(db *sql.DB) {

  var url string
  var visited int
  var worker int

  rows, err := db.Query("SELECT * FROM urls")

  fmt.Println("-----------")
  fmt.Println("Table: urls")
  fmt.Println("-----------")
  for rows.Next() {
    if err = rows.Scan(&url, &visited, &worker); err == nil {
      fmt.Println(url, "\t", visited)
    }
  }
  fmt.Println("-----------")
}

// Send GET request for the given URL, return linked URLs and
// keywords. Also updates map "done".
func fetch(db *sql.DB, worker int) (links []string, keywords []string, err error) {

  // get a url from the database that hasn't been visited yet
  // do this indefinitely
  for {

    var url string
    var visited int
    var worker int

    // claim an unvisited url
    _, err := db.Exec("UPDATE visited FROM urls WHERE visited = ? AND worker = ? LIMIT 1", 0, 0)

    if err == nil {

      // if success in claim, get the entry
      row := db.QueryRow("SELECT * FROM urls WHERE visited = ? AND worker = ?", 0, worker)
      row.Scan(&url, visited, worker)
      fmt.Println(url)

      // read it
      // doc, err := goquery.NewDocument(url)
      resp, err := http.Get(url)
      if err != nil {
        log.Println("error", err, "fetching", url)
        return nil, nil, err
      }
      defer resp.Body.Close()

      // TODO: get the links inside of it

    } else {
      // wait for a little bit
      // fmt.Println("sleeping")
      time.Sleep(snooze)
    }
  }

  return
}


