package main

import (
	"fmt"
	"github.com/PuerkitoBio/gocrawl"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	MAXINDEX = 500
	DEPTH    = 5
)

var (
	downloaded map[string]bool = make(map[string]bool)
	lockx                      = make(chan int, 1)
	imgDir     string

	par = make(chan int, 20)

	baseUrl = "http://jandan.net"
	partUrl = "ooxx"

	visitQuit = make(chan int, 1)
)

func addDownloadImgUrl(url string) {
	lockx <- 1
	downloaded[url] = true
	<-lockx
}

func getName(url string) string {
	namelist := strings.Split(url, "/")
	fileName := namelist[len(namelist)-1]
	return fileName
}

func downImg(url string, chann chan int) {

	defer func() {
		chann <- 1
	}()
	if downloaded[url] {
		fmt.Println("====url: ", url, "已下载")
		return
	}

	addDownloadImgUrl(url)

	resp, err := http.Get(url)
	delay := time.AfterFunc(10*time.Second, func() {
		fmt.Println("超时 取消下载: ", url)
		return
	})

	defer delay.Stop()

	if err != nil {
		fmt.Println("下载图片: ", url, "失败, 原因: ", err.Error())
		return
	}

	defer resp.Body.Close()

	if resp.ContentLength < 10000 {
		fmt.Println("图片过小. 取消下载")
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("读取图片: ", url, "失败, 原因: ", err.Error())
		return
	}

	path := "./img/" + getName(url)

	f, err := os.Create(path)
	if err != nil {
		fmt.Println("创建文件失败, path: ", path, "原因: ", err.Error())
		return
	}
	defer f.Close()

	f.Write(body)
	fmt.Println("----", resp.Request.URL)

}

func parsingImgUrl(resp *http.Response, quit chan int) {
	fmt.Println("解析图片链接, 来自: ", resp.Request.URL)

	if quit != nil {
		defer func() {
			<-quit
		}()
	}
	if resp == nil {
		fmt.Println("resp 为空")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("读取网页失败")
	}

	reg_n := regexp.MustCompile(`\s`)
	bodystr := reg_n.ReplaceAllString(string(body), " ")

	//查找匹配 <ol class="commentlist"><"/ol">
	re1 := regexp.MustCompile(`<ol class="commentlist".*</ol>`)
	li_comment_m := re1.Find([]byte(bodystr))

	//查找匹配 <p><img src="xxx" /></p>
	re_i := regexp.MustCompile(`<p><img src="(.+?)"`)
	img_urls := re_i.FindAllSubmatch(li_comment_m, -1)

	subChan2 := make(chan int, len(img_urls))
	fmt.Println("算法2图片数量: ", len(img_urls))
	for _, imageUrlTmp := range img_urls {
		go downImg(string(imageUrlTmp[1]), subChan2)
	}
	for i := 0; i < len(img_urls); i++ {
		<-subChan2
	}
}

type ExampleExtender struct {
	*gocrawl.DefaultExtender
}

func (this *ExampleExtender) Visit(ctx *gocrawl.URLContext, res *http.Response, doc *goquery.Document) (interface{}, bool) {
	fmt.Println("visit url: ", ctx.URL(), "state: ", ctx.State)

	go parsingImgUrl(res, nil)

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

type jandanooxxExtender struct {
	*gocrawl.DefaultExtender
}

func (this *jandanooxxExtender) Visit(ctx *gocrawl.URLContext, res *http.Response, doc *goquery.Document) (interface{}, bool) {
	fmt.Println("visit url: ", ctx.URL(), "state: ", ctx.State)
	par <- 1
	go parsingImgUrl(res, par)
	i, _ := ctx.State.(int)
	nextDepth := i + 1
	if nextDepth > MAXINDEX {
		return nil, false
	}

	url := fmt.Sprintf("%s/%s/page-%d", baseUrl, partUrl, nextDepth)
	// for _, u := range urls {
	// 	links[u] = nextDepth
	// }
	links := make(map[string]interface{})
	links[url] = nextDepth
	return links, false
}

func (this *jandanooxxExtender) Filter(ctx *gocrawl.URLContext, isVisited bool) bool {
	// if ctx.SourceURL() == nil {
	// 	ctx.State = DEPTH
	// 	return !isVisited
	// }
	// if ctx.State != nil {
	// 	i, ok := ctx.State.(int)
	// 	if ok && i > 0 {
	// 		return !isVisited
	// 	}
	// } else {
	// 	fmt.Println("ctx.state nil, ctx.sourceURL: ", ctx.SourceURL())
	// }
	return !isVisited
}

//copy from worker.go
func processLinks(doc *goquery.Document) (result []*url.URL) {
	urls := doc.Find("a[href]").Map(func(_ int, s *goquery.Selection) string {
		val, _ := s.Attr("href")
		return val
	})
	for _, s := range urls {
		if len(s) > 0 && !strings.HasPrefix(s, "#") {
			if parsed, e := url.Parse(s); e == nil {
				parsed = doc.Url.ResolveReference(parsed)
				result = append(result, parsed)
			}
		}
	}
	return
}

func main() {
	cDIr, _ := filepath.Abs("")
	imgDir = cDIr + "/img"
	os.MkdirAll("img", 0755)
	fmt.Println(os.Args[0])
	fmt.Println(len(os.Args))
	// return
	opts := gocrawl.NewOptions(new(jandanooxxExtender))

	opts.CrawlDelay = 0
	opts.LogFlags = gocrawl.LogNone
	opts.EnqueueChanBuffer = 10000
	// opts.MaxVisits = 4
	c := gocrawl.NewCrawlerWithOptions(opts)

	for {
		c.Run(gocrawl.S{"http://jandan.net/ooxx/page-1": 1})
		delay := time.After(5 * time.Minute)
		<-delay
	}
}

func loop() {

}
