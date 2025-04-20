package main

import (
	"flag"
	"fmt"
)

func main() {
	// urlFlag := flag.String("u", "", "The URL to connect to.")
	// searchFlag := flag.String("s", "", "The search term to look for.")
	helpFlag := flag.Bool("h", false, "Display help information.")

	flag.Parse()

	if *helpFlag || (flag.NFlag() == 0) {
		fmt.Println("Usage of go2web:")
		fmt.Println("  go2web -u <URL>         # make an HTTP request to the specified URL and print the response")
		fmt.Println("  go2web -s <search-term> # make an HTTP request to search the term using a search engine and print top 10 results")
		fmt.Println("  go2web -h               # show this help")
		return
	}

	

}
