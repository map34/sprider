package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

func main() {
	client := http.Client{
		Timeout: time.Second * 1,
	}

	sourceURL := "https://go.4cinsights.com/_/166/pinterest/admanager/campaigns"
	resp, err := client.Get(sourceURL)

	if err != nil {
		fmt.Println(err)
		return
	}

	defer resp.Body.Close()
	tokenizer := html.NewTokenizer(resp.Body)
	var wg sync.WaitGroup

	for {
		tokenType := tokenizer.Next()
		switch {

		case tokenType == html.ErrorToken:
			// End of the document, we're done
			fmt.Println("End of document")
			wg.Wait()
			return
		case tokenType == html.StartTagToken || tokenType == html.SelfClosingTagToken:
			token := tokenizer.Token()
			if token.Data == "img" {
				downloadImageToken(&token, &wg, sourceURL)
			}
		}
	}
}

func downloadImageToken(token *html.Token, waitGroup *sync.WaitGroup, sourceURL string) {
	if token.Data == "img" {
		for _, element := range token.Attr {
			parser, err := url.ParseRequestURI(element.Val)
			if err != nil {
				continue
			}
			waitGroup.Add(1)
			go func(urlParser *url.URL) {
				defer waitGroup.Done()
				if urlParser.Host != "" {
					err := DownloadFile(strings.Trim(urlParser.Path, "/"), urlParser.String())
					if err != nil {
						fmt.Println(err)
					}
				} else if urlParser.Host == "" && urlParser.Path != "" {
					sourceURLParser, err := url.ParseRequestURI(sourceURL)
					if err != nil {
						fmt.Println(err)
					}
					destURL := sourceURLParser.Scheme + "://" + sourceURLParser.Host + urlParser.Path
					err = DownloadFile(strings.Trim(urlParser.Path, "/"), destURL)
					if err != nil {
						fmt.Println(err)
					}
				}
			}(parser)
		}
	}
}

func DownloadFile(filepath string, url string) error {
	client := http.Client{
		Timeout: time.Second * 60,
	}

	// Get the data
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	realPath := "data/" + filepath

	err = os.MkdirAll(path.Dir(realPath), 0777)
	if err != nil {
		return err
	}

	out, err := os.Create(realPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Create the file

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(url)
	// fmt.Println("haha")

	return nil
}
