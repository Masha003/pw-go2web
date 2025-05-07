package search

import (
	"fmt"
	"log"
	"net/url"
	"pw-go2web/internal/client"
	"pw-go2web/internal/parser"
	"pw-go2web/models"
	"strings"

	"golang.org/x/net/html"
)

func SearchTerm(term string) ([]models.SearchResult, error) {
	// Prepare the search URL using DuckDuckGo lite
	encodedTerm := url.QueryEscape(term)
	searchURL := fmt.Sprintf("https://lite.duckduckgo.com/lite?q=%s", encodedTerm)
	
	// Make the request
	response, err := client.MakeRequest(searchURL, false)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %v", err)
	}
	
	// Parse the HTML and extract search results
	doc, err := html.Parse(strings.NewReader(response))
	if err != nil {
		return nil, fmt.Errorf("failed to parse search results: %v", err)
	}
	
	// Extract the results using techniques from the working version
	results := ParseSearchResults(doc)
	
	// Limit to 10 results
	if len(results) > 10 {
		results = results[:10]
	}
	
	if len(results) == 0 {
		return nil, fmt.Errorf("no search results found")
	}
	
	return results, nil
}


func ParseSearchResults(doc *html.Node) []models.SearchResult {
	var results []models.SearchResult
	
	var findResultLinks func(*html.Node)
	findResultLinks = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			isResultLink := false
			var href string
			
			// Check if it's a result link
			for _, attr := range n.Attr {
				if attr.Key == "class" && attr.Val == "result-link" {
					isResultLink = true
				}
				if attr.Key == "href" {
					href = attr.Val
				}
			}
			
			if isResultLink && href != "" {
				// Extract title
				title := parser.ExtractTextContent(n)
				
				// Find parent <tr> and extract description and URL from adjacent rows
				resultRow := parser.FindParentByTagName(n, "tr")
				if resultRow != nil {
					var description, originalURL string
					
					// Look for snippet in next row
					nextRow := parser.FindNextSibling(resultRow)
					if nextRow != nil {
						snippetCell := parser.FindElementByClass(nextRow, "result-snippet")
						if snippetCell != nil {
							description = parser.ExtractTextContent(snippetCell)
						}
					}
					
					// Look for URL in the row after that
					urlRow := parser.FindNextSibling(nextRow)
					if urlRow != nil {
						linkTextSpan := parser.FindElementByClass(urlRow, "link-text")
						if linkTextSpan != nil {
							originalURL = parser.ExtractTextContent(linkTextSpan)
						}
					}
					
					// Extract actual URL from DuckDuckGo redirect URL
					actualURL := ExtractActualURL(href)
					if actualURL == "" {
						actualURL = originalURL
					}
					
					if title != "" && actualURL != "" {
						results = append(results, models.SearchResult{
							Title:       title,
							URL:         actualURL,
							Description: description,
						})
					}
				}
			}
		}
		
		// Continue traversing the tree
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findResultLinks(c)
		}
	}
	
	findResultLinks(doc)
	
	if len(results) == 0 {
		log.Println("No Results found.")
	}
	
	return results
}

func ExtractActualURL(ddgURL string) string {
	if strings.Contains(ddgURL, "duckduckgo.com/l/?uddg=") {
		parts := strings.Split(ddgURL, "uddg=")
		if len(parts) > 1 {
			encodedURL := parts[1]
			if idx := strings.Index(encodedURL, "&"); idx > 0 {
				encodedURL = encodedURL[:idx]
			}
			
			if decodedURL, err := url.QueryUnescape(encodedURL); err == nil {
				return decodedURL
			}
		}
	}
	
	if !strings.HasPrefix(ddgURL, "//") {
		return ddgURL
	}
	
	return "https:" + ddgURL
}