//go:build windows

package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/KalebHawkins/nwbot/bot"
)

var Version = ""
var version bool

func parseFlags() {
	flag.BoolVar(&version, "version", false, "display version information")
	flag.Parse()
}

func main() {
	parseFlags()

	if version {
		fmt.Printf("Version Commit: %s\n", Version)
		return
	}

	nwBot, err := bot.NewNwBot()
	if err != nil {
		log.Println(err)
	}

	if err := nwBot.Run(); err != nil {
		log.Println(err)
	}
}
