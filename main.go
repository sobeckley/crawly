package main

import (
  "fmt"
  "database/sql"
  _ "github.com/mattn/go-sqlite3"
)

const (
    numWorkers  = 10 // how many goroutines to dispatch
)

func main() {

  // open the database
  db, err := sql.Open("sqlite3", "./todo.db")
  if err != nil {
    fmt.Println("uh oh")
  }
  defer db.Close()

  // lookup the urls table
  fmt.Println("Looking for table...")
  res, err := db.Exec("SELECT name FROM sqlite_master " +
    "WHERE type='table' AND name='{table_name}';")

  // if it exists, delete it
  if res != nil {
    fmt.Println("Table exists, deleting.")
    db.Exec("DROP TABLE IF EXISTS urls")
  }

  fmt.Println("Creating new table.")
  res, err = db.Exec("CREATE TABLE urls ( " +
    "url TEXT PRIMARY KEY NOT NULL, " +
    "visited INT NOT NULL, " +
    "worker INT NOT NULL" +
  ");")

  // delete everything from the table
  // res, err = db.Exec("DELETE FROM urls")

  // check if there's anything in it
  rows, _ := db.Query("SELECT * FROM urls")

  // if nothing in it, put something in it
  if !rows.Next() {

    fmt.Println("Table was empty. Inserting some urls...")

    urls := []string {
      "http://www.cs.jhu.edu/~phf/",
      "https://www.google.com",
      "http://lmgtfy.com/?q=hello",
    }

    // zero means no worker currently working on it
    for _, url := range urls {
      res, err = db.Exec("INSERT into urls (url, visited, worker) VALUES " +
        "('" + url + "', 0, 0);")
      if err != nil {
        fmt.Println(err)
      }
    }
  }

  // print the table and wait for enter to proceed
  displayTable(db)
  fmt.Scanln()

  // now dispatch go routines
  for i := 1; i <= numWorkers; i++ {
    go fetch(db, i)
  }

  // wait for enter to exit
  fmt.Scanln()

}

