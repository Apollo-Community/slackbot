package slackbot

import (
	"bufio"
	"math/rand"
	"os"
	"strings"
)

type GoonQuote struct {
	File  string
	Quote string
}

func load_goon_quotes() ([]*GoonQuote, error) {
	f, e := os.Open(GOON_QUOTES_FILE)
	if e != nil {
		return nil, e
	}
	defer f.Close()

	var quotes []*GoonQuote
	s := bufio.NewScanner(f)
	for s.Scan() {
		q := parse_goon_quote(s.Text())
		if q != nil {
			quotes = append(quotes, q)
		}
	}

	if e = s.Err(); e != nil {
		return nil, e
	}

	return quotes, nil
}
func parse_goon_quote(line string) *GoonQuote {
	tmp := strings.SplitN(line, " ", 2)
	if len(tmp) != 2 {
		return nil
	}
	quote := strings.TrimSpace(tmp[1])
	file := strings.TrimSpace(tmp[0])
	fparts := strings.SplitN(file, "goonstation", 2)
	file = fparts[1]

	return &GoonQuote{file, quote}
}
func (i *Instance) random_goon_quote() *GoonQuote {
	line := rand.Intn(len(i.goon))
	return i.goon[line]
}
