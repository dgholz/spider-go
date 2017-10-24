package main

import ("flag";"fmt";"os";"path")
import ("net/http";"net/url")

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
    res.Body.Close()

}
