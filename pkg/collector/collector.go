package collector

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	logline "github.com/korney4eg/bee-queen/pkg/logline"
	ua "github.com/mileusna/useragent"
)

const (
	maxLineLength = 45
)

type Collector struct {
	Hits           int `json:"hits"`
	Users          int `json:"users"`
	Domain         string
	UsersList      []string       `json:"users_list"`
	PageViews      map[string]int `json:"views_by_page"`
	ViewsByBrowser map[string]int `json:"views_by_browser"`
	ViewsByOS      map[string]int `json:"views_by_os"`
	TagViews       map[string]int `json:"views_by_tag"`
	ArchiveViews   map[string]int `json:"views_by_archive"`
	Referers       map[string]int `json:"referers"`
}

type PageViews struct {
	Page  string
	Views int
}

func (col *Collector) Accumulate(line *logline.SingleLogLine) error {
	col.Hits++
	uniqUserName := fmt.Sprintf("%x", sha256.Sum256([]byte(line.RemoteAddr+line.HTTPUserAgent)))
	userIn := false
	for _, user := range col.UsersList {
		if user == uniqUserName {
			userIn = true
			break
		}
	}
	if !userIn {
		col.UsersList = append(col.UsersList, uniqUserName)
		col.Users++
	}
	page := strings.Split(line.Request, " ")[1]
	if col.PageViews == nil {
		col.PageViews = make(map[string]int)
	}
	if col.TagViews == nil {
		col.TagViews = make(map[string]int)
	}
	if col.ArchiveViews == nil {
		col.ArchiveViews = make(map[string]int)
	}
	decodedPage, err := url.QueryUnescape(page)
	if err != nil {
		return err
	}
	if strings.HasPrefix(decodedPage, "/tags/") {
		col.TagViews[strings.Split(decodedPage, "/")[2]] += 1
	} else if strings.HasPrefix(decodedPage, "/archives/") {
		col.ArchiveViews[strings.Split(decodedPage, "/")[2]] += 1
	} else if decodedPage != "/about/" {
		col.PageViews[strings.Split(decodedPage, "/")[4]] += 1
	} else {
		col.PageViews[decodedPage] += 1
	}

	ua := ua.Parse(line.HTTPUserAgent)

	if col.ViewsByBrowser == nil {
		col.ViewsByBrowser = make(map[string]int)
	}
	col.ViewsByBrowser[ua.Name] += 1
	if col.ViewsByOS == nil {
		col.ViewsByOS = make(map[string]int)
	}
	col.ViewsByOS[ua.OS] += 1

	if col.Referers == nil {
		col.Referers = make(map[string]int)
	}
	ref := ""
	if line.HTTPReferer == "-" {
		ref = "-"
	} else {
		ref = strings.Split(line.HTTPReferer, "/")[2]
	}
	col.Referers[ref] += 1
	return nil
}

func (col *Collector) GetViews(obj map[string]int) (views string) {
	p := make(PairList, len(obj))
	i := 0
	for k, v := range obj {
		p[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(p))

	buf := new(bytes.Buffer)
	w := tabwriter.NewWriter(buf, 0, 0, 3, ' ', tabwriter.TabIndent)
	tmpStr := ""
	for _, k := range p {
		if len(k.Key) > maxLineLength {
			tmpStr = k.Key[0:maxLineLength]
		} else {
			tmpStr = k.Key
		}
		fmt.Fprintln(w, tmpStr+"\t"+strconv.Itoa(k.Value))
	}
	w.Flush()
	return buf.String()
}

// A data structure to hold key/value pairs
type Pair struct {
	Key   string
	Value int
}

// A slice of pairs that implements sort.Interface to sort by values
type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }

func (col *Collector) GetHits() int {
	return col.Hits
}

func (col *Collector) GetUsers() int {
	return col.Users
}
