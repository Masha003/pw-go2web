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
	"os"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

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
		response, err := searchTerm(*searchFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(response)
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

func makeRequest(url string) (string, error) {
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

	// request := fmt.Sprintf("GET %s HTTP/1.1\r\nHost: %s\r\nConnection: close\r\nUser-Agent: go2web/1.0\r\nAccept: text/html,application/json\r\n\r\n", path, host)
	request := fmt.Sprintf("GET %s HTTP/1.1\r\nHost: %s\r\nConnection: close\r\nUser-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36\r\nAccept: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8\r\nAccept-Language: en-US,en;q=0.5\r\n\r\n", path, host)

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

	return processResponse(response.String()), nil
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

func searchTerm(term string) (string, error) {
    // Try a simpler search engine that might be easier to scrape
    searchURL := "https://lite.duckduckgo.com/lite?q=" + strings.ReplaceAll(term, " ", "+")
    
    fmt.Fprintf(os.Stderr, "Searching for: %s\n", term)
    fmt.Fprintf(os.Stderr, "Search URL: %s\n", searchURL)
    
    response, err := makeRequest(searchURL)
    if err != nil {
        return "", fmt.Errorf("search failed: %v", err)
    }
    
    // Print first 200 characters of response for debugging
    fmt.Fprintf(os.Stderr, "Response preview (first 200 chars):\n%s\n", response[:min(200, len(response))])
    
    // Extract the top 10 results using HTML parser
    doc, err := html.Parse(strings.NewReader(response))
    if err != nil {
        return "", fmt.Errorf("failed to parse search results: %v", err)
    }
    
    // Find all search results
    var results []string
    count := 0
    
    // Function to recursively find search result elements for lite.duckduckgo.com
    var findResults func(*html.Node)
    findResults = func(n *html.Node) {
        if count >= 10 {
            return
        }
        
        // For lite.duckduckgo.com, look for <a> elements inside <td> elements
        if n.Type == html.ElementNode && n.Data == "a" {
            // Filter for links that are likely search results
            isResultLink := false
            var href string
            
            for _, attr := range n.Attr {
                if attr.Key == "href" && strings.HasPrefix(attr.Val, "http") {
                    isResultLink = true
                    href = attr.Val
                    break
                }
            }
            
            if isResultLink && n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
                title := strings.TrimSpace(n.FirstChild.Data)
                if title != "" && len(title) > 3 && !strings.HasPrefix(title, "More") {
                    // Format the result
                    result := fmt.Sprintf("%d. %s (%s)", count+1, title, href)
                    results = append(results, result)
                    count++
                }
            }
        }
        
        // Continue traversing the DOM
        for c := n.FirstChild; c != nil; c = c.NextSibling {
            findResults(c)
        }
    }
    
    findResults(doc)
    
    // If we didn't find any results with the above method, try with the original algorithm
    if len(results) == 0 {
        fmt.Fprintf(os.Stderr, "No results found with lite version algorithm, trying original...\n")
        
        // Try with the original DuckDuckGo HTML structure
        var findOriginalResults func(*html.Node)
        findOriginalResults = func(n *html.Node) {
            if count >= 10 {
                return
            }
            
            // Look for result containers
            if n.Type == html.ElementNode && n.Data == "div" {
                // Check if this is a result div
                isResult := false
                for _, attr := range n.Attr {
                    if attr.Key == "class" && (strings.Contains(attr.Val, "result__body") || 
                                              strings.Contains(attr.Val, "result") || 
                                              strings.Contains(attr.Val, "web-result")) {
                        isResult = true
                        break
                    }
                }
                
                if isResult {
                    var title string
                    
                    // Extract title text from this div and its children
                    var extractTitle func(*html.Node)
                    extractTitle = func(node *html.Node) {
                        if node.Type == html.TextNode && len(strings.TrimSpace(node.Data)) > 3 {
                            title = strings.TrimSpace(node.Data)
                            return
                        }
                        
                        for c := node.FirstChild; c != nil; c = c.NextSibling {
                            if title == "" {
                                extractTitle(c)
                            }
                        }
                    }
                    
                    extractTitle(n)
                    
                    if title != "" {
                        results = append(results, fmt.Sprintf("%d. %s", count+1, title))
                        count++
                    }
                }
            }
            
            // Continue traversing the DOM
            for c := n.FirstChild; c != nil; c = c.NextSibling {
                findOriginalResults(c)
            }
        }
        
        count = 0
        findOriginalResults(doc)
    }
    
    // Simplest fallback - just find any links
    if len(results) == 0 {
        fmt.Fprintf(os.Stderr, "Still no results, trying generic link extraction...\n")
        
        var findLinks func(*html.Node)
        findLinks = func(n *html.Node) {
            if count >= 10 {
                return
            }
            
            if n.Type == html.ElementNode && n.Data == "a" {
                var text string
                var href string
                
                // Get href attribute
                for _, attr := range n.Attr {
                    if attr.Key == "href" {
                        href = attr.Val
                        break
                    }
                }
                
                // Get link text
                if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
                    text = strings.TrimSpace(n.FirstChild.Data)
                }
                
                // Skip empty links, short links, or likely navigation
                if text != "" && len(text) > 5 && !strings.HasPrefix(text, "More") && 
                   !strings.HasPrefix(text, "Next") && href != "" {
                    results = append(results, fmt.Sprintf("%d. %s", count+1, text))
                    count++
                }
            }
            
            for c := n.FirstChild; c != nil; c = c.NextSibling {
                findLinks(c)
            }
        }
        
        count = 0
        findLinks(doc)
    }
    
    // If still no results, try to extract any meaningful text
    if len(results) == 0 {
        fmt.Fprintf(os.Stderr, "No links found, extracting any meaningful text...\n")
        
        var texts []string
        var extractMeaningfulText func(*html.Node)
        extractMeaningfulText = func(n *html.Node) {
            if len(texts) >= 10 {
                return
            }
            
            if n.Type == html.TextNode {
                text := strings.TrimSpace(n.Data)
                if len(text) > 20 && !strings.Contains(text, "DuckDuckGo") && 
                   !strings.Contains(text, "Privacy") && !strings.HasPrefix(text, "!") {
                    texts = append(texts, text)
                }
            }
            
            for c := n.FirstChild; c != nil; c = c.NextSibling {
                extractMeaningfulText(c)
            }
        }
        
        extractMeaningfulText(doc)
        
        for i, text := range texts {
            if i < 10 {
                results = append(results, fmt.Sprintf("%d. %s", i+1, text))
            }
        }
    }
    
    // If still no results, return a helpful message
    if len(results) == 0 {
        return "No search results found. The response from the search engine might be:\n1. Empty due to blocking of automated requests\n2. In a format this program can't parse\n\nTry with a more specific search term or using the -u option with a specific URL.", nil
    }
    
    // Join the results with newlines
    return strings.Join(results, "\n"), nil
}

// Helper function to get minimum of two integers
func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}