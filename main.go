package main

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

func FetchContent(url string) ([]byte, error) {
	resp, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return body, nil
}

func HashRemoteContent(url string) string {
	content, err := FetchContent(url)
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("%x", md5.Sum(content))
}

type IndexedContent struct {
	index   int
	content string
}

func worker(id int, jobs <-chan IndexedContent, results chan<- IndexedContent) {
	for job := range jobs {
		results <- IndexedContent{index: job.index, content: HashRemoteContent(job.content)}
	}
}

func HashAllRemoteContent(numWorkers int, urls []string) []string {
	res := make([]string, len(urls))
	jobs := make(chan IndexedContent, len(urls))
	results := make(chan IndexedContent, len(urls))

	for w := 1; w <= numWorkers; w++ {
		go worker(w, jobs, results)
	}

	for i, url := range urls {
		jobs <- IndexedContent{index: i, content: url}
	}
	close(jobs)

	for a := 0; a < len(urls); a++ {
		ans := <-results
		res[ans.index] = ans.content
	}

	return res
}

func main() {
	filename := os.Args[1]
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	urls := make([]string, 0)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	var numWorkers int64 = 100
	if len(os.Args) > 2 {
		numWorkers, err = strconv.ParseInt(os.Args[2], 10, 0)
		if err != nil {
			panic(err)
		}
	}

	for _, e := range HashAllRemoteContent(int(numWorkers), urls) {
		fmt.Println(e)
	}
}
