package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"pw-go2web/internal/browser"
	"pw-go2web/internal/client"
	"pw-go2web/internal/search"
	"strconv"
	"strings"
)

func main() {
	urlFlag := flag.String("u", "", "The URL to connect to.")
	searchFlag := flag.String("s", "", "The search term to look for.")
	helpFlag := flag.Bool("h", false, "Display help information.")

	flag.Parse()

	if *helpFlag || (flag.NFlag() == 0) {
		printHelp()
		return
	}

	if *urlFlag != "" {
		handleURLRequest(*urlFlag)
		return
	}

	if *searchFlag != "" {
		handleSearchRequest(*searchFlag)
		return
	}
}

func printHelp() {
	fmt.Println("Usage of go2web:")
	fmt.Println("  go2web -u <URL>         # make an HTTP request to the specified URL and print the response")
	fmt.Println("  go2web -s <search-term> # make an HTTP request to search the term using a search engine and print top 10 results")
	fmt.Println("  go2web -h               # show this help")
}

func handleURLRequest(url string) {
	response, err := client.MakeRequest(url, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(response)
}

func handleSearchRequest(term string) {
	results, err := search.SearchTerm(term)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Format the results
	for i, result := range results {
		fmt.Printf("%d. %s\n   URL: %s\n\n", i+1, result.Title, result.URL)
	}
	
	// Ask user if they want to open one of the results
	fmt.Print("Enter a number to open a result (1-10), or press Enter to exit: ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	
	if input != "" {
		// Try to convert the input to a number
		num, err := strconv.Atoi(input)
		if err != nil || num < 1 || num > len(results) {
			fmt.Println("Invalid selection.")
			return
		}
		
		// Open the URL in the default browser
		selectedURL := results[num-1].URL
		fmt.Printf("Opening: %s\n", selectedURL)
		browser.OpenBrowser(selectedURL)
	}
}