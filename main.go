package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Apollo-Community/slackbot/src"
)

var f_debug = flag.Bool("debug", false, "debug mode")

func main() {
	flag.Parse()
	token := os.Getenv("SLACKBOT_TOKEN")

	fmt.Println(`
  ▄▄▄▄▄▄▄▄                ▄▄     ▄▄▄▄▄▄                
▄█▀▀▀▀▀▀██                ██     ██▀▀▀▀██       ██    
██▄     ██   ▄█████▄▄███████ ▄██▀██    ██▄███████████ 
 ▀████▄ ██   ▀ ▄▄▄███▀    ██▄██  █████████▀  ▀████    
     ▀████  ▄██▀▀▀███     ██▀██▄ ██    ███    ████    
█▄▄▄▄▄█▀██▄▄██▄▄▄██▀██▄▄▄▄██  ▀█▄██▄▄▄▄█▀██▄▄██▀██▄▄▄ 
 ▀▀▀▀▀   ▀▀▀▀▀▀▀▀ ▀▀ ▀▀▀▀▀▀▀   ▀▀▀▀▀▀▀▀▀  ▀▀▀▀   ▀▀▀▀ V` + slackbot.VERSION)
	bot := slackbot.NewInstance(token, *f_debug, "U0JPJSUJU")
	bot.Run()

}
