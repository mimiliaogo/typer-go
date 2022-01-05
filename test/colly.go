package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/gocolly/colly"
)

func init() {
	// Define the other flags here
	// flag.IntVar()
}

func main() {

	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.OnResponse(func(r *colly.Response) {
		doc, err := htmlquery.Parse(strings.NewReader(string(r.Body)))
		if err != nil {
			log.Fatal(err)
		}
		// nodes := htmlquery.FindOne(doc, `//*[@id="mw-content-text"]/div[1]/p[2]`)
		nodes := htmlquery.FindOne(doc, `//*[@id="mw-content-text"]/div[1]/p[1]`)
		s := htmlquery.InnerText(nodes)
		fmt.Println(s)

		// remove non ascii
		re := regexp.MustCompile("[[:^ascii:]]")
		t := re.ReplaceAllLiteralString(s, "")
		fmt.Println(t)

		// remove consecutive spaces
		re = regexp.MustCompile("[ ]{2,}")
		t = re.ReplaceAllLiteralString(t, " ")
		fmt.Println(t)
		// for _, node := range nodes {
		// 	a := htmlquery.FindOne(node, "./a[@href]")
		// 	fmt.Println(htmlquery.SelectAttr(a, "href"), htmlquery.InnerText(a))
		// }
	})
	// c.Visit("https://en.wikipedia.org/wiki/Special:Random")
	c.Visit("https://en.wikipedia.org/wiki/Arca_Totok_Kerot")
}
