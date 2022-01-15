package main

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
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
		fmt.Println("worker", id, "started  job", job.index, " ", job.content)
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
	urls := []string{"http://example.com/",
		"https://github.githubassets.com/images/mona-loading-default.gif",
		"https://upload.wikimedia.org/wikipedia/commons/d/d3/Trex_pixel.png"}

	expected := []string{
		"84238dfc8092e5d9c0dac8ef93371a07",
		"c502cd01c910b4f53d86603d6bd078ff",
		"e9d2e589fdf6b0e8517aa22c1c8b89a4"}

	fmt.Println(reflect.DeepEqual(expected, HashAllRemoteContent(2, urls)))
}
