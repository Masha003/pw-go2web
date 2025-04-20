package main

import (
	"flag"
	"fmt"
	"os"
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


func makeRequest(url string) (string, error) {
	fmt.Println(url)
	return "", nil
}

func searchTerm(term string) (string, error) {
	fmt.Println(term)
	return "", nil
}
