package main

import ("flag";"fmt";"io";"os";"path")
import ("net/http";"net/url")
import "golang.org/x/net/html"

func parseArgs() {
    flag.Usage = func() {
        fmt.Printf("usage: %s URL\n", path.Base(os.Args[0]))
        flag.PrintDefaults()
    }

    flag.Parse()
    if flag.NArg() == 0 {
        flag.Usage()
        os.Exit(1)
    }
}

func extractAnchorHref(getElement <-chan *html.Node, sendUrl chan<- *url.URL) {
    for node := range getElement {
        if node.Type == html.ElementNode && node.Data == "a" {
            for _, attr := range node.Attr {
                if attr.Key == "href" && attr.Val != "" {
                    url, err := url.Parse(attr.Val)
                    if err == nil {
                        sendUrl <- url
                    }
                }
            }
        }
    }
    close(sendUrl)
}

func visitAllChildren(doc *html.Node, sendElement chan<- *html.Node) {
    var f func(*html.Node)
    f = func(n *html.Node) {
        for c := n.FirstChild; c != nil; c = c.NextSibling {
            sendElement <- c
            f(c)
        }
    }
    f(doc)
    close(sendElement)
}

func getLinks(stream io.Reader) {
    doc, err := html.Parse(stream)
    if err != nil {
        fmt.Println(err)
        return
    }
    elements := make(chan *html.Node)
    urls := make(chan *url.URL)
    go visitAllChildren(doc, elements)
    go extractAnchorHref(elements, urls)
    for url := range urls {
        fmt.Println(url)
    }
}

func main() {
    parseArgs()
    url, err := url.Parse(flag.Arg(0))
    if err != nil {
        fmt.Println("Trouble parsing the site to spider!", err)
        os.Exit(1)
    }
    res, err := http.Get(url.String())
    if err != nil {
        fmt.Println("Trouble fetching the URL!", err)
        os.Exit(1)
    }
    getLinks(res.Body)
    res.Body.Close()

}
