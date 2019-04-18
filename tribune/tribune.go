package tribune

import (
	"encoding/csv"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
)

const (
	// TSVBackend denotes a TSV tribune backend
	TSVBackend = 0

	// XMLBackend denotes a TSV tribune backend
	XMLBackend = 1
)

// Tribune parameters
type Tribune struct {
	Name        string
	BackendURL  string
	BackendType int
}

//Poll Retrieve backend
func (tribune *Tribune) Poll() (posts Posts, err error) {
	resp, err := http.Get(tribune.BackendURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	r := csv.NewReader(resp.Body)
	r.Comma = '\t'
	r.LazyQuotes = true

	posts = make(Posts, 0, 200)
	for {
		record, err := r.Read()
		if nil != err {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}
		if nil == record {
			break
		}
		if len(record) < 5 {
			continue
		}
		id, err := strconv.ParseInt(record[0], 10, 64)
		if nil != err {
			log.Println(err)
			continue
		}
		post := Post{ID: id, Time: record[1], Info: record[2], Login: record[3], Message: record[4], Tribune: tribune.Name}
		posts = append(posts, post)
	}
	sort.Sort(posts)
	return posts, nil
}
