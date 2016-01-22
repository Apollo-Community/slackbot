package main

import (
	"fmt"

	"github.com/Apollo-Community/slackbot/src"
)

func main() {
	fmt.Println(`
  ▄▄▄▄▄▄▄▄                ▄▄     ▄▄▄▄▄▄                
▄█▀▀▀▀▀▀██                ██     ██▀▀▀▀██       ██    
██▄     ██   ▄█████▄▄███████ ▄██▀██    ██▄███████████ 
 ▀████▄ ██   ▀ ▄▄▄███▀    ██▄██  █████████▀  ▀████    
     ▀████  ▄██▀▀▀███     ██▀██▄ ██    ███    ████    
█▄▄▄▄▄█▀██▄▄██▄▄▄██▀██▄▄▄▄██  ▀█▄██▄▄▄▄█▀██▄▄██▀██▄▄▄ 
 ▀▀▀▀▀   ▀▀▀▀▀▀▀▀ ▀▀ ▀▀▀▀▀▀▀   ▀▀▀▀▀▀▀▀▀  ▀▀▀▀   ▀▀▀▀ V` + slackbot.VERSION)
	bot := slackbot.NewInstance(false, "U0JPJSUJU")
	bot.Run()

}
