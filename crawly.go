// Crawly
// ------
// The unit of concurrency for the web crawler.
// Has a:
// - TCP socket
// - map of (url, (resource, diff)) = diffy

package main

import (
  "fmt"
)

type Crawly struct {
  f bool
  // tcp socket
  // map
}

func (c *Crawly) CrawlyFunc() {
  fmt.Println("Crawly func!")
}

