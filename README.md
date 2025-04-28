# go2web

A simple CLI tool for making HTTP requests and searching the web from the command line.

## Overview

go2web allows you to:

- Make HTTP requests to websites and view their content
- Search the web using DuckDuckGo and open results in your browser

## Installation

### Build from Source

1. Clone the repository:

   ```bash
   git clone https://github.com/Masha003/pw-go2web.git
   cd pw-go2web
   ```

2. Build the application:

   ```bash
   go build -o go2web ./cmd/
   ```

3. Make the binary executable (Linux/macOS):

   ```bash
   chmod +x go2web
   ```

4. (Optional) Move the binary to your PATH for system-wide access:

   ```bash
   # Linux/macOS
   sudo mv go2web /usr/local/bin/

   # Or add to PATH in your shell profile (~/.bashrc or ~/.zshrc)
   echo 'export PATH="$PATH:$(pwd)"' >> ~/.bashrc  # For Bash
   # OR
   echo 'export PATH="$PATH:$(pwd)"' >> ~/.zshrc   # For Zsh
   source ~/.bashrc  # Or ~/.zshrc
   ```

   On Windows, you can add the directory to your PATH environment variable.

## Usage

The program displays the following menu:

```
go2web -u <URL>         # make an HTTP request to the specified URL and print the response
go2web -s <search-term> # make an HTTP request to search the term using your favorite search engine and print top 10 results
go2web -h               # show this help
```

### Get Help

```bash
go2web -h
```

### Make an HTTP Request

```bash
go2web -u example.com
```

The tool will display the processed response content. It handles HTML and JSON responses appropriately.

### Search the Web

```bash
go2web -s "golang"
```

This will show the top 10 search results and allow you to open one in your browser.

## Gif with working example
