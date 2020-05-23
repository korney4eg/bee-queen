package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jessevdk/go-flags"
	collector "github.com/korney4eg/bee-queen/pkg/collector"
	logline "github.com/korney4eg/bee-queen/pkg/logline"
)

type opts struct {
	Period         string `short:"p" long:"period" required:"true" choice:"any" choice:"day" choice:"week" choice:"month"`
	FileName       string `short:"f" long:"file" required:"false"`
	TelegramToken  string `short:"t" long:"telegram-token" required:"false"`
	TelegramChatId string `short:"c" long:"telegram-chat-id" required:"false"`
}

func main() {
	o := opts{}
	if _, err := flags.Parse(&o); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var err error
	var source io.Reader
	if o.FileName != "" {
		source, err = os.Open(o.FileName)

		if err != nil {
			log.Fatalf("failed opening file: %s", err)
		}

	} else {
		source = os.Stdin
	}
	scanner := bufio.NewScanner(source)
	scanner.Split(bufio.ScanLines)
	collection := &collector.Collector{}

	for scanner.Scan() {
		var sll logline.SingleLogLine
		if err := sll.New(scanner.Text()); err != nil {
			log.Fatalf("failed opening file: %s", err)
		}
		if !sll.MatchAllRequirements("any") {
			continue
		}
		if err := collection.Accumulate(&sll); err != nil {
			log.Panic(err)
		}
	}
	// if o.FileName != "" {
	// 	source.Close()
	// }
	msg := fmt.Sprintf("*Users*: %d |||  *Hits*: %d\n*Popular pages*: %+v\n", collection.Users, collection.Hits, collection.GetViewsByPage())
	if o.TelegramToken != "" && o.TelegramChatId != "" {
		bot, err := tgbotapi.NewBotAPI(o.TelegramToken)
		if err != nil {
			log.Panic(err)
		}

		log.Printf("Authorized on account %s", bot.Self.UserName)
		n, err := strconv.ParseInt(o.TelegramChatId, 10, 64)
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
