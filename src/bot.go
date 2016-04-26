package slackbot

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/Apollo-Community/slackbot/src/cleverbot"

	"github.com/arbovm/levenshtein"
	"github.com/nlopes/slack"
)

type Instance struct {
	Users    map[string]slack.User // TODO: not thread safe
	Channels map[string]slack.Channel
	Groups   map[string]slack.Group
	Latency  time.Duration
	Debug    bool
	BotId    string

	slack        *slack.Client
	rtm          *slack.RTM
	running      bool
	scores       map[string]map[string]int
	polls        map[string]int
	mutes        map[string]time.Time
	forum_topics []*ForumTopic
	goon         []*Quote
	apollo       []*Quote
	talking      bool
	cleverbot    *cleverbot.Session
}

func NewInstance(token string, debug bool, botid string) *Instance {
	s := slack.New(token)
	s.SetDebug(debug)

	i := &Instance{
		Users:    make(map[string]slack.User),
		Channels: make(map[string]slack.Channel),
		Groups:   make(map[string]slack.Group),
		Debug:    debug,
		BotId:    botid,
		slack:    s,
		running:  true,
		scores:   make(map[string]map[string]int),
		polls:    make(map[string]int),
		mutes:    make(map[string]time.Time),
		talking:  false,
	}

	var e error
	i.goon, e = load_quotes(GOON_QUOTES_FILE)
	if e != nil {
		fmt.Println("Warning: couldn't load goon quotes:", e)
	}
	i.apollo, e = load_quotes(APOLLO_QUOTES_FILE)
	if e != nil {
		fmt.Println("Warning: couldn't load apollo quotes:", e)
	}

	i.cleverbot, e = cleverbot.New()
	if e != nil {
		fmt.Println("Warning: couldn't initialize the cleverbot:", e)
	}

	// avoid using the same seed all the time (it defaults to 1)
	rand.Seed(time.Now().Unix())

	return i
}

func (i *Instance) Run() {
	rtm := i.slack.NewRTM()
	i.rtm = rtm
	go rtm.ManageConnection()

	for i.running {
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.ConnectedEvent:
				i.slack.SetUserPresence("auto")
				for _, u := range ev.Info.Users {
					i.Users[u.ID] = u
				}
				for _, c := range ev.Info.Channels {
					i.Channels[c.ID] = c
				}
				for _, g := range ev.Info.Groups {
					i.Groups[g.ID] = g
				}
				go i.announce_latest_forum_topics()

			case *slack.MessageEvent:
				i.HandleMsg(ev)

			case *slack.LatencyReport:
				i.Latency = ev.Value

			case *slack.RTMError:
				fmt.Printf("Error: %s\n", ev.Error())

			case *slack.InvalidAuthEvent:
				panic("Invalid credentials")
			}
		}
	}
}

func (i *Instance) ChannelMsg(channel, msg string) {
	t := i.mutes[channel]
	if time.Now().Before(t) {
		return
	}
	m := i.rtm.NewOutgoingMessage(msg, channel)
	i.rtm.SendMessage(m)
}

func (i *Instance) UserMsg(user, msg string) error {
	_, _, cid, e := i.slack.OpenIMChannel(user)
	if e != nil {
		return e
	}
	i.ChannelMsg(cid, msg)
	return nil
}

func (i *Instance) AddScore(amount int, score_type, user string) {
	tmp, ok := i.scores[score_type]
	if ok != true {
		tmp = make(map[string]int)
	}
	tmp[user] += amount
	i.scores[score_type] = tmp
	fmt.Println(i.scores)
}

func (i *Instance) GetScore(score_type, user string) int {
	return i.scores[score_type][user]
}

func (i *Instance) StartVote(channel string) {
	i.polls[channel] = 0
}

func (i *Instance) StopVote(channel string) int {
	val := i.polls[channel]
	delete(i.polls, channel)
	return val

}

func (i *Instance) Vote(val int, channel string) {
	i.polls[channel] += val
}

func (i *Instance) Mute(channel string, dur time.Duration) {
	i.mutes[channel] = time.Now().Add(dur)
}

type Message struct {
	User      string
	Channel   string
	Timestamp time.Time
	Message   string
}

func (i *Instance) HandleMsg(msg *slack.MessageEvent) {
	if msg.User == i.BotId {
		// Sometimes we get a message from ourselves..
		return
	}

	sec, e := strconv.ParseInt(strings.Split(msg.Timestamp, ".")[0], 10, 64)
	if e != nil {
	}
	nsec, e := strconv.ParseInt(strings.Split(msg.Timestamp, ".")[1], 10, 64)
	if e != nil {
	}

	m := &Message{
		User:      msg.User,
		Channel:   msg.Channel,
		Timestamp: time.Unix(sec, nsec),
		Message:   msg.Text,
	}

	mention := fmt.Sprintf("<@%s>", i.BotId)
	if strings.HasPrefix(m.Message, mention) {
		m.Message = strings.TrimPrefix(m.Message, mention)
		m.Message = strings.TrimSpace(strings.TrimPrefix(m.Message, ":"))
		i.parse_msg(m)
	}
}

func (i *Instance) parse_msg(m *Message) {
	e := i.parse_cmd(m)
	if e != nil {
		i.UserMsg(m.User, e.Error())
	}
}

func (i *Instance) parse_cmd(m *Message) error {
	tokens := strings.Fields(m.Message)
	if len(tokens) < 1 {
		return nil
	}
	cmd := strings.ToLower(tokens[0])

	var bestscore float64
	var bestcmd *Command
	for _, c := range COMMANDS {
		score := calcScore(cmd, c.Name)
		if bestcmd == nil || score > bestscore {
			bestscore = score
			bestcmd = c
		}
	}
	if bestscore >= 0.75 { // 50% match is good enough!
		return bestcmd.Func(i, m, tokens[1:])
	}

	if m.Channel == TALK_CHANNEL && i.talking {
		resp, e := i.cleverbot.Ask(m.Message)
		if e != nil {
			fmt.Println("cleverbot error:", e)
			return nil
		}
		i.ChannelMsg(TALK_CHANNEL, resp)
		return nil
	}

	log.Printf("%v: %v\n", i.Users[m.User].Name, m.Message)
	return fmt.Errorf("Did you mean `%s`?", bestcmd.Name)
}

func calcScore(a, b string) float64 {
	score := levenshtein.Distance(a, b)
	lena := len(a)
	lenb := len(b)
	var bigger float64
	if lena > lenb {
		bigger = float64(lena)
	} else {
		bigger = float64(lenb)
	}
	return (bigger - float64(score)) / bigger
}
