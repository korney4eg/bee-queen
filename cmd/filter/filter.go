package filter

import (
	"bufio"
	"fmt"
	"io"
	"os"

	logline "github.com/korney4eg/bee-queen/pkg/logline"
)

type Command struct {
	Period   string `short:"p" long:"period" required:"true" choice:"any" choice:"day" choice:"week" choice:"month"`
	FileName string `short:"f" long:"file" required:"false"`
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

	for scanner.Scan() {
		var sll logline.SingleLogLine
		if err := sll.New(scanner.Text()); err != nil {
			return err
		}
		if !sll.MatchAllRequirements(c.Period) {
			continue
		}
		fmt.Println(scanner.Text())
	}
	return nil
}
