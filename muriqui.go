package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Config struct {
	Token string `json:"token"`
}

func sendMsg(ds *discordgo.Session, userID string, msg string) {
	channel, err := ds.UserChannelCreate(userID)
	if err != nil {
		fmt.Println("Error creating channel:", err)
		return
	}
	_, err = ds.ChannelMessageSend(channel.ID, msg)
	if err != nil {
		fmt.Println("Error sending message:", err)
		return
	}
}

func main() {
	confB, err := os.ReadFile("config.json")
	if err != nil {
		fmt.Println("Error reading config file:", err)
		os.Exit(1)
	}
	var conf Config
	err = json.Unmarshal(confB, &conf)
	if err != nil {
		fmt.Println("Error parsing config file:", err)
		os.Exit(1)
	}
	token := conf.Token

	ds, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session:", err)
		os.Exit(1)
	}

	err = ds.Open()
	if err != nil {
		fmt.Println("Error opening connection:", err)
		os.Exit(1)
	}

	<-time.After(30 * time.Second)
	sendMsg(ds, "0", "Goodbye, world!") // replace "0" with your user ID

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	ds.Close()
}
