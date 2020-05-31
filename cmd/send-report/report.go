package report

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	collector "github.com/korney4eg/bee-queen/pkg/collector"
	logline "github.com/korney4eg/bee-queen/pkg/logline"
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
		if !sll.MatchAllRequirements(c.Period) {
			continue
		}
		if err := collection.Accumulate(&sll); err != nil {
			return err
		}
	}
	// if o.FileName != "" {
	// 	source.Close()
	// }
	msg := fmt.Sprintf("*%s*\n_Users: %d |  Hits: %d_\n", collection.Domain, collection.Users, collection.Hits)
	msg += fmt.Sprintf("*Popular pages*:\n```\n%+v\n", collection.GetViews(collection.PageViews))
	msg += fmt.Sprintf("*Rags*:\n```\n%+v\n```\n", collection.GetViews(collection.TagViews))
	msg += fmt.Sprintf("*Referers*:\n```\n%+v```\n", collection.GetViews(collection.Referers))
	msg += fmt.Sprintf("*Browsers*:\n```\n%+v```\n", collection.GetViews(collection.ViewsByBrowser))
	msg += fmt.Sprintf("*OS*:\n```\n%+v```\n", collection.GetViews(collection.ViewsByOS))
	if c.TelegramToken != "" && c.TelegramChatId != "" {
		bot, err := tgbotapi.NewBotAPI(c.TelegramToken)
		if err != nil {
			return err
		}

		log.Printf("Authorized on account %s", bot.Self.UserName)
		n, err := strconv.ParseInt(c.TelegramChatId, 10, 64)
		if err == nil {
			return err
		}

		message := tgbotapi.NewMessage(n, msg)
		message.ParseMode = "markdown"
		bot.Send(message)

	} else {
		fmt.Println(msg)

	}
	return nil
}
