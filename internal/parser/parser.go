package parser

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)


func ProcessResponse(rawResponse string) string {
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
		return FormatJSON(body)
	}

	// Handle HTML content
	if strings.Contains(contentType, "text/html") {
		return ExtractTextFromHTML(body)
	}

	// Default case - just return the body with minimal processing
	return strings.TrimSpace(body)

}

// Format JSON nicely
func FormatJSON(jsonStr string) string {
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
func ExtractTextFromHTML(htmlStr string) string {
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		// Fall back to regex-based stripping if parsing fails
		re := regexp.MustCompile("<[^>]*>")
		return strings.TrimSpace(re.ReplaceAllString(htmlStr, " "))
	}
	
	var buf bytes.Buffer
	ExtractText(doc, &buf)
	
	// Clean up whitespace
	result := buf.String()
	spaceRegex := regexp.MustCompile(`\s+`)
	result = spaceRegex.ReplaceAllString(result, " ")
	
	return strings.TrimSpace(result)
}

// Recursive function to extract text from HTML nodes
func ExtractText(n *html.Node, buf *bytes.Buffer) {
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
		ExtractText(c, buf)
	}
	
	// Add newlines after certain block elements
	if n.Type == html.ElementNode {
		switch n.Data {
		case "div", "p", "tr", "li", "h1", "h2", "h3", "h4", "h5", "h6":
			buf.WriteString("\n")
		}
	}
}


func ExtractTextContent(n *html.Node) string {
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

func FindParentByTagName(n *html.Node, tagName string) *html.Node {
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

func FindNextSibling(n *html.Node) *html.Node {
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

func FindElementByClass(n *html.Node, className string) *html.Node {
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