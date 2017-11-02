package main

import ("flag";"fmt";"os";"path";"strings")
import ("net/http";"net/url")
import "golang.org/x/net/html"

func parseArgs() {
    flag.Usage = func() {
        fmt.Printf("usage: %s URL\n", path.Base(os.Args[0]))
        flag.PrintDefaults()
    }

    flag.Parse()
    if flag.NArg() != 1 {
        flag.Usage()
        os.Exit(1)
    }
}

func extractAnchorHref(getElement <-chan *html.Node) (<-chan *url.URL) {
    sendUrl := make(chan *url.URL)
    go func() {
        for node := range getElement {
            if node.Type != html.ElementNode || node.Data != "a" {
                continue
            }

            for _, attr := range node.Attr {
                if attr.Key != "href" {
                    continue
                }

                val := strings.TrimSpace(attr.Val)
                if val != "" {
                    url, err := url.Parse(val)
                    if err == nil {
                        sendUrl <- url
                    }
                    break
                }
            }
        }
        close(sendUrl)
    }()
    return sendUrl
}

func visitAllChildren(doc *html.Node) (<-chan *html.Node) {
    sendElement := make(chan *html.Node)
    go func() {
        var f func(*html.Node)
        f = func(n *html.Node) {
            for c := n.FirstChild; c != nil; c = c.NextSibling {
                sendElement <- c
                f(c)
            }
        }
        f(doc)
        close(sendElement)
    }()
    return sendElement
}

func getLinks(url *url.URL) (<-chan *url.URL) {
    res, err := http.Get(url.String())
    if err != nil {
        fmt.Println("Trouble fetching the URL!", err)
        return nil
    }

    doc, err := html.Parse(res.Body)
    res.Body.Close()
    if err != nil {
        fmt.Println(err)
        return nil
    }

    elements := visitAllChildren(doc)
    urls := extractAnchorHref(elements)
    sameSite := matchNetloc(url)(urls)
    return sameSite
}

type urlfilter func(<-chan *url.URL) (<-chan *url.URL)

func matchNetloc(baseUrl *url.URL) urlfilter {
    return func(getUrl <-chan *url.URL) (<-chan *url.URL) {
        filteredUrls := make(chan *url.URL)
        go func() {
            for url := range getUrl {
                abs := baseUrl.ResolveReference(url)
                if(abs.Host == baseUrl.Host) {
                    filteredUrls <- abs
                }
            }
            close(filteredUrls)
        }()
        return filteredUrls
    }
}

type VisitLink struct {
    from, to *url.URL
}
func (v VisitLink) get() (from *url.URL, to *url.URL) {
    return v.from, v.to
}

func spider(toVisit <-chan *url.URL) (<-chan VisitLink) {
    sendVisited := make(chan VisitLink)
    go func() {
        for nextUrl := range toVisit {
            go func() {
                for visited := range getLinks(url) {
                    visitedUrls <- VisitLink{url, visited}
                }
                close(visitedUrls)
            }()
        }
    }()
    return sendVisited
}

func main() {
    parseArgs()
    startUrl, err := url.Parse(flag.Arg(0))
    if err != nil {
        fmt.Println("Trouble parsing the site to spider!", err)
        os.Exit(1)
    }
    toVisit := make(chan *url.URL)
    visited := spider(toVisit)
    toVisit <- startUrl

    type visitRecord map[*url.URL]bool
    seen := map[*url.URL]visitRecord {
        startUrl: make(visitRecord),
    }

    for link := range visited {
        origin, dest := link.get()
        dest.Fragment = ""
        seen[origin][dest] = true
        _, alreadySeen := seen[dest]
        if ! alreadySeen {
             seen[dest] = make(visitRecord)
             toVisit <- dest
        }
    }

    for k, v := range seen {
        fmt.Println(k)
        for s := range v {
            if s != k {
                fmt.Println(" * ", s)
            }
        }
    }
}
