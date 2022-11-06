package main

import (
	"fmt"
	"github.com/alexflint/go-arg"
	"github.com/bwmarrin/discordgo"
	"github.com/mmcdole/gofeed"
	"github.com/robfig/cron/v3"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

const DiscordMessage = `
Title : %s
Lien : %s
Categories : %s
Date de publication = %s
`

var AppStart time.Time

type StockEvent struct {
	Url        string
	Categories []string
	Publish    time.Time
	Title      string
}

type argsType struct {
	Token   string `arg:"required"`
	CronRss string `default:"0 0/1 * * * *"`
}

func init() {
	AppStart = time.Now()
}

func RssFeed(c chan StockEvent, cronRss string) {
	fp := gofeed.NewParser()
	cr := cron.New(cron.WithSeconds())

	lastUpdate := time.Now().Add(time.Hour * 24 * -4)

	var lastUpdateMutex sync.RWMutex

	cr.AddFunc(cronRss, func() {
		feed, _ := fp.ParseURL("https://rpilocator.com/feed/?cat=PI4")

		lastUpdateMutex.Lock()
		defer lastUpdateMutex.Unlock()

		if feed.UpdatedParsed.After(lastUpdate) {
			for _, item := range feed.Items {
				if *item.PublishedParsed == lastUpdate || item.PublishedParsed.After(lastUpdate) {
					stockEvent := StockEvent{
						Categories: item.Categories,
						Publish:    *item.PublishedParsed,
						Title:      item.Title,
						Url:        item.Link,
					}
					c <- stockEvent
				}
			}
			lastUpdate = *feed.UpdatedParsed
		}

	})

	cr.Run()
}

func main() {

	var args argsType

	arg.MustParse(&args)

	chanPiStock := make(chan StockEvent)

	go RssFeed(chanPiStock, args.CronRss)

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + args.Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	dg.AddHandler(onAddServer)

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuilds | discordgo.IntentGuilds

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()

	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	go sendAll(dg, chanPiStock)

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()

}

func sendAll(session *discordgo.Session, c chan StockEvent) {
	for str := range c {
		for _, guild := range session.State.Guilds {
			message := fmt.Sprintf(
				DiscordMessage,
				str.Title,
				str.Url,
				strings.Join(str.Categories, ", "),
				str.Publish.Format(time.RFC822),
			)
			session.ChannelMessageSend(guild.SystemChannelID, message)
		}
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	// If the message is "ping" reply with "Pong!"
	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}

	// If the message is "pong" reply with "Ping!"
	if m.Content == "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}
}

func onAddServer(s *discordgo.Session, guild *discordgo.GuildCreate) {
	if guild.JoinedAt.After(AppStart) {
		s.ChannelMessageSend(guild.SystemChannelID, "Welcome !!!!")
	}
}
