package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
	_ "github.com/glebarez/go-sqlite"
)

type Config struct {
	Token    string `json:"token"`
	Database string `json:"database"`
}

func sendMsg(ds *discordgo.Session, userID string, msg string) {
	channel, err := ds.UserChannelCreate(userID)
	if err != nil {
		log.Println("Error creating channel:", err)
		return
	}
	_, err = ds.ChannelMessageSend(channel.ID, msg)
	if err != nil {
		log.Println("Error sending message:", err)
		return
	}
}

func main() {
	// Load configuration
	confB, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatalln("Error reading config file:", err)
	}
	var conf Config
	err = json.Unmarshal(confB, &conf)
	if err != nil {
		log.Fatalln("Error parsing config file:", err)
	}
	token := conf.Token
	dbFile := conf.Database
	// Open database
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		log.Fatalln("Error opening database:", err)
	}
	// Open Discord session
	ds, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalln("Error creating Discord session:", err)
	}
	// Open discord websocket
	err = ds.Open()
	if err != nil {
		log.Fatalln("Error opening connection:", err)
	}
	defer ds.Close()

	rows, err := db.Query("SELECT discord_id FROM cleni WHERE jmeno = 'Prokop'")
	if err != nil {
		log.Fatalln("Error querying database:", err)
	}
	defer rows.Close()
	rows.Next()
	var id string
	err = rows.Scan(&id)
	if err != nil {
		log.Fatalln("Error scanning row:", err)
	}
	sendMsg(ds, id, "Hello, world!")
}
