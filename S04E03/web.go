package main

import (
	"ai_devs_3_tasks/helpers"
	"fmt"
	"net/http"
	"regexp"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
)

func GetLinks(url string, alloweDomain string) []string {
	collector := colly.NewCollector(
		colly.AllowedDomains(alloweDomain),
		colly.DisallowedURLFilters(

			regexp.MustCompile(".*czescizamienne.*"),
			regexp.MustCompile(".*loop.*"),
			regexp.MustCompile(".*cennik.*"),
		),
	)
	var links []string
	seen := make(map[string]bool)
	collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		links = helpers.AppendUnique(links, e.Request.AbsoluteURL(link), seen)
		collector.Visit(e.Request.AbsoluteURL(link))
	})
	collector.Visit(url)
	links = helpers.RemoveFromSliceIfContains(links, "czescizamienne")
	links = helpers.RemoveFromSliceIfContains(links, "loop")
	links = helpers.RemoveFromSliceIfContains(links, "cennik")
	links = helpers.RemoveFromSliceIfContains(links, "banan")
	return links
}

func GetMain(url string) (response []string, err error) {
	res, err := http.Get(url)
	if err != nil {
		return
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return
	}
	doc.Find("main").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the title
		text := s.Find("div").Text()
		doc.Find("a").Each(func(index int, item *goquery.Selection) {
			href, exists := item.Attr("href")
			if exists {
				title, _ := item.Attr("title")
				text = text + fmt.Sprintf("Link: %s, Title: %s", href, title)
			}
		})
		response = append(response, text)
	})
	tmp := doc.Find("article").Text()
	response = append(response, tmp)
	return
}
