package tribune

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

const (
	// TSVBackend denotes a TSV tribune backend
	TSVBackend = 0

	// XMLBackend denotes a TSV tribune backend
	XMLBackend = 1
)

const (
	// NoAuthentification Anonymous post
	NoAuthentification = 0
	// OAuth2Authentification annoying dlfp authentification
	OAuth2Authentification = 1
)

// Tribune parameters
type Tribune struct {
	Name                 string
	BackendURL           string
	PostURL              string
	PostField            string
	BackendType          int
	AuthentificationType int
}

type dlfpToken struct {
	accessToken string `json:"access_token"`
}

// Post message to tribune
func (tribune *Tribune) Post(inRequest *http.Request) {
	data := url.Values{
		tribune.PostField: []string{inRequest.PostFormValue("message")},
	}
	if tribune.AuthentificationType == OAuth2Authentification {
		var token dlfpToken
		err := json.Unmarshal([]byte(inRequest.PostFormValue("auth")), token)
		if nil == err {
			data.Set("bearer_token", token.accessToken)
		} else {
			log.Println(err)
		}
	}
	outRequest, err := http.NewRequest("POST", tribune.PostURL, strings.NewReader(data.Encode()))
	if nil != err {
		log.Println(err)
		return
	}
	outRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	outRequest.Header.Set("User-Agent", inRequest.Header.Get("User-Agent"))
	_, err = http.DefaultClient.Do(outRequest)
	if nil != err {
		log.Println(err)
	}
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
