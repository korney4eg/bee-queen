package main

import (
	"os"

	"fmt"

	"github.com/jessevdk/go-flags"
	filter "github.com/korney4eg/bee-queen/cmd/filter"
	report "github.com/korney4eg/bee-queen/cmd/send-report"
)

type FilterCmd struct{}

type opts struct {
	SendReport report.Command `command:"send-report"`
	Filter     filter.Command `command:"filter"`
}

func (f *FilterCmd) Execute(_ []string) error {
	fmt.Println("Filtered ...")
	return nil
}

func main() {
	o := opts{}
	if _, err := flags.Parse(&o); err != nil {
		os.Exit(1)
	}
}
