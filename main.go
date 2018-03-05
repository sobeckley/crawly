/**
 * main program for the web crawler
 */

package main

import (
  "fmt"
)

func main() {
  crawler := Crawly{}
  crawler.CrawlyFunc()
  crawler.Get()
  fmt.Println("Hello World!")
}

