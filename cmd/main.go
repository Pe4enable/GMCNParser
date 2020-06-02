package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
)

var (
	URL          = flag.String("url", "https://gmcngine-api.globalmissingkids.org/api/cases/search", "Default search URL")
	URLCase      = flag.String("urlcase", "https://gmcngine-api.globalmissingkids.org/api/cases/", "URL Prefix for case search")
	SearchString = flag.String("search", "{\"request\":{\"page\":0,\"size\":1512,\"sort\":[{\"missingSince\":\"desc\"},\"fullName\"],\"search\":\"\",\"status\":\"open\"}}", "Search string")
	Origin       = flag.String("origin", "https://find.globalmissingkids.org", "Value for HTTP Origin header")
	Referer      = flag.String("referer", "https://find.globalmissingkids.org/", "Value for HTTP Referer header")
	InDataFile   = flag.String("indatafile", "", "Use this file as input instead of network request")
	OutDataFile  = flag.String("outdatafile", "", "Save into file result of network request")
	Output       = flag.String("output", "../output/output.csv", "Name of output resulting CSV file")
	CacheDir     = flag.String("cachedir", "../output/cache", "Directory for storing image cache")

	NumConn = flag.Int("numconn", 10, "Number of simultaneous image retrieving instances")
)

func main() {
	ctx := context.Background()
	ctxWithCancel, cancelFunction := context.WithCancel(ctx)

	flag.Parse()

	content, err := getData()
	if err != nil {
		log.Fatal(err)
	}
	if *OutDataFile != "" {
		saveData(*OutDataFile, content)
	}

	var resultData SearchCasesResult
	err = json.Unmarshal([]byte(content), &resultData)
	if err != nil {
		// fmt.Print("Response: ", content)
		log.Fatal("Error JSON unmarshal: ", err)
	}

	LoaderChan := make(chan ResultInfo)
	ReadyChan := make(chan []string)
	ErrChan := make(chan error)

	// Run data loaders
	for i := 0; i < *NumConn; i++ {
		go resolver(ctxWithCancel, LoaderChan, ReadyChan, ErrChan)
	}

	// Create output file
	if *Output == "" {
		log.Fatal("Output file is not specified, skipping result generation")
	}
	fOut, err := os.OpenFile(*Output, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		log.Fatal("Error creating output file [", *Output, "]: ", err)
	}
	csvWriter := csv.NewWriter(fOut)
	defer fOut.Close()

	// Start OS TERMINATE REQUEST processor
	cOsTerminate := make(chan os.Signal, 1)
	signal.Notify(cOsTerminate, os.Interrupt)

	// Start sender
	go sender(ctxWithCancel, resultData.Cases.Results, LoaderChan)

	// Scan for writer
	wCount := 0
	for {
		select {
		case r := <-ErrChan:
			fmt.Printf("[ERROR] %v\n", r)
			wCount++

		case r := <-ReadyChan:
			if err := csvWriter.Write(r); err != nil {
				log.Fatal("Fatal error writing CSV data: ", err)
			}
			wCount++

			if wCount >= len(resultData.Cases.Results) {
				fmt.Println("Work done!\n")
				csvWriter.Flush()
				return
			}

		case <-cOsTerminate:
			cancelFunction()
			fmt.Println("OS Terminate requested")
		}
	}
}
