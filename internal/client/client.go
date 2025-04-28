package client

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"pw-go2web/internal/parser"
	"regexp"
	"strconv"
	"strings"
)

// func MakeRequest(url string) (string, error) {
// 	return MakeRequestRaw(url, true)
// }

func MakeRequest(url string, processHtml bool) (string, error) {
	scheme, host, path, port, err := ParseURL(url)
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

	// Read the response headers first
	var statusCode int
	var statusLine string
	var headers = make(map[string]string)
	
	// Read the status line
	statusLine, err = reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error reading status line: %v", err)
	}
	response.WriteString(statusLine)
	
	// Extract status code
	statusMatch := regexp.MustCompile(`HTTP/[\d.]+\s+(\d+)`).FindStringSubmatch(statusLine)
	if len(statusMatch) > 1 {
		statusCode, _ = strconv.Atoi(statusMatch[1])
	}

		// Read headers
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return "", fmt.Errorf("error reading header: %v", err)
			}
			response.WriteString(line)
			
			// Store headers in a map
			line = strings.TrimSpace(line)
			if line == "" {
				break // End of headers
			}
			
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				headerName := strings.TrimSpace(parts[0])
				headerValue := strings.TrimSpace(parts[1])
				headers[strings.ToLower(headerName)] = headerValue
			}
		}
		
		// Check for redirection (status codes 301, 302, 303, 307, 308)
		if statusCode >= 300 && statusCode < 400 {
			location, ok := headers["location"]
			if ok {
				fmt.Printf("Following redirect to: %s\n", location)
				
				// Handle relative URLs
				if !strings.HasPrefix(location, "http") {
					if strings.HasPrefix(location, "/") {
						// Absolute path
						location = fmt.Sprintf("%s://%s%s", scheme, host, location)
					} else {
						// Relative path
						basePath := path
						if idx := strings.LastIndex(basePath, "/"); idx != -1 {
							basePath = basePath[:idx+1]
						} else {
							basePath = "/"
						}
						location = fmt.Sprintf("%s://%s%s%s", scheme, host, basePath, location)
					}
				}
				
				// Recursively follow the redirect
				return MakeRequest(location, processHtml)
			}
		}
		
		// Read the body
		var bodyBuilder strings.Builder
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					// We've reached the end of the response
					break
				}
				return "", fmt.Errorf("error reading response: %v", err)
			}
			bodyBuilder.WriteString(line)
		}
		
		// Append the body to the full response
		response.WriteString(bodyBuilder.String())
		
		// If we don't want to process the HTML, extract the body and return it
		if !processHtml {
			return bodyBuilder.String(), nil
		}
	
		// Otherwise process as usual
		return parser.ProcessResponse(response.String()), nil
}


func ParseURL(rawURL string) (string, string, string, int, error) {
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