package main

import (
	"fmt"
	"github.com/gocolly/colly"
	"github.com/xuri/excelize/v2"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type securePokemonData struct {
	PokemonData []string
	Mutex       sync.Mutex
}

func main() {
	file := excelize.NewFile()
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	var lastPages int
	collector := colly.NewCollector()
	runtime.GOMAXPROCS(2)
	now := time.Now()
	fmt.Println("start time : ", now)
	collector.OnRequest(func(r *colly.Request) {
		fmt.Println("On Request URL ", r.URL, ", time : ", time.Now())
	})

	collector.OnResponse(func(response *colly.Response) {
		fmt.Println("Status Code : ", response.StatusCode)
		for key, val := range *response.Headers {
			fmt.Println("key ", key, "val ", val)
		}
	})

	collector.OnScraped(func(response *colly.Response) {
		fmt.Println("Url ", response.Request.URL, " scrapped", ", time : ", time.Now())
	})

	collector.OnHTML("nav.woocommerce-pagination", func(element *colly.HTMLElement) {
		if element.Index == 0 {
			var pages = make([]string, 0)
			element.ForEach("a.page-numbers", func(i int, element *colly.HTMLElement) {
				//fmt.Println("index : ", i, "text : ", element.Text)
				pages = append(pages, element.Text)
			})
			//fmt.Println("pages count : ", len(pages))
			//fmt.Println("Last pages : ", pages[len(pages)-2])
			lastPages, _ = strconv.Atoi(pages[len(pages)-2])
		}
	})

	collector.Visit("https://scrapeme.live/shop/")
	fmt.Println("after scrapped last data page : ", lastPages)
	securePokemonData := securePokemonData{
		PokemonData: make([]string, 0),
		Mutex:       sync.Mutex{},
	}

	pagesToVisit := []string{}
	for i := 0; i < lastPages; i++ {
		pagesToVisit = append(pagesToVisit, fmt.Sprint("https://scrapeme.live/shop/page/", i+1))
	}

	c := colly.NewCollector(colly.Async(true))
	c.Limit(&colly.LimitRule{Parallelism: 2})

	c.OnRequest(func(request *colly.Request) {
		fmt.Println("Requested async : ", request.URL)
	})
	c.OnHTML("a.woocommerce-LoopProduct-link", func(element *colly.HTMLElement) {
		fmt.Println("a.woocommerce-LoopProduct-link found", "-", element.Request.URL, "-time : ", time.Now())
		//securePokemonData.Mutex.Lock()
		//securePokemonData.PokemonData = append(securePokemonData.PokemonData, fmt.Sprint(element.Attr("href"), " - ", element.Request.URL))
		//https://scrapeme.live/shop/page/27/
		file.SetActiveSheet(element.Index + 1)
		//file.SetSheetName()
		sheetName := element.Request.URL.String()
		var sheetIndex int
		if len(sheetName) > 30 {
			fmt.Println("get url page : ", sheetName[32:len(sheetName)-1])
			sheetIndex, _ = strconv.Atoi(sheetName[32 : len(sheetName)-1])
		} else {
			sheetIndex = 1
		}

		sheetName = fmt.Sprint(sheetName[17:21], "-", sheetIndex)

		if _, err := file.NewSheet(sheetName); err != nil {
			fmt.Println(err)
		}

		rows, _ := file.GetRows(sheetName)
		if err := file.SetCellValue(sheetName, fmt.Sprint("A", len(rows)+1), fmt.Sprint(element.Attr("href"), " - ", element.Request.URL, " - ", sheetName)); err != nil {
			fmt.Println(err)
		}
		//securePokemonData.Mutex.Unlock()

	})

	for _, s := range pagesToVisit {
		err := c.Visit(s)
		if err != nil {
			fmt.Println("error visit async : ", err.Error())
		}
	}
	c.Wait()

	fmt.Println("before loop : ", len(securePokemonData.PokemonData))
	for _, datum := range securePokemonData.PokemonData {
		fmt.Println(datum)
	}

	fmt.Println("done : ", time.Now().Sub(now).Seconds())
	if err := file.SaveAs("ScrapResult.xlsx"); err != nil {
		fmt.Println(err)
	}
}
