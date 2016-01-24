package slackbot

import (
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/justinian/dice"
)

type Command struct {
	Name string
	Desc string
	Func func(*Instance, *Message, string) error
}

var COMMANDS []*Command

func init() {
	COMMANDS = []*Command{
		&Command{"help", "Show a list of commands.", cmd_help},
		&Command{"status", "Show my current status.", cmd_status},
		&Command{"say", "Repeat a message in the current channel.", cmd_say},
		&Command{"duck", "Start a duck hunt.", cmd_duck},
		&Command{"shoot", "Shoot a duck.", cmd_shoot_duck},
		&Command{"catch", "Catch a duck.", cmd_catch_duck},
		&Command{"vote", "start/stop a vote or vote yes/no during a vote.", cmd_vote},
		&Command{"mute", "Mute my messages to this channel, for a while.", cmd_mute},
		&Command{"roll", "Throw a dice roll.", cmd_roll},
		&Command{"wiki", "Quote a page from our SS13 wiki.", cmd_wiki},
		&Command{"pun", "Tell a random pun.", cmd_pun},
		&Command{"badpun", "Tell a random, bad pun.", cmd_pun},
		&Command{"catfact", "Tell a random cat fact.", cmd_catfact},
	}
}

func cmd_help(i *Instance, m *Message, args string) error {
	// TODO: sort the list
	tmp := "Available commands:\n"
	for _, c := range COMMANDS {
		tmp += fmt.Sprintf("- `%s\t%s`\n", c.Name, c.Desc)
	}
	i.UserMsg(m.User, tmp)
	return nil
}

func cmd_say(i *Instance, m *Message, args string) error {
	if len(args) > 1 {
		i.ChannelMsg(m.Channel, args)
	} else {
		return fmt.Errorf("Too short message for `say`.")
	}
	return nil
}

func cmd_duck(i *Instance, m *Message, args string) error {
	i.ModeOn("duck_" + m.Channel)
	i.ChannelMsg(m.Channel, "Oh look, a random *duck* appeard! Try to `shoot` or `catch` it!")
	return nil
}

func cmd_shoot_duck(i *Instance, m *Message, args string) error {
	if i.ModeStatus("duck_" + m.Channel) {
		i.ModeOff("duck_" + m.Channel)
		i.AddScore(1, "duck", m.User)
		tmp := fmt.Sprintf("Pew pew! *%s* has shot the duck!\n*%s* has now shot %d ducks.", i.Users[m.User].Name, i.Users[m.User].Name, i.GetScore("duck", m.User))
		i.ChannelMsg(m.Channel, tmp)
		return nil
	}
	return fmt.Errorf("What are you shooting at, cowboy?")
}

func cmd_catch_duck(i *Instance, m *Message, args string) error {
	if i.ModeStatus("duck_" + m.Channel) {
		i.ModeOff("duck_" + m.Channel)
		tmp := fmt.Sprintf("Yoink! *%s* has caught the duck but let it go after a while...", i.Users[m.User].Name)
		i.ChannelMsg(m.Channel, tmp)
		return nil
	}
	return fmt.Errorf("Nuh uh, can't touch me.")
}

func cmd_vote(i *Instance, m *Message, args string) error {
	if args == "start" {
		i.StartVote(m.Channel)
		i.ChannelMsg(m.Channel, fmt.Sprintf("@%s has started a new vote! Tell me if you would like to `vote yes` or `vote no` on it.", i.Users[m.User].Name))
	} else if args == "stop" {
		votes := i.StopVote(m.Channel)
		var result string
		if votes > 0 {
			result = "*YES* has won"
		} else if votes < 0 {
			result = "*NO* has won"
		} else {
			result = "It's a tie! No one won"
		}
		i.ChannelMsg(m.Channel, fmt.Sprintf("@%s has stopped the vote! The result is...\n%s (score: %d)!", i.Users[m.User].Name, result, votes))
	} else if args == "yes" {
		i.Vote(+1, m.Channel)
	} else if args == "no" {
		i.Vote(-1, m.Channel)
	}
	return nil
}

func cmd_mute(i *Instance, m *Message, args string) error {
	dur, e := time.ParseDuration(args)
	if e != nil {
		return fmt.Errorf("Couldn't parse a duration.")
	}
	if dur.Minutes() < 1 {
		dur = time.Duration(5) * time.Minute
	}
	if dur.Minutes() > 60 {
		dur = time.Duration(60) * time.Minute
	}

	i.ChannelMsg(m.Channel, fmt.Sprintf("I will now shut up for ~%.0f minutes.", dur.Minutes()))
	go func() {
		time.Sleep(5)
		i.Mute(m.Channel, dur)
	}()
	return nil
}

func cmd_status(i *Instance, m *Message, args string) error {
	tmp := "My current status:\n"
	tmp += fmt.Sprintf("`Running: %v`\n", i.running)
	tmp += fmt.Sprintf("`Version: v%v`\n", VERSION)
	tmp += fmt.Sprintf("`Debug mode: %v`\n", i.Debug)
	tmp += fmt.Sprintf("`Latency: %v`\n", i.Latency)
	tmp += fmt.Sprintf("`My User ID: @%v`\n", i.BotId)
	tmp += fmt.Sprintf("`Tracked channels: %v`\n", len(i.Channels))
	tmp += fmt.Sprintf("`Tracked users: %v`\n", len(i.Users))
	tmp += fmt.Sprintf("\nI am currently muted on channels:\n")
	if len(i.mutes) < 1 {
		tmp += "- `None`\n"
	} else {
		for k, v := range i.mutes {
			c := i.Channels[k].Name
			t := v.Sub(time.Now())
			tmp += fmt.Sprintf("- `%v (ends in ~%.0f minutes)`\n", c, t.Minutes())
		}
	}
	i.UserMsg(m.User, tmp)
	return nil
}

func cmd_roll(i *Instance, m *Message, args string) error {
	ret, res, e := dice.Roll(args)
	if e != nil {
		return fmt.Errorf("Bad dice format.")
	}
	i.ChannelMsg(m.Channel, fmt.Sprintf("@%s has rolled: %v, %v\n", i.Users[m.User].Name, ret, res))
	return nil
}

func cmd_wiki(i *Instance, m *Message, args string) error {
	u := WIKI_URL + url.QueryEscape(args)
	doc, e := goquery.NewDocument(u)
	if e != nil {
		return fmt.Errorf("Couldn't open the wiki for you (%v).", u)
	}

	node := doc.Find("div #mw-content-text > p").First()
	text := strings.TrimSpace(node.Text())
	if node.Length() < 1 || len(text) < 1 {
		return fmt.Errorf("Couldn't quote that page for you (%v).", u)
	}

	i.ChannelMsg(m.Channel, fmt.Sprintf(">>>%s\n`Source: %s`", text, u))
	return nil
}

func cmd_pun(i *Instance, m *Message, args string) error {
	resp, e := http.Get("http://www.punoftheday.com/cgi-bin/arandompun.pl")
	if e != nil {
		return fmt.Errorf("Sorry, couldn't make up a pun for you.")
	}
	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return fmt.Errorf("Sorry, couldn't make up a pun for you.")
	}
	// Yep this is a fucking mess.
	s := strings.TrimSpace(string(body))
	s = strings.TrimPrefix(s, "document.write('&quot;")
	index := strings.Index(s, "&quot;")
	s = html.UnescapeString(s[:index])
	i.ChannelMsg(m.Channel, fmt.Sprintf(">>>%v\n`© 1996-2011 PunoftheDay.com`", s))
	return nil
}

type CatFacts struct {
	Facts   []string
	Success string
}

func cmd_catfact(i *Instance, m *Message, args string) error {
	resp, e := http.Get("http://catfacts-api.appspot.com/api/facts?number=1")
	if e != nil {
		return fmt.Errorf("Sorry, couldn't find any cat facts.")
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var cf CatFacts
	e = decoder.Decode(&cf)
	if e != nil {
		return fmt.Errorf("Sorry, couldn't find any cat facts.")
	}
	if len(cf.Facts) > 0 {
		msg := fmt.Sprintf(">>>%v\n`© http://catfacts-api.appspot.com`", cf.Facts[0])
		i.ChannelMsg(m.Channel, msg)
		return nil
	}
	return fmt.Errorf("Sorry, couldn't find any cat facts.")
}
