package export

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/korney4eg/bee-queen/pkg/collector"
	logline "github.com/korney4eg/bee-queen/pkg/logline"
)

type Command struct {
	// Period        string `short:"p" long:"period" required:"true" choice:"any" choice:"day" choice:"week" choice:"month"`
	Destination   string `short:"d" long:"destination" required:"false" default:"." description:"Folder where files should be saved"`
	Day           string `short:"o" long:"only-day" required:"false" description:"Get statistics only for provided date. Example '01.02.2020'"`
	SplitPerYear  bool   `short:"y" long:"year-split" required:"true" description:"Will split files by year"`
	SplitPerMonth bool   `short:"m" long:"month-split" required:"true" description:"Will split files by month"`
}

func (c *Command) Execute(_ []string) error {
	// var err error
	var source io.Reader
	source = os.Stdin
	scanner := bufio.NewScanner(source)
	scanner.Split(bufio.ScanLines)
	stats := make(map[string]collector.Collector)
	stats_raw := make(map[string][]string)

	// fmt.Println(stats)
	for scanner.Scan() {
		var sll logline.SingleLogLine
		if err := sll.New(scanner.Text()); err != nil {
			log.Fatal(err)
		}
		if !sll.MatchAllWithoutPeriod() {
			continue
		}
		t, _ := time.Parse("02/Jan/2006:15:04:05 -0700", sll.TimeLocal)
		year, month, day := t.Date()
		key := fmt.Sprintf("%d-%02d-%02d", year, int(month), day)
		if c.Day != "" {
			res1, _ := regexp.MatchString("^[0-9][0-9].[0-9][0-9].20[0-9][0-9]$", c.Day)
			if !res1 {
				log.Fatalf("Wrong date format '%s'. Example '01.02.2020'", c.Day)
			}
			day_t, _ := time.Parse("02.01.2006", c.Day)
			d_year, d_month, d_day := day_t.Date()
			if fmt.Sprintf("%d-%02d-%02d", d_year, int(d_month), d_day) != key {
				continue
			}
		}
		stats_raw[key] = append(stats_raw[key], scanner.Text())
	}
	filesWritten := 0
	for date, lines := range stats_raw {
		destFolder := c.Destination + "/"
		if c.SplitPerYear {
			year := strings.Split(date, "-")[0]
			destFolder += year + "/"
		}
		if c.SplitPerMonth {
			month := strings.Split(date, "-")[1]
			destFolder += month + "/"
		}
		err := os.MkdirAll(filepath.Dir(destFolder), os.FileMode(int(0755)))
		if err != nil {
			log.Fatal(err)
		}

		collection := &collector.Collector{Domain: date}
		for _, line := range lines {
			var sl logline.SingleLogLine
			if err := sl.New(line); err != nil {
				log.Fatal(err)
			}
			if err := collection.Accumulate(&sl); err != nil {
				log.Fatal(err)
			}
			stats[date] = *collection

		}
		file, err := json.MarshalIndent(*collection, "", " ")
		if err != nil {
			log.Fatal(err)
		}
		err = ioutil.WriteFile(destFolder+date+".json", file, 0644)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Creating file: %s", destFolder+date+".json")
		filesWritten += 1

	}
	log.Printf("Files written: %d\n", filesWritten)
	return nil
}
