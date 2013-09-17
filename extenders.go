package main

import (
	"github.com/PuerkitoBio/gocrawl"
)

type jandanooxxExtender struct {
	gocrawl.DefaultExtender
}

func (this *ExampleExtender) Visit(ctx *gocrawl.URLContext, res *http.Response, doc *goquery.Document) (interface{}, bool) {
	fmt.Println("visit url: ", ctx.URL(), "state: ", ctx.State)
	go parsingImgUrl(res)
	urls := processLinks(doc)
	links := make(map[*url.URL]interface{})
	i, _ := ctx.State.(int)
	nextDepth := i - 1
	if nextDepth <= 0 {
		return nil, false
	}
	for _, u := range urls {
		links[u] = nextDepth
	}
	return links, false
}

func (this *ExampleExtender) Filter(ctx *gocrawl.URLContext, isVisited bool) bool {
	if ctx.SourceURL() == nil {
		ctx.State = DEPTH
		return !isVisited
	}
	if ctx.State != nil {
		i, ok := ctx.State.(int)
		if ok && i > 0 {
			return !isVisited
		}
	} else {
		fmt.Println("ctx.state nil, ctx.sourceURL: ", ctx.SourceURL())
	}
	return false
}
