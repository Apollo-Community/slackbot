package slackbot

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const ANNOUNCE_TIMEOUT = 30 // minutes

type Forum struct {
	Title     string
	URL       string
	ChannelID string
}

type ForumTopic struct {
	Title string
	URL   string
}

var FORUMS = []*Forum{
	&Forum{"Staff Applications", "https://apollo-community.org/viewforum.php?f=15", "G0JPDL06A"},
	&Forum{"Head Applications", "https://apollo-community.org/viewforum.php?f=16", "G0JPDL06A"},
	&Forum{"Race Applications", "https://apollo-community.org/viewforum.php?f=17", "G0JPDL06A"},
	&Forum{"Ban Requests", "https://apollo-community.org/viewforum.php?f=33", "G0JPDL06A"},
	&Forum{"Unban Requests", "https://apollo-community.org/viewforum.php?f=34", "G0JPDL06A"},
}

func (i *Instance) get_latest_forum_topics(forum_url string, only_new bool) ([]*ForumTopic, error) {
	doc, e := goquery.NewDocument(forum_url)
	if e != nil {
		return nil, fmt.Errorf("Couldn't load the forum page.")
	}

	var items []*ForumTopic
	doc.Find(".topics a.topictitle").Each(func(_ int, s *goquery.Selection) {
		title := s.Text()
		url := strings.TrimPrefix(s.AttrOr("href", ""), "./")
		url = "https://apollo-community.org/" + strings.Split(url, "&sid=")[0]

		t := &ForumTopic{title, url}
		if only_new {
			for _, ft := range i.forum_topics {
				if t.URL == ft.URL {
					// Ignore appeals already seen
					return
				}
			}
			// Else make sure we remember it for the future
			i.forum_topics = append(i.forum_topics, t)
		}
		items = append(items, t)
	})
	return items, nil
}

func (i *Instance) announce_latest_forum_topics() {
	// make sure we have seen the latest topics
	for _, f := range FORUMS {
		i.get_latest_forum_topics(f.URL, true)
	}

	// We need a real ticker, because go threads won't sleep properly with time.Sleep
	t := time.NewTicker(time.Duration(ANNOUNCE_TIMEOUT) * time.Minute)
	defer t.Stop()

	for {
		<-t.C // sleep for now

		for _, f := range FORUMS {
			i.try_forum_announce(f)
		}
	}
}

func (i *Instance) try_forum_announce(f *Forum) {
	items, e := i.get_latest_forum_topics(f.URL, true)
	if e != nil {
		log.Printf("Failed to get latest topics for %v: %v\n", f.Title, e)
	} else if len(items) > 0 {
		tmp := fmt.Sprintf("New topics on *%v*:\n>>>", f.Title)
		for _, t := range items {
			tmp += fmt.Sprintf("%v `%v`\n", t.Title, t.URL)
		}
		i.ChannelMsg(f.ChannelID, tmp)
	}
}
