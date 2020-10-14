package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

type CommandHandler func(*discordgo.Session, *discordgo.MessageCreate)

var Commands = map[string]CommandHandler{
	"ping": ping,
	"pong": pong,
}

func main() {

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		log.Panic("error creating Discord session,", err)
	}

	// Register the distributor func as a callback for MessageCreate events.
	dg.AddHandler(distributor)

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		log.Panic("error opening connection,", err)
	}

	log.Println("discord-accountant is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	log.Println("Closing connection")
	err = dg.Close()
	if err != nil {
		log.Panic("error closing connection,", err)
	}
}

func ping(s *discordgo.Session, m *discordgo.MessageCreate) {
	_, err := s.ChannelMessageSend(m.ChannelID, "Pong!")
	if err != nil {
		log.Print("error sending pong,", err)
	}
}

func pong(s *discordgo.Session, m *discordgo.MessageCreate) {
	channel, err := getPrivateChannel(s, m)
	if err != nil {
		log.Println("error getting private channel,", err)
	}
	_, err = s.ChannelMessageSend(channel.ID, "Ping!")
	if err != nil {
		log.Println("error sending ping,", err)
	}
}

func getPrivateChannel(s *discordgo.Session, m *discordgo.MessageCreate) (*discordgo.Channel, error) {
	channel, err := s.State.Channel(m.ChannelID)
	if err != nil || channel.Type != discordgo.ChannelTypeDM {
		channel, err = s.UserChannelCreate(m.Author.ID)
	}
	return channel, err
}

func distributor(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if commandHandler, ok := Commands[m.Content]; ok {
		commandHandler(s, m)
	}
}
