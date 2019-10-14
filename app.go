package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"text/template"
)

var wg sync.WaitGroup

/*
SitemapIndex defines our main splice of sitemap locations
*/
type SitemapIndex struct {
	Locations []string `xml:"sitemap>loc"`
}

/*
News defines our main splice for each location from the sitemap
*/
type News struct {
	Titles    []string `xml:"url>news>title"`
	Keywords  []string `xml:"url>news>keywords"`
	Locations []string `xml:"url>loc"`
}

/*
NewsMap defines our main splice for our articles
*/
type NewsMap struct {
	Keyword  string
	Location string
}

/*
NewsAggPage ...
*/
type NewsAggPage struct {
	Title string
	News  map[string]NewsMap
}

func newsRoutine(c chan News, Location string) {
	defer wg.Done()
	var n News
	Location = strings.TrimSpace(Location)
	resp, _ := http.Get(Location)
	bytes, _ := ioutil.ReadAll(resp.Body)
	xml.Unmarshal(bytes, &n)
	resp.Body.Close()
	c <- n
}

func newsAggHandler(w http.ResponseWriter, r *http.Request) {
	var s SitemapIndex
	newsMap := make(map[string]NewsMap)

	resp, _ := http.Get("https://www.washingtonpost.com/news-sitemaps/index.xml")
	bytes, _ := ioutil.ReadAll(resp.Body)
	xml.Unmarshal(bytes, &s)
	resp.Body.Close()
	queue := make(chan News, 30)
	for _, Location := range s.Locations {
		wg.Add(1)
		go newsRoutine(queue, Location)
	}

	wg.Wait()
	close(queue)

	for elem := range queue {
		for idx := range elem.Keywords {
			newsMap[elem.Titles[idx]] = NewsMap{elem.Keywords[idx], elem.Locations[idx]}
		}
	}

	p := NewsAggPage{Title: "Amazing News Aggregator", News: newsMap}
	t, _ := template.ParseFiles("newsaggtemplate.html")
	fmt.Println(t.Execute(w, p))
}

func main() {
	http.HandleFunc("/agg", newsAggHandler)
	http.ListenAndServe(":1111", nil)

}
