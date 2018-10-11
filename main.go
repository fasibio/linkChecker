package main

import (
	"bufio"
	"flag"
	"log"
	"net/http"
	"os"
	"regexp"
	"sync"
)

type deathLink struct {
	lineNumber int
	link       string
	httpStatus int
}

func main() {
	path := flag.String("f", "", "Path to markdown file")
	flag.Parse()
	if *path == "" {
		log.Fatal("Flag f is missing to set the path to the markdown file")
	}
	file, err := os.Open(*path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	var wg sync.WaitGroup

	for scanner.Scan() {
		wg.Add(1)
		lineNumber = lineNumber + 1
		go func(scanText string, lineNum int) {
			findLinksAndCheckHttpStatus(scanText, lineNum)
			wg.Done()
		}(scanner.Text(), lineNumber)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	log.Println("Wait till ", lineNumber, "go routines will end")
	wg.Wait()
}

type links = []string

func findLinksAndCheckHttpStatus(line string, lineNumber int) {
	links := findLinks(line)
	if len(links) != 0 {
		for _, one := range links {
			resp, err := http.Get(one)
			if err != nil {
				log.Println("line:", lineNumber, ":", one, "failed", "by error", err)
			} else {
				if resp.StatusCode != 200 {
					log.Println("line:", lineNumber, ":", one, "failed", "by statusCode", resp.StatusCode)
				}
			}
		}

	}
}

func findLinks(line string) links {
	re := regexp.MustCompile("(http|https)://([\\w_-]+(?:(?:\\.[\\w_-]+)+))([\\w.,@?^=%&:/~+#-]*[\\w@?^=%&/~+#-])?")

	match := re.FindAllString(line, -1)
	return match
}
