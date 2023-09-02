package main

import (
	"fmt"
	"github.com/gocolly/colly"
)

func main() {
	collector := colly.NewCollector()
	collector.OnRequest(func(r *colly.Request) {
		fmt.Println("On Request URL ", r.URL)
	})

	collector.OnResponse(func(response *colly.Response) {
		println("Status Code : ", response.StatusCode)
		for key, val := range *response.Headers {
			fmt.Println("key ", key, "val ", val)
		}
	})

	collector.OnScraped(func(response *colly.Response) {
		println("Url ", response.Request.URL, " scrapped")
	})

	collector.OnHTML("a.woocommerce-LoopProduct-link", func(element *colly.HTMLElement) {
		println("Tag Name : ", element.Name)
		println("Index : ", element.Index)
		println("Text : ", element.Text)
		println("Href : ", element.Attr("href"))
		if val, exists := element.DOM.ChildrenFiltered("img").Attr("src"); exists {
			println("Attr : ", val)
		}
	})

	collector.Visit("https://scrapeme.live/shop/")
}
