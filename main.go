package main

import (
	"fmt"

	"github.com/Apollo-Community/slackbot/src"
)

func main() {
	fmt.Println("Running slackbot...")
	bot := slackbot.NewInstance(false, "U0JPJSUJU")
	bot.Run()

}
