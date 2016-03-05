package main

import (
	"flag"
	"fmt"

	"github.com/Apollo-Community/slackbot/src"
)

var f_token = flag.String("token", "", "slack bot token")
var f_debug = flag.Bool("debug", false, "debug mode")

func main() {
	flag.Parse()

	fmt.Println(`
  ▄▄▄▄▄▄▄▄                ▄▄     ▄▄▄▄▄▄                
▄█▀▀▀▀▀▀██                ██     ██▀▀▀▀██       ██    
██▄     ██   ▄█████▄▄███████ ▄██▀██    ██▄███████████ 
 ▀████▄ ██   ▀ ▄▄▄███▀    ██▄██  █████████▀  ▀████    
     ▀████  ▄██▀▀▀███     ██▀██▄ ██    ███    ████    
█▄▄▄▄▄█▀██▄▄██▄▄▄██▀██▄▄▄▄██  ▀█▄██▄▄▄▄█▀██▄▄██▀██▄▄▄ 
 ▀▀▀▀▀   ▀▀▀▀▀▀▀▀ ▀▀ ▀▀▀▀▀▀▀   ▀▀▀▀▀▀▀▀▀  ▀▀▀▀   ▀▀▀▀ V` + slackbot.VERSION)
	bot := slackbot.NewInstance(*f_token, *f_debug, "U0JPJSUJU")
	bot.Run()

}
