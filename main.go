package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

type SearchResult struct {
	Title       string
	URL         string
	Description string
}

func main() {
	urlFlag := flag.String("u", "", "The URL to connect to.")
	searchFlag := flag.String("s", "", "The search term to look for.")
	helpFlag := flag.Bool("h", false, "Display help information.")

	flag.Parse()

	if *helpFlag || (flag.NFlag() == 0) {
		fmt.Println("Usage of go2web:")
		fmt.Println("  go2web -u <URL>         # make an HTTP request to the specified URL and print the response")
		fmt.Println("  go2web -s <search-term> # make an HTTP request to search the term using a search engine and print top 10 results")
		fmt.Println("  go2web -h               # show this help")
		return
	}

	if *urlFlag != "" {
		response, err := makeRequest(*urlFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(response)
		return
	}

	if *searchFlag != "" {
		results, err := searchTerm(*searchFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	
		// Format the results like in the project
		for i, result := range results {
			fmt.Printf("%d. %s\n   URL: %s\n\n", i+1, result.Title, result.URL)
		}
		return
	}
}

func parseURL(rawURL string) (string, string, string, int, error) {
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "http://" + rawURL
	}

	var scheme string
	var port int
	if strings.HasPrefix(rawURL, "https://") {
		scheme = "https"
		port = 443
		rawURL = strings.TrimPrefix(rawURL, "https://")
	} else {
		scheme = "http"
		port = 80
		rawURL = strings.TrimPrefix(rawURL, "http://")
	}

	parts := strings.SplitN(rawURL, "/", 2)
	host := parts[0]
	
	if strings.Contains(host, ":") {
		hostParts := strings.SplitN(host, ":", 2)
		host = hostParts[0]
		fmt.Sscanf(hostParts[1], "%d", &port)
	}
	
	var path string
	if len(parts) > 1 {
		path = "/" + parts[1]
	} else {
		path = "/"
	}

	return scheme, host, path, port, nil
}

// func makeRequest(url string) (string, error) {
// 	scheme, host, path, port, err := parseURL(url)
// 	if err != nil {
// 		return "", err
// 	}

// 	var conn io.ReadWriteCloser

// 	if scheme == "https" {
// 		tlsConfig := &tls.Config{
// 			ServerName: host,
// 		}

// 		tlsConnection, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", host, port), tlsConfig)
// 		if err != nil {
// 			return "", fmt.Errorf("error connecting to %s: %v", host, err)
// 		}

// 		conn = tlsConnection
// 	} else {
// 		tlsConnection, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
// 		if err != nil {
// 			return "", fmt.Errorf("error connecting to %s: %v", host, err)
// 		}
// 		conn = tlsConnection
// 	}
// 	defer conn.Close()

// 	request := fmt.Sprintf("GET %s HTTP/1.1\r\n"+
// 	"Host: %s\r\n"+
// 	"Connection: close\r\n"+
// 	"User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36\r\n"+
// 	"Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8\r\n"+
// 	"Accept-Language: en-US,en;q=0.5\r\n\r\n", path, host)

// 	_, err = conn.Write([]byte(request))
// 	if err != nil {
// 		return "", fmt.Errorf("failed to send request: %v", err)
// 	}

// 	var response strings.Builder
// 	reader := bufio.NewReader(conn)

// 	for {
// 		line, err := reader.ReadString('\n')
// 		if err != nil {
// 			if err == io.EOF {
// 				// We've reached the end of the response
// 				break
// 			}
// 			return "", fmt.Errorf("error reading response: %v", err)
// 		}
// 		response.WriteString(line)
// 	}

// 	return processResponse(response.String()), nil
// }

func makeRequestRaw(url string, processHtml bool) (string, error) {
	scheme, host, path, port, err := parseURL(url)
	if err != nil {
		return "", err
	}

	var conn io.ReadWriteCloser

	if scheme == "https" {
		tlsConfig := &tls.Config{
			ServerName: host,
		}

		tlsConnection, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", host, port), tlsConfig)
		if err != nil {
			return "", fmt.Errorf("error connecting to %s: %v", host, err)
		}

		conn = tlsConnection
	} else {
		tlsConnection, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
		if err != nil {
			return "", fmt.Errorf("error connecting to %s: %v", host, err)
		}
		conn = tlsConnection
	}
	defer conn.Close()

	request := fmt.Sprintf("GET %s HTTP/1.1\r\n"+
	"Host: %s\r\n"+
	"Connection: close\r\n"+
	"User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36\r\n"+
	"Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8\r\n"+
	"Accept-Language: en-US,en;q=0.5\r\n\r\n", path, host)

	_, err = conn.Write([]byte(request))
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}

	var response strings.Builder
	reader := bufio.NewReader(conn)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// We've reached the end of the response
				break
			}
			return "", fmt.Errorf("error reading response: %v", err)
		}
		response.WriteString(line)
	}

	// If we don't want to process the HTML, extract the body and return it
	if !processHtml {
		rawResponse := response.String()
		parts := strings.SplitN(rawResponse, "\r\n\r\n", 2)
		if len(parts) < 2 {
			parts = strings.SplitN(rawResponse, "\n\n", 2)
			if len(parts) < 2 {
				return rawResponse, nil
			}
		}
		return parts[1], nil
	}

	// Otherwise process as usual
	return processResponse(response.String()), nil
}

// Original makeRequest function - keep this for normal requests
func makeRequest(url string) (string, error) {
	return makeRequestRaw(url, true)
}


func processResponse(rawResponse string) string {
	parts := strings.SplitN(rawResponse, "\r\n\r\n", 2)

	if len(parts) < 2 {
		parts = strings.SplitN(rawResponse, "\n\n", 2)
		if len(parts) < 2 {
			return rawResponse
		}
	}

	headers := parts[0]
	body := parts[1]

	contentTypeMatch := regexp.MustCompile(`(?i)Content-Type:\s*([^\r\n]+)`).FindStringSubmatch(headers)
	var contentType string
	if len(contentTypeMatch) > 1 {
		contentType = strings.ToLower(contentTypeMatch[1])
	}

	// Handle JSON content
	if strings.Contains(contentType, "application/json") {
		return formatJSON(body)
	}

	// Handle HTML content
	if strings.Contains(contentType, "text/html") {
		return extractTextFromHTML(body)
	}

	// Default case - just return the body with minimal processing
	return strings.TrimSpace(body)

}

// Format JSON nicely
func formatJSON(jsonStr string) string {
	var jsonObj interface{}
	
	// Try to parse the JSON
	err := json.Unmarshal([]byte(jsonStr), &jsonObj)
	if err != nil {
		// If parsing fails, return the original string
		return jsonStr
	}
	
	// Pretty print with indentation
	prettyJSON, err := json.MarshalIndent(jsonObj, "", "  ")
	if err != nil {
		return jsonStr
	}
	
	return string(prettyJSON)
}

// Extract text from HTML using a proper HTML parser
func extractTextFromHTML(htmlStr string) string {
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		// Fall back to regex-based stripping if parsing fails
		re := regexp.MustCompile("<[^>]*>")
		return strings.TrimSpace(re.ReplaceAllString(htmlStr, " "))
	}
	
	var buf bytes.Buffer
	extractText(doc, &buf)
	
	// Clean up whitespace
	result := buf.String()
	spaceRegex := regexp.MustCompile(`\s+`)
	result = spaceRegex.ReplaceAllString(result, " ")
	
	return strings.TrimSpace(result)
}

// Recursive function to extract text from HTML nodes
func extractText(n *html.Node, buf *bytes.Buffer) {
	if n.Type == html.TextNode {
		// Skip script and style content
		if n.Parent != nil && (n.Parent.Data == "script" || n.Parent.Data == "style") {
			return
		}
		buf.WriteString(n.Data)
		buf.WriteString(" ")
	}
	
	// Add extra spaces for block elements to preserve some formatting
	if n.Type == html.ElementNode {
		switch n.Data {
		case "div", "p", "br", "li", "h1", "h2", "h3", "h4", "h5", "h6", "tr":
			buf.WriteString("\n")
		}
	}
	
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractText(c, buf)
	}
	
	// Add newlines after certain block elements
	if n.Type == html.ElementNode {
		switch n.Data {
		case "div", "p", "tr", "li", "h1", "h2", "h3", "h4", "h5", "h6":
			buf.WriteString("\n")
		}
	}
}

func searchTerm(term string) ([]SearchResult, error) {
	// Prepare the search URL using DuckDuckGo lite
	encodedTerm := url.QueryEscape(term)
	searchURL := fmt.Sprintf("https://lite.duckduckgo.com/lite?q=%s", encodedTerm)
	
	// Make the request
	response, err := makeRequestRaw(searchURL, false)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %v", err)
	}
	
	// Parse the HTML and extract search results
	doc, err := html.Parse(strings.NewReader(response))
	if err != nil {
		return nil, fmt.Errorf("failed to parse search results: %v", err)
	}
	
	// Extract the results using techniques from the working version
	results := parseSearchResults(doc)
	
	// Limit to 10 results
	if len(results) > 10 {
		results = results[:10]
	}
	
	if len(results) == 0 {
		return nil, fmt.Errorf("no search results found")
	}
	
	return results, nil
}


func parseSearchResults(doc *html.Node) []SearchResult {
	var results []SearchResult
	
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
				title := extractTextContent(n)
				
				// Find parent <tr> and extract description and URL from adjacent rows
				resultRow := findParentByTagName(n, "tr")
				if resultRow != nil {
					var description, originalURL string
					
					// Look for snippet in next row
					nextRow := findNextSibling(resultRow)
					if nextRow != nil {
						snippetCell := findElementByClass(nextRow, "result-snippet")
						if snippetCell != nil {
							description = extractTextContent(snippetCell)
						}
					}
					
					// Look for URL in the row after that
					urlRow := findNextSibling(nextRow)
					if urlRow != nil {
						linkTextSpan := findElementByClass(urlRow, "link-text")
						if linkTextSpan != nil {
							originalURL = extractTextContent(linkTextSpan)
						}
					}
					
					// Extract actual URL from DuckDuckGo redirect URL
					actualURL := extractActualURL(href)
					if actualURL == "" {
						actualURL = originalURL
					}
					
					if title != "" && actualURL != "" {
						results = append(results, SearchResult{
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
		// Fallback method if no results found with the primary method
	if len(results) == 0 {
		var findLinks func(*html.Node)
		findLinks = func(n *html.Node) {
			if n.Type == html.ElementNode && n.Data == "a" {
				var href string
				
				// Extract href
				for _, attr := range n.Attr {
					if attr.Key == "href" && strings.HasPrefix(attr.Val, "http") {
						href = attr.Val
						break
					}
				}
				
				if href != "" && n.FirstChild != nil {
					title := extractTextContent(n)
					
					// Skip certain common non-result links
					if title != "" && len(title) > 5 && 
						!strings.HasPrefix(title, "More") && 
						!strings.HasPrefix(title, "Next") {
						results = append(results, SearchResult{
							Title: title,
							URL:   href,
						})
					}
				}
			}
			
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				findLinks(c)
			}
		}
		
		findLinks(doc)
	}
	
	return results
}

func extractTextContent(n *html.Node) string {
	if n == nil {
		return ""
	}
	
	var sb strings.Builder
	
	var extract func(*html.Node)
	extract = func(node *html.Node) {
		if node.Type == html.TextNode {
			sb.WriteString(node.Data)
		}
		
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}
	}
	
	extract(n)
	return strings.TrimSpace(sb.String())
}

// Helper function to find parent by tag name
func findParentByTagName(n *html.Node, tagName string) *html.Node {
	if n == nil {
		return nil
	}
	
	current := n.Parent
	for current != nil {
		if current.Type == html.ElementNode && current.Data == tagName {
			return current
		}
		current = current.Parent
	}
	
	return nil
}

// Helper function to find next sibling element
func findNextSibling(n *html.Node) *html.Node {
	if n == nil {
		return nil
	}
	
	current := n.NextSibling
	for current != nil {
		if current.Type == html.ElementNode {
			return current
		}
		current = current.NextSibling
	}
	
	return nil
}

// Helper function to find element by class
func findElementByClass(n *html.Node, className string) *html.Node {
	if n == nil {
		return nil
	}
	
	var find func(*html.Node) *html.Node
	find = func(node *html.Node) *html.Node {
		if node.Type == html.ElementNode {
			for _, attr := range node.Attr {
				if attr.Key == "class" && attr.Val == className {
					return node
				}
			}
		}
		
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			if result := find(c); result != nil {
				return result
			}
		}
		
		return nil
	}
	
	return find(n)
}

// Helper function to extract actual URL from DuckDuckGo redirect URL
func extractActualURL(ddgURL string) string {
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