package collector

import (
	"crypto/sha256"
	"fmt"
	"net/url"
	"sort"
	"strings"

	logline "github.com/korney4eg/bee-queen/pkg/logline"
	ua "github.com/mileusna/useragent"
)

type Collector struct {
	Hits           int
	Users          int
	usersList      []string
	ViewsByPage    map[string]int
	ViewsByBrowser map[string]int
	ViewsByOS      map[string]int
}

type PageViews struct {
	Page  string
	Views int
}

func (col *Collector) Accumulate(line *logline.SingleLogLine) error {
	col.Hits++
	uniqUserName := fmt.Sprintf("%x", sha256.Sum256([]byte(line.RemoteAddr+line.HTTPUserAgent)))
	userIn := false
	for _, user := range col.usersList {
		if user == uniqUserName {
			userIn = true
			break
		}
	}
	if !userIn {
		col.usersList = append(col.usersList, uniqUserName)
		col.Users++
	}
	page := strings.Split(line.Request, " ")[1]
	if col.ViewsByPage == nil {
		col.ViewsByPage = make(map[string]int)
	}
	decodedPage, err := url.QueryUnescape(page)
	if err != nil {
		return err
	}
	col.ViewsByPage[decodedPage] += 1

	ua := ua.Parse(line.HTTPUserAgent)

	if col.ViewsByBrowser == nil {
		col.ViewsByBrowser = make(map[string]int)
	}
	col.ViewsByBrowser[ua.Name] += 1
	if col.ViewsByOS == nil {
		col.ViewsByOS = make(map[string]int)
	}
	col.ViewsByOS[ua.OS] += 1
	return nil
}

func (col *Collector) GetViewsByPage() (views string) {
	p := make(PairList, len(col.ViewsByPage))
	i := 0
	for k, v := range col.ViewsByPage {
		p[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(p))
	for _, k := range p {
		views += fmt.Sprintf("%s:%d\n", k.Key, k.Value)
	}
	return views
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
