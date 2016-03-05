package slackbot

import (
	"bufio"
	"math/rand"
	"os"
	"strings"
)

type Quote struct {
	File  string
	Quote string
}

func load_quotes(file string) ([]*Quote, error) {
	f, e := os.Open(file)
	if e != nil {
		return nil, e
	}
	defer f.Close()

	var quotes []*Quote
	s := bufio.NewScanner(f)
	for s.Scan() {
		q := parse_quote(s.Text())
		if q != nil {
			quotes = append(quotes, q)
		}
	}

	if e = s.Err(); e != nil {
		return nil, e
	}

	return quotes, nil
}
func parse_quote(line string) *Quote {
	tmp := strings.SplitN(line, " ", 2)
	if len(tmp) != 2 {
		return nil
	}
	quote := strings.TrimSpace(tmp[1])
	file := strings.TrimSpace(tmp[0])
	fparts := strings.SplitN(file, "code", 2)
	file = "/code" + fparts[1]

	return &Quote{file, quote}
}
func (i *Instance) random_goon_quote() *Quote {
	line := rand.Intn(len(i.goon))
	return i.goon[line]
}
