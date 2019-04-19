package tribune

import "encoding/xml"

//Post tribune post
type Post struct {
	ID      int64  `json:"id" xml:"id,attr"`
	Time    string `json:"time" xml:"time,attr"`
	Info    string `json:"info" xml:"info"`
	Login   string `json:"login" xml:"login"`
	Message string `json:"message" xml:"message"`
	Tribune string `json:"tribune"`
}

//Posts post array
type Posts []Post

//Board struct for xml parsing
type Board struct {
	XMLName xml.Name `xml:"board"`
	Posts   Posts    `xml:"post"`
}

func (p Posts) Len() int           { return len(p) }
func (p Posts) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p Posts) Less(i, j int) bool { return p[i].ID < p[j].ID }
