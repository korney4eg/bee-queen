package report

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	collector "github.com/korney4eg/bee-queen/pkg/collector"
	logline "github.com/korney4eg/bee-queen/pkg/logline"
)

const (
	MessageLimit = 4000
)

type Command struct {
	Period         string `short:"p" long:"period" required:"true" choice:"any" choice:"day" choice:"week" choice:"month"`
	FileName       string `short:"f" long:"file" required:"false"`
	TelegramToken  string `short:"t" long:"telegram-token" required:"false"`
	TelegramChatId string `short:"c" long:"telegram-chat-id" required:"false"`
	DomainName     string `short:"d" long:"domain" required:"true"`
}

func (c *Command) Execute(_ []string) error {
	var err error
	var source io.Reader
	if c.FileName != "" {
		source, err = os.Open(c.FileName)

		if err != nil {
			return err
		}

	} else {
		source = os.Stdin
	}
	scanner := bufio.NewScanner(source)
	scanner.Split(bufio.ScanLines)
	collection := &collector.Collector{Domain: c.DomainName}

	for scanner.Scan() {
		var sll logline.SingleLogLine
		if err := sll.New(scanner.Text()); err != nil {
			return err
		}
		if !sll.MatchAllRequirements(c.Period, time.Now()) {
			continue
		}
		if err := collection.Accumulate(&sll); err != nil {
			return err
		}
	}
	// if o.FileName != "" {
	// 	source.Close()
	// }
	messages := make([]string, 0)
	msgLine := fmt.Sprintf("*%s*\n_Users: %d |  Hits: %d_\n", collection.Domain, collection.Users, collection.Hits)
	msgLine += fmt.Sprintf("*Popular pages*:\n```\n%+v\n```\n", collection.GetViews(collection.PageViews))
	msgLine += fmt.Sprintf("*Tags*:\n```\n%+v\n```\n", collection.GetViews(collection.TagViews))
	msgLine += fmt.Sprintf("*Referers*:\n```\n%+v```\n", collection.GetViews(collection.Referers))
	msgLine += fmt.Sprintf("*Browsers*:\n```\n%+v```\n", collection.GetViews(collection.ViewsByBrowser))
	msgLine += fmt.Sprintf("*OS*:\n```\n%+v```\n", collection.GetViews(collection.ViewsByOS))
	curMsg := ""
	properlyFinished := true
	for _, msg := range strings.Split(msgLine, "\n") {
		if len(curMsg+msg) > MessageLimit {
			if !properlyFinished {
				curMsg += "```\n"
			}
			messages = append(messages, curMsg)
			if !properlyFinished {
				curMsg = "```\n"
			} else {
				curMsg = ""
			}

		}
		curMsg += msg + "\n"
		if msg == "```" {
			properlyFinished = !properlyFinished
		}
	}
	messages = append(messages, curMsg)

	if c.TelegramToken != "" && c.TelegramChatId != "" {
		bot, err := tgbotapi.NewBotAPI(c.TelegramToken)
		bot.Debug = true
		if err != nil {
			return err
		}

		log.Printf("Authorized on account %s", bot.Self.UserName)
		n, err := strconv.ParseInt(c.TelegramChatId, 10, 64)
		if err != nil {
			log.Fatalf("Got error:%v\n", err)
		}

		for _, msg := range messages {
			message := tgbotapi.NewMessage(n, msg)
			message.ParseMode = "markdown"
			bot.Send(message)
			log.Println("Send message")
		}

	} else {
		for _, msg := range messages {
			fmt.Println(msg)
			fmt.Println("======================================================")
		}

	}
	return nil
}
