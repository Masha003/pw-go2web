package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
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

	request := fmt.Sprintf("GET %s HTTP/1.1\r\nHost: %s\r\nConnection: close\r\nUser-Agent: go2web/1.0\r\nAccept: text/html,application/json\r\n\r\n", path, host)
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

	body := parts[1]

	
	return body

}


func searchTerm(term string) (string, error) {
	fmt.Println(term)
	return "", nil
}
