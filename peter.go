// A quick hack to generate a data set of URLs and keywords.
//
// - given some root URLs to start from
// - crawl the web up to maxLinks links from each root
// - extract up to maxKeywords keywords for each URL
// - produce output file according to Assignment 9
//
// Currently we extract keywords only from <p> tags; we split on
// whitespace and convert to lower case; we filter out a set of
// common stop words; then we pick the up to maxKeywords keywords
// that occurr most frequently.

package main

import (
    "flag"
    "fmt"
    "github.com/PuerkitoBio/goquery"
    "github.com/phf/go-queue/queue"
    "golang.org/x/text/transform"
    "golang.org/x/text/unicode/norm"
    "io/ioutil"
    "log"
    "net/http"
    "net/url"
    "os"
    "regexp"
    "strings"
    "time"
    "unicode"
)

const (
    maxDepth    = 3  // how many links to follow from a root
    maxKeywords = 15 // how many keywords to extract per URL
    snooze      = 250 * time.Millisecond
)

var todo *queue.Queue
var tolev *queue.Queue
var done map[string]bool // url -> status code (-1 for error)
var stop map[string]bool // string -> stop (ignore that word)
var transformer transform.Transformer

// Return the keys of a map as a slice.
func getKeys(m map[string]bool) []string {
    keys := make([]string, len(m))
    i := 0
    for k := range m {
        keys[i] = k
        i++
    }
    return keys
}

// Initialize our stop words from Sedgewick's data set.
func fetchStoppers() error {
    resp, err :=
http.Get("http://introcs.cs.princeton.edu/java/data/stopwords.txt")
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // transform away crazy unicode stuff if possible
    r := transform.NewReader(resp.Body, transformer)
    //    r := resp.Body

    all, err := ioutil.ReadAll(r)
    if err != nil {
        return err
    }

    stop = make(map[string]bool)
    words := strings.Fields(string(all))
    for _, w := range words {
        stop[strings.ToLower(w)] = true
    }

    // a few manual additions
    stop["https"] = true
    stop["http"] = true

    return nil
}

// Check if we have a valid/useful keyword.
func useful(keyword string) bool {
    if len(keyword) <= 2 {
        return false
    }
    if _, ok := stop[keyword]; ok {
        return false
    }
    return true
}

func mapTagA(index int, sel *goquery.Selection) string {
    url, ok := sel.Attr("href")
    if !ok {
        return "" // pretend we saw nothing, resolve later
    }
    url = strings.Trim(url, " ")
    return url
}

func mapTagMeta(index int, sel *goquery.Selection) string {
    name, ok := sel.Attr("name")
    if !ok {
        return "" // nothing to see
    }
    if name != "keywords" && name != "description" && name != "author" {
        return "" // nothing to see
    }
    info, ok := sel.Attr("content")
    if !ok {
        return "" // pretend we saw nothing, resolve later
    }
    return info
}

// Extract keywords from a "blob" of text given to us as one big string.
// Finding a decent strategy for this is moderately convoluted... :-/
func extract(text string) (words []string, err error) {
    // Step zero is to turn off all the Unicode insanity and focus
    // on good old ASCII. So let's do that.
    ascii, _, err := transform.String(transformer, text)
    if err != nil {
        log.Println("error", err, "during transform")
        return nil, err
    }

    // Then let's make sure it's all lower case.
    lower := strings.ToLower(ascii)

    // Then let's replace characters we certainly don't like with
    // spaces, leaving only (a) things we moderately like and (b)
    // whitespace.
    dontLike := regexp.MustCompile("\\s|_|\\W")
    like := dontLike.ReplaceAllString(lower, " ")

    // Finally let's split it all on whitespace into "words" as we
    // see them...
    words = strings.Fields(like)
    return
}

func max(a, b int) int {
    if a > b {
        return a
    } else {
        return b
    }
}

// Determine (up to) top "count" words from a wordcount map.
func top(counts map[string]int, count int) (words []string) {
    words = nil

    // determine the maximum count
    max := 0
    for _, v := range counts {
        if v > max {
            max = v
        }
    }

    // try to obtain "count" words
    got := 0
    cur := max
    for got < count && len(counts) > 0 {
        for k, v := range counts {
            if v == cur {
                words = append(words, k)
                got++
                delete(counts, k)
                if got >= count {
                    break
                }
            }
        }
        if cur > 0 {
            cur--
        }
    }
    return
}

// Send GET request for the given URL, return linked URLs and
// keywords. Also updates map "done".
func fetch(from string) (links []string, keywords []string, err error) {
    var uniqueUrls map[string]bool = make(map[string]bool)
    var wordCounts map[string]int = make(map[string]int)
    if _, ok := done[from]; ok {
        err := fmt.Errorf("duplicate fetch %s", from)
        return nil, nil, err
    }
    updateCounts := func(words []string) {
        for _, w := range words {
            if !useful(w) {
                continue // skip
            }
            wordCounts[w]++
        }
    }

    time.Sleep(snooze)

    // make sure it's HTML!!!
    resp, err := http.Head(from)
    if err != nil {
        log.Println("error", err, "fetching", from)
        return nil, nil, err
    }
    defer resp.Body.Close()
    content := resp.Header.Get("Content-Type")
    if !strings.HasPrefix(content, "text/html") {
        err := fmt.Errorf("wrong content type %s", content)
        return nil, nil, err
    }

    time.Sleep(snooze)

    doc, err := goquery.NewDocument(from)
    if err != nil {
        err := fmt.Errorf("failed to make document for %s", from)
        return nil, nil, err
    }

    atags := doc.Find("a").Map(mapTagA)
    for _, x := range atags {
        u, err := url.Parse(x)
        if err != nil {
            log.Println(err)
            continue // skip
        }
        r := doc.Url.ResolveReference(u)
        uniqueUrls[r.String()] = true
    }
    links = getKeys(uniqueUrls)

    metatags := doc.Find("meta").Map(mapTagMeta)
    for _, x := range metatags {
        w, err := extract(x)
        if err == nil {
            updateCounts(w)
        }
    }

    ptags := doc.Find("p")
    for i := range ptags.Nodes {
        single := ptags.Eq(i)
        text := single.Text()
        w, err := extract(text)
        if err == nil {
            updateCounts(w)
        }
    }

    litags := doc.Find("li")
    for i := range litags.Nodes {
        single := litags.Eq(i)
        text := single.Text()
        w, err := extract(text)
        if err == nil {
            updateCounts(w)
        }
    }

    h1tags := doc.Find("h1")
    for i := range h1tags.Nodes {
        single := h1tags.Eq(i)
        text := single.Text()
        w, err := extract(text)
        if err == nil {
            updateCounts(w)
        }
    }

    keywords = top(wordCounts, maxKeywords)

    done[from] = true
    return
}

func main() {
    log.SetOutput(os.Stderr) // don't let logging interfere with output

    todo = queue.New()
    tolev = queue.New()
    done = make(map[string]bool)
    isMn := func(r rune) bool {
        return unicode.Is(unicode.Mn, r)
    }
    transformer = transform.Chain(norm.NFD,
transform.RemoveFunc(isMn), norm.NFC)

    err := fetchStoppers()
    if err != nil {
        log.Fatal(err)
    }

    flag.Parse()
    args := flag.Args()

    for _, arg := range args {
        todo.PushBack(arg)
        tolev.PushBack(0)
    }

    for todo.Len() > 0 {
        url := todo.PopFront().(string)
        lev := tolev.PopFront().(int)

        if lev > maxDepth {
            log.Println("skipped", url, "at level", lev, "too deep")
            continue
        }

        if done[url] {
            log.Println("skipped", url, "at level", lev, "already done")
            continue
        }

        if strings.ContainsAny(url, "#?=%") {
            log.Println("skipped", url, "at level", lev, "just hate it")
            continue
        }

        log.Println("trying", url, "at level", lev)

        links, keywords, err := fetch(url)
        if err != nil {
            log.Println("error", err, "fetching", url)
            continue
        }

        for _, l := range links {
            todo.PushBack(l)
            tolev.PushBack(lev + 1)
        }

        fmt.Println(url)
        for i, v := range keywords {
            fmt.Print(v)
            if i < len(keywords)-1 {
                fmt.Print(" ")
            }
        }
        fmt.Println()
    }
}

