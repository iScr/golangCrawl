package extenders

import (
	"github.com/PuerkitoBio/gocrawl"
)

var (
	pageStart = 1
	pageEnd   = 201
	baseUrl   = "http://jandan.net"
	partUrl   = "ooxx"
)

type jandanooxxExtender struct {
	gocrawl.DefaultExtender
}

func (this *jandanooxxExtender) Visit(ctx *gocrawl.URLContext, res *http.Response, doc *goquery.Document) (interface{}, bool) {
	fmt.Println("visit url: ", ctx.URL(), "state: ", ctx.State)
	go parsingImgUrl(res)
	// urls := processLinks(doc)
	// links := make(map[*url.URL]interface{})
	i, _ := ctx.State.(int)
	nextDepth := i - 1
	if nextDepth <= 0 {
		return nil, false
	}

	url := f.Sprintf("%s/%s/page-%d", baseUrl, partUrl, nextDepth)
	// for _, u := range urls {
	// 	links[u] = nextDepth
	// }
	return url, false
}

func (this *jandanooxxExtender) Filter(ctx *gocrawl.URLContext, isVisited bool) bool {
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
