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

func SendReport(fileName, domainName, period, telegramToken, telegramChatId string) {
	var err error
	var source io.Reader
	if fileName != "" {
		source, err = os.Open(fileName)

		if err != nil {
			log.Fatalf("failed opening file: %s", err)
		}

	} else {
		source = os.Stdin
	}
	scanner := bufio.NewScanner(source)
	scanner.Split(bufio.ScanLines)
	collection := &collector.Collector{Domain: domainName}

	for scanner.Scan() {
		var sll logline.SingleLogLine
		if err := sll.New(scanner.Text()); err != nil {
			log.Fatalf("failed opening file: %s", err)
		}
		if !sll.MatchAllRequirements(period) {
			continue
		}
		if err := collection.Accumulate(&sll); err != nil {
			log.Panic(err)
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
	if telegramToken != "" && telegramChatId != "" {
		bot, err := tgbotapi.NewBotAPI(telegramToken)
		if err != nil {
			log.Panic(err)
		}

		log.Printf("Authorized on account %s", bot.Self.UserName)
		n, err := strconv.ParseInt(telegramChatId, 10, 64)
		if err == nil {
			fmt.Printf("%d of type %T", n, n)
		}

		message := tgbotapi.NewMessage(n, msg)
		message.ParseMode = "markdown"
		bot.Send(message)

	} else {
		fmt.Println(msg)

	}
}
