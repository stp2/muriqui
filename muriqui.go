package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/glebarez/go-sqlite"
)

type Config struct {
	Token    string `json:"token"`
	Database string `json:"database"`
	Admin    string `json:"admin"`
}

type Schuzka struct {
	Id         int
	Nazev      string
	Kdy        int64
	Jmeno      string
	DiscordID  string
	Upozorneno int
}

const (
	timeFormat = "2.1. 2006 15:04"
)

var sendAdmin func(error)

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

func sendNotification(ds *discordgo.Session, db *sql.DB) {
	var schuzka Schuzka
	row := db.QueryRow("SELECT schuzky.id, nazev,kdy,jmeno,discord_id, upozorneno FROM schuzky JOIN cleni on schuzky.cleni_id=cleni.id WHERE kdy - unixepoch(datetime()) > 0 ORDER BY kdy ASC LIMIT 1")
	err := row.Scan(&schuzka.Id, &schuzka.Nazev, &schuzka.Kdy, &schuzka.Jmeno, &schuzka.DiscordID, &schuzka.Upozorneno)
	if err != nil {
		sendAdmin(err)
		log.Fatalln("Error getting next meeting:", err)
	}
	date := time.Unix(schuzka.Kdy, 0)
	if schuzka.Upozorneno == 0 {
		if time.Until(date) > 5*24*time.Hour {
			<-time.After(time.Until(date) - 5*24*time.Hour)
		}
		sendMsg(ds, schuzka.DiscordID, "Za 5 dní ("+date.Format(timeFormat)+") máš schůzku "+schuzka.Nazev+"!")
		_, err = db.Exec("UPDATE schuzky SET upozorneno=1 WHERE id=?", schuzka.Id)
		if err != nil {
			sendAdmin(err)
			log.Fatalln("Error updating meeting:", err)
		}
	}
	<-time.After(time.Until(date) + time.Hour)
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
	// Open database
	db, err := sql.Open("sqlite", conf.Database)
	if err != nil {
		log.Fatalln("Error opening database:", err)
	}
	db.Exec("PRAGMA foreign_keys = ON;") // Enable foreign keys
	// Open Discord session
	ds, err := discordgo.New("Bot " + conf.Token)
	if err != nil {
		log.Fatalln("Error creating Discord session:", err)
	}
	// Open discord websocket
	err = ds.Open()
	if err != nil {
		log.Fatalln("Error opening connection:", err)
	}
	defer ds.Close()
	sendAdmin = func(err error) {
		sendMsg(ds, conf.Admin, err.Error())
	}

	for {
		sendNotification(ds, db)
	}
}
