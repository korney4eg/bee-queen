package filter

import (
	"fmt"
	"regexp"
	"time"
)

type SingleLogLine struct {
	RemoteAddr    string
	TimeLocal     string
	Request       string
	Status        string
	BodyBytesSent string
	HTTPReferer   string
	HTTPUserAgent string
}

func (obj *SingleLogLine) New(logLine string) (err error) {
	regexPattern := `(?P<remote_addr>\d+\.\d+\.\d+\.\d+) - -` +
		` \[(?P<time_local>[^\]]+)\] \"(?P<request>.*)\" (?P<status>[0-9]+)` +
		` (?P<body_bytes_sent>[0-9]+) \"(?P<http_referer>.*)\" \"` +
		`(?P<http_user_agent>.+)\"`
	re := regexp.MustCompile(regexPattern)
	// log.Println(logLine)
	match := re.FindStringSubmatch(logLine)
	if match == nil {
		return nil
	}
	for i, name := range re.SubexpNames() {
		if i != 0 {
			switch name {
			case "remote_addr":
				obj.RemoteAddr = match[i]
			case "time_local":
				obj.TimeLocal = match[i]
			case "request":
				// obj.Request = strings.TrimLeft(match[i], "GET ")
				request := match[i]
				obj.Request = request
			case "status":
				obj.Status = match[i]
			case "body_bytes_sent":
				obj.BodyBytesSent = match[i]
			case "http_referer":
				obj.HTTPReferer = match[i]
			case "http_user_agent":
				obj.HTTPUserAgent = match[i]
			}
		}
	}
	return nil
}

func (obj *SingleLogLine) MatchAllRequirements(period string, endTime time.Time) bool {
	switch {
	case !obj.MatchAllWithoutPeriod() || !dateIsInInterval(obj.TimeLocal, period, endTime):
		return false
	default:
		return true
	}
}
func (obj *SingleLogLine) MatchAllWithoutPeriod() bool {
	request := regexp.MustCompile(`GET \/(\d{4}\/\d{2}\/\d{2}\/[^\/]+|tags\/[^\/]+|about|archives\/[^\/]+)\/ HTTP\/[12]\.[10]`)
	http_user_agent := regexp.MustCompile(`.*([Bb]ot|vkShare|Google-AMPHTML|feedly|[cC]rawler|[Pp]arser|curl|-|[Dd]isqus|[Dd]isqus|Daum|[Ss]pider|[Qq]wantify|facebookexternalhit|http\.rb|Anthill|okhttp|aria2).*`)
	switch {
	case obj.Status != "200":
		return false
	case !request.MatchString(obj.Request):
		return false
	case http_user_agent.MatchString(obj.HTTPUserAgent):
		return false
	default:
		return true
	}
}
func ConvertMapToLogLine(parsedLine map[string]string) string {
	remote_addr := parsedLine["remote_addr"]
	time_local := parsedLine["time_local"]
	request := parsedLine["request"]
	status := parsedLine["status"]
	body_bytes_sent := parsedLine["body_bytes_sent"]
	http_referer := parsedLine["http_referer"]
	http_user_agent := parsedLine["http_user_agent"]
	return fmt.Sprintf(`%s - - [%s] "%s" %s %s "%s" "%s"`, remote_addr,
		time_local, request, status, body_bytes_sent, http_referer,
		http_user_agent)
}

func GetStartDate(period string, endTime time.Time) (startDate time.Time) {
	switch period {
	case "day":
		return endTime.AddDate(0, 0, -1)
	case "week":
		return endTime.AddDate(0, 0, -7)
	case "month":
		return endTime.AddDate(0, -1, 0)
	case "year":
		return endTime.AddDate(-1, 0, 0)
	default:
		return time.Now()
	}

}

func dateIsInInterval(line string, period string, endTime time.Time) bool {
	var startDate time.Time
	switch period {
	case "week":
		startDate = GetStartDate(period, endTime)

	case "month":
		startDate = GetStartDate(period, endTime)
	case "day":
		startDate = GetStartDate(period, endTime)
	case "year":
		startDate = GetStartDate(period, endTime)
	case "any":
		return true
	default:
		return false
	}
	t, _ := time.Parse("02/Jan/2006:15:04:05 -0700", line)
	// if endTime.After(t) && startDate.Before(t) {
	// 	log.Printf("| %v === %s === %v", endTime, line, startDate)
	// }
	return endTime.After(t) && startDate.Before(t)
}
