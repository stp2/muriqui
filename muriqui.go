package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/glebarez/go-sqlite"
)

type Config struct {
	Token         string `json:"token"`
	Database      string `json:"database"`
	Admin         string `json:"admin"`
	NotifyChannel string `json:"notifyChannel"`
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

var adminID string
var db *sql.DB
var ds *discordgo.Session

func sendMsg(userID string, msg string) {
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

func sendAdmin(err error) {
	sendMsg(adminID, err.Error())
}
func sendChannelMsg(channelID string, msg string) {
	_, err := ds.ChannelMessageSend(channelID, msg)
	if err != nil {
		sendAdmin(err)
		log.Println("Error sending message:", err)
		return
	}
}

func sendNotification(channelID string) {
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
			return
		}
		sendMsg(schuzka.DiscordID, "Za 5 dní ("+date.Format(timeFormat)+") máš schůzku "+schuzka.Nazev+"!")
		sendChannelMsg(channelID, "<@"+schuzka.DiscordID+"> Za 5 dní ("+date.Format(timeFormat)+") má "+schuzka.Jmeno+" schůzku "+schuzka.Nazev+"!")
		_, err = db.Exec("UPDATE schuzky SET upozorneno=1 WHERE id=?", schuzka.Id)
		if err != nil {
			sendAdmin(err)
			log.Fatalln("Error updating meeting:", err)
		}
	}
}

func listMeetings(all bool) string {
	var out string
	var rows *sql.Rows
	var err error

	if all {
		rows, err = db.Query("SELECT schuzky.id, nazev,kdy,jmeno,discord_id, upozorneno FROM schuzky JOIN cleni on schuzky.cleni_id=cleni.id ORDER BY kdy ASC")
	} else {
		rows, err = db.Query("SELECT schuzky.id, nazev,kdy,jmeno,discord_id, upozorneno FROM schuzky JOIN cleni on schuzky.cleni_id=cleni.id WHERE kdy - unixepoch(datetime()) > 0 ORDER BY kdy ASC")
	}
	if err != nil {
		sendAdmin(err)
		log.Fatalln("Error getting meetings:", err)
	}
	defer rows.Close()
	for rows.Next() {
		var schuzka Schuzka
		err = rows.Scan(&schuzka.Id, &schuzka.Nazev, &schuzka.Kdy, &schuzka.Jmeno, &schuzka.DiscordID, &schuzka.Upozorneno)
		if err != nil {
			sendAdmin(err)
			log.Fatalln("Error scanning meetings:", err)
		}
		date := time.Unix(schuzka.Kdy, 0)
		out += strings.Join([]string{strconv.Itoa(schuzka.Id), schuzka.Nazev, date.Format(timeFormat), schuzka.Jmeno, schuzka.DiscordID, strconv.Itoa(schuzka.Upozorneno)}, "|")
		out += "\n"
	}
	return out
}

func listMembers() string {
	var out string
	rows, err := db.Query("SELECT * FROM cleni")
	if err != nil {
		sendAdmin(err)
		log.Fatalln("Error getting members:", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var jmeno string
		var discordID string
		err = rows.Scan(&id, &jmeno, &discordID)
		if err != nil {
			sendAdmin(err)
			log.Fatalln("Error scanning members:", err)
		}
		out += strings.Join([]string{strconv.Itoa(id), jmeno, discordID}, "|")
		out += "\n"
	}
	return out
}

func addMember(jmeno string, discordID string) bool {
	_, err := db.Exec("INSERT INTO cleni (jmeno, discord_id) VALUES (?, ?)", jmeno, discordID)
	if err != nil {
		sendAdmin(err)
		log.Println("Error adding member:", err)
		return false
	}
	return true
}

func addMeeting(nazev string, kdyS string, cleniID int) bool {
	kdy, err := time.Parse(timeFormat, kdyS)
	if err != nil {
		sendAdmin(err)
		log.Println("Error parsing date:", err)
		return false
	}
	_, err = db.Exec("INSERT INTO schuzky (nazev, kdy, cleni_id) VALUES (?, ?, ?)", nazev, kdy.Unix(), cleniID)
	if err != nil {
		sendAdmin(err)
		log.Fatalln("Error adding meeting:", err)
	}
	return true
}

func commandHandler(ds *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == ds.State.User.ID {
		return
	}
	if m.Author.ID != adminID {
		return
	}
	switch {
	case m.Content == "help":
		sendMsg(m.Author.ID, "Commands: ls, la, lm, ac, as")
		sendMsg(m.Author.ID, "ls - list meetings")
		sendMsg(m.Author.ID, "la - list all meetings")
		sendMsg(m.Author.ID, "lc - list members")
		sendMsg(m.Author.ID, "ac jmeno|discordID - add member")
		sendMsg(m.Author.ID, "as nazev|datum|clenID - add meeting")
	case m.Content == "ls":
		sendMsg(m.Author.ID, listMeetings(false))
	case m.Content == "la":
		sendMsg(m.Author.ID, listMeetings(true))
	case m.Content == "lc":
		sendMsg(m.Author.ID, listMembers())
	case strings.HasPrefix(m.Content, "ac "):
		parts := strings.Split(m.Content[3:], "|")
		if len(parts) != 2 {
			sendMsg(m.Author.ID, "Usage: ac jmeno|discordID")
			return
		}
		if addMember(parts[0], parts[1]) {
			sendMsg(m.Author.ID, "Member added")
		}
	case strings.HasPrefix(m.Content, "as "):
		parts := strings.Split(m.Content[3:], "|")
		if len(parts) != 3 {
			sendMsg(m.Author.ID, "Usage: as nazev|datum|clenID")
			return
		}
		cleniID, err := strconv.Atoi(parts[2])
		if err != nil {
			sendMsg(m.Author.ID, "Invalid member ID")
			return
		}
		if addMeeting(parts[0], parts[1], cleniID) {
			sendMsg(m.Author.ID, "Meeting added")
		}
	default:
		sendMsg(m.Author.ID, "Unknown command")
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
	adminID = conf.Admin
	// Open database
	db, err = sql.Open("sqlite", conf.Database)
	if err != nil {
		log.Fatalln("Error opening database:", err)
	}
	db.Exec("PRAGMA foreign_keys = ON;") // Enable foreign keys
	// Open Discord session
	ds, err = discordgo.New("Bot " + conf.Token)
	if err != nil {
		log.Fatalln("Error creating Discord session:", err)
	}
	// Register command handler
	ds.AddHandler(commandHandler)
	ds.Identify.Intents = discordgo.IntentsDirectMessages
	// Open discord websocket
	err = ds.Open()
	if err != nil {
		log.Fatalln("Error opening connection:", err)
	}
	defer ds.Close()

	log.Println("Bot is now running.")

	for {
		sendNotification(conf.NotifyChannel)
		time.Sleep(1 * time.Minute)
	}
}
