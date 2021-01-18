package main

import (
	"archive/zip"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Docs struct {
	url      string
	fileName string
	path     string
	err      error
	buff     *[]byte
}

func main() {

	csvFile := flag.String("csv", "urls.csv", "the csv file that contains the document urls")
	maxWorkers := flag.Int("maxworker", 10, "maximum number of workers")
	logFlag := flag.String("log", "report", "value can be 'report' or 'all'")
	timeout := flag.Int("timeout", 5, "the number of timeout seconds for a download")
	flag.Parse()

	// 1. Load all Urls from CSV and create Docs struct
	docs, err := getDocs(*csvFile)
	if err != nil {
		log.Fatal(err)
	}

	// 2. Create Jobs channel and Results channel
	numJobs := len(docs)
	jobsCh := make(chan Docs, numJobs)
	resultsCh := make(chan Docs, numJobs)

	// 3. Start MaxWorker worker
	for i := 0; i < *maxWorkers; i++ {
		go func(id int, jobs chan Docs, results chan Docs) {
			for doc := range jobs {

				client := http.Client{
					Timeout: time.Duration(*timeout) * time.Second,
				}
				resp, err := client.Get(strings.TrimSpace(doc.url))
				if err != nil {
					log.Println(err)
					doc.err = err
					resultsCh <- doc
					continue
				}

				bs, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Println(err)
					doc.err = err
					resp.Body.Close()
					resultsCh <- doc
					continue
				}
				doc.buff = &bs
				resp.Body.Close()
				resultsCh <- doc
			}
		}(i, jobsCh, resultsCh)
	}

	// 4. Add jobs to the Jobs channel
	start := time.Now().UnixNano()
	for _, doc := range docs {
		jobsCh <- doc
	}
	close(jobsCh)

	// 5. Reading Reults channel and build Zip file
	f, err := os.Create("documents.zip")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	su, er := 0, 0
	for a := 0; a < numJobs; a++ {
		doc := <-resultsCh
		if doc.err != nil {
			er++
			log.Printf("error: %s bad file\n", doc.fileName)
			continue
		}
		su++
		addToZip(zw, doc)
		if *logFlag == "all" {
			log.Printf("Added to Zip: %s%s\n", doc.path, doc.fileName)
		}
	}
	zw.Close()
	end := time.Now().UnixNano()
	fmt.Printf("Found %d recodrds in %s CSV file.\n", len(docs), *csvFile)
	fmt.Printf("Download success: %d, failed: %d in time %d ms\n", su, er, (end-start)/int64(time.Millisecond))
}

func addToZip(zw *zip.Writer, doc Docs) {
	zf, err := zw.Create(doc.path + doc.fileName)
	if err != nil {
		log.Fatal(err)
	}

	_, err = zf.Write(*doc.buff)
	if err != nil {
		log.Fatal(err)
	}
}

func getDocs(fileName string) ([]Docs, error) {
	var docs []Docs
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	for {
		rec, err := r.Read()

		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		docs = append(docs, Docs{
			url:      rec[0],
			fileName: rec[1],
			path:     rec[2],
		})
	}
	return docs, nil
}
