// Crawly
// ------
// The unit of concurrency for the web crawler.
// Has a:
// - TCP socket
// - map of (url, (resource, diff)) = diffy

package main

import (
  "fmt"
  "net/http"
)

type Crawly struct {
  f bool
  //urls *diffy
  // map
}

func (c *Crawly) Get() {
  rep, err := http.Get("http://www.google.com")
  defer rep.Body.Close()
  fmt.Println(err)
  fmt.Println(rep)
}

func (c *Crawly) CrawlyFunc() {
  fmt.Println("Crawly func!")
}

