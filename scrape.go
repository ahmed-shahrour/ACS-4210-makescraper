package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/gocolly/colly"
)

type Data struct {
	Cursor string `json:"cursor"`
	Jobs   []Jobs `json:"jobs"`
}
type Jobs struct {
	Title     string `json:"title"`
	Link      string `json:"link"`
	Timestamp string `json:"timestamp"`
}

func normalizeLink(link string) string {
	if link[0] != 'h' {
		link = "https://news.ycombinator.com/" + link
	}

	return link
}

func exportData(jobs []Jobs, cursor string, db Data) {
	fmt.Println("Exporting data...")
	db.Cursor = cursor
	db.Jobs = append(db.Jobs, jobs...)

	jsonData, _ := json.MarshalIndent(db, "", " ")
	ioutil.WriteFile("jobs.json", jsonData, 0644)

	color.Set(color.FgHiGreen)
	fmt.Print("Export complete. ")
	color.Unset()

	info, _ := os.Stat("jobs.json")
	kilobytes := float64(info.Size()) / 1024
	fmt.Println(fmt.Sprintf("(%.2f", kilobytes), "KB)")
}

func main() {
	byteValue, _ := ioutil.ReadFile("jobs.json")
	var db Data
	json.Unmarshal(byteValue, &db)

	cursor := ""
	jobs := []Jobs{}
	timestampIndex := 0
	c := colly.NewCollector()

	// Job Details
	c.OnHTML("tr.athing > td.title > a.titlelink", func(e *colly.HTMLElement) {
		link := e.Attr("href")

		jobs = append(jobs, Jobs{
			Title: e.Text,
			Link:  normalizeLink(link),
		})
	})

	// Job Timestamp
	c.OnHTML("tr > td.subtext > span.age", func(e *colly.HTMLElement) {
		timestamp := e.Attr("title")

		jobs[timestampIndex].Timestamp = timestamp
		timestampIndex++
	})

	// Cursor
	c.OnHTML("tr > td.title > a.morelink", func(e *colly.HTMLElement) {
		cursor = normalizeLink(e.Attr("href"))

		exportData(jobs, cursor, db)

		time.Sleep(5 * time.Second)
		e.Request.Visit(cursor)
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting:", r.URL.String())
	})

	c.OnError(func(_ *colly.Response, err error) {
		fmt.Println("Something went wrong:", err)
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished", r.Request.URL)
	})

	c.Visit(db.Cursor)
}
