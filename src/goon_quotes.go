package slackbot

import (
	"bufio"
	"math/rand"
	"os"
	"regexp"
	"strings"
)

var r_quote = regexp.MustCompile(`([^:]*):.*\/\/(.*)`)

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
	tmp := r_quote.FindStringSubmatch(line)
	if tmp == nil || len(tmp) != 3 {
		return nil
	}

	file := strings.TrimSpace(tmp[1])
	quote := strings.TrimSpace(tmp[2])
	if len(file) < 1 || len(quote) < 1 {
		return nil
	}

	return &GoonQuote{file, quote}
}

func (i *Instance) random_goon_quote() *GoonQuote {
	line := rand.Intn(len(i.goon))
	return i.goon[line]
}
