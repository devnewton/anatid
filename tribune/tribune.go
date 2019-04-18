package tribune 

import (
	"encoding/csv"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
)

//PollTsv Retrieve TSV backend
func PollTsv(backendURL string, tribuneName string) (posts Posts, err error) {
	resp, err := http.Get(backendURL)
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
		post := Post{ID: id, Time: record[1], Info: record[2], Login: record[3], Message: record[4], Tribune: tribuneName}
		posts = append(posts, post)
	}
	sort.Sort(posts)
	return posts, nil
}
