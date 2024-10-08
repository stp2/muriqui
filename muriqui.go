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
	//_ "github.com/glebarez/go-sqlite"
	_ "github.com/mattn/go-sqlite3"
)

type Config struct {
	Token         string `json:"token"`
	Database      string `json:"database"`
	Admin         string `json:"admin"`
	NotifyChannel string `json:"notifyChannel"`
}

type Schuzka struct {
	Id         int64
	Nazev      string
	Kdy        int64
	Upozorneno bool
}

type Poradajici struct {
	ID        int64
	Jmeno     string
	DiscordID string
	ZpravaID  string
}

type Meeting struct {
	ID         int64
	Nazev      string
	Kdy        int64
	Kdo        string
	Upozorneno bool
	ZpravaID   string
}

type Member struct {
	ID        int64
	Jmeno     string
	DiscordID string
}

const (
	timeFormat = "2.1. 2006 15:04"
	notifyTime = 5 * 24 * 3600
)

func sendMsg(ds *discordgo.Session, userID string, msg string) string {
	channel, err := ds.UserChannelCreate(userID)
	if err != nil {
		log.Println("Error creating channel:", err)
		return ""
	}
	m, err := ds.ChannelMessageSend(channel.ID, msg)
	if err != nil {
		log.Println("Error sending message:", err)
		return ""
	}
	return m.ID
}

var sendAdmin func(*discordgo.Session, error) // closure in main

func sendChannelMsg(ds *discordgo.Session, channelID string, msg string) string {
	m, err := ds.ChannelMessageSend(channelID, msg)
	if err != nil {
		sendAdmin(ds, err)
		log.Println("Error sending message:", err)
		return ""
	}
	return m.ID
}

func reacted(ds *discordgo.Session, mID string, uID string) bool {
	if mID == "" {
		return false
	}
	channel, err := ds.UserChannelCreate(uID)
	if err != nil {
		log.Println("Error creating channel:", err)
		return false
	}
	msg, err := ds.ChannelMessage(channel.ID, mID)
	if err != nil {
		log.Println("Error getting message:", err)
		return false
	}
	if len(msg.Reactions) == 0 {
		return false
	}
	return true
}

func sendNotification(ds *discordgo.Session, db *sql.DB, channelID string) {
	schuzky := make([]Schuzka, 0)
	row, err := db.Query(
		`SELECT id, nazev, kdy, upozorneno
        FROM schuzky
        WHERE kdy - unixepoch(datetime()) >= 0 AND kdy - unixepoch(datetime()) <= ?;`,
		notifyTime)
	if err != nil {
		sendAdmin(ds, err)
		log.Fatalln("Error getting next meeting:", err)
	}
	for row.Next() {
		var schuzka Schuzka
		row.Scan(&schuzka.Id, &schuzka.Nazev, &schuzka.Kdy, &schuzka.Upozorneno)
		schuzky = append(schuzky, schuzka)
	}
	row.Close()
	for _, schuzka := range schuzky {
		date := time.Unix(schuzka.Kdy, 0)
		cleni := make([]Poradajici, 0)
		cleniRow, err := db.Query(
			`SELECT cleni.id, jmeno, discord_id, zprava_id
            FROM cleni
            JOIN porada ON cleni.id = porada.cleni_id
            WHERE porada.schuzky_id = ?;`,
			schuzka.Id)
		if err != nil {
			sendAdmin(ds, err)
			log.Fatalln("Error getting members:", err)
		}
		for cleniRow.Next() {
			var clen Poradajici
			cleniRow.Scan(&clen.ID, &clen.Jmeno, &clen.DiscordID, &clen.ZpravaID)
			cleni = append(cleni, clen)
		}
		cleniRow.Close()
		if !schuzka.Upozorneno {
			mentions := ""
			for _, clen := range cleni {
				mentions += "<@" + clen.DiscordID + "> "
			}
			if len(cleni) == 1 {
				sendChannelMsg(ds, channelID, mentions+date.Format(timeFormat)+" má schůzku "+schuzka.Nazev+"!")
			} else if len(cleni) > 1 {
				sendChannelMsg(ds, channelID, mentions+date.Format(timeFormat)+" mají schůzku "+schuzka.Nazev+"!")
			} else {
				sendMsg(ds, channelID, date.Format(timeFormat)+") je schůzka "+schuzka.Nazev+"!")
			}
			_, err = db.Exec(`UPDATE schuzky
            SET upozorneno=1
            WHERE id=?`, schuzka.Id)
			if err != nil {
				sendAdmin(ds, err)
				log.Fatalln("Error updating meeting:", err)
			}
		}
		for _, clen := range cleni {
			if !reacted(ds, clen.ZpravaID, clen.DiscordID) {
				mID := sendMsg(ds, clen.DiscordID, date.Format(timeFormat)+" máš schůzku "+schuzka.Nazev+"!\nReaguj na tuto zprávu, pokud jsi upozorněn.\nNa tu poslední.")
				_, err = db.Exec(`UPDATE porada
                SET zprava_id=?
                WHERE cleni_id=? AND schuzky_id=?`,
					mID, clen.ID, schuzka.Id)
				if err != nil {
					sendAdmin(ds, err)
					log.Fatalln("Error updating meeting:", err)
				}
			}
		}
	}
}

func listMeetings(ds *discordgo.Session, db *sql.DB, all bool) string {
	var out string
	var rows *sql.Rows
	var err error
	var notify string

	if all {
		rows, err = db.Query(`SELECT schuzky.id, nazev,kdy,jmeno,upozorneno,zprava_id
        FROM schuzky
        JOIN porada ON schuzky.id = porada.schuzky_id
        JOIN cleni ON porada.cleni_id = cleni.id
        ORDER BY kdy ASC;`)
	} else {
		rows, err = db.Query(`SELECT schuzky.id, nazev,kdy,jmeno,upozorneno,zprava_id
        FROM schuzky
        JOIN porada ON schuzky.id = porada.schuzky_id
        JOIN cleni ON porada.cleni_id = cleni.id
        WHERE kdy - unixepoch(datetime()) >= 0
        ORDER BY kdy ASC;`)
	}
	if err != nil {
		sendAdmin(ds, err)
		log.Fatalln("Error getting meetings:", err)
	}
	defer rows.Close()
	for rows.Next() {
		var schuzka Meeting
		err = rows.Scan(&schuzka.ID, &schuzka.Nazev, &schuzka.Kdy, &schuzka.Kdo, &schuzka.Upozorneno, &schuzka.ZpravaID)
		if err != nil {
			sendAdmin(ds, err)
			log.Fatalln("Error scanning meetings:", err)
		}
		date := time.Unix(schuzka.Kdy, 0)
		if schuzka.Upozorneno {
			notify = "ano"
		} else {
			notify = "ne"
		}
		out += strings.Join([]string{strconv.Itoa(int(schuzka.ID)), schuzka.Nazev, date.Format(timeFormat), schuzka.Kdo, notify, schuzka.ZpravaID}, "|")
		out += "\n"
	}
	if out == "" {
		out = "No meetings found"
	}
	return out
}

func listMembers(ds *discordgo.Session, db *sql.DB) string {
	var out string
	rows, err := db.Query("SELECT id,jmeno,discord_id FROM cleni")
	if err != nil {
		sendAdmin(ds, err)
		log.Fatalln("Error getting members:", err)
	}
	defer rows.Close()
	for rows.Next() {
		var member Member
		err = rows.Scan(&member.ID, &member.Jmeno, &member.DiscordID)
		if err != nil {
			sendAdmin(ds, err)
			log.Fatalln("Error scanning members:", err)
		}
		out += strings.Join([]string{strconv.Itoa(int(member.ID)), member.Jmeno, member.DiscordID}, "|")
		out += "\n"
	}
	if out == "" {
		out = "No members found"
	}
	return out
}

func addMember(ds *discordgo.Session, db *sql.DB, jmeno string, discordID string) bool {
	_, err := db.Exec("INSERT INTO cleni (jmeno, discord_id) VALUES (?, ?)", jmeno, discordID)
	if err != nil {
		sendAdmin(ds, err)
		log.Println("Error adding member:", err)
		return false
	}
	return true
}

func addMeeting(ds *discordgo.Session, db *sql.DB, nazev string, kdyS string, cleniID []int) bool {
	loc, err := time.LoadLocation("Europe/Prague")
	if err != nil {
		sendAdmin(ds, err)
		log.Println("Error loading location:", err)
		return false
	}
	kdy, err := time.ParseInLocation(timeFormat, kdyS, loc)
	if err != nil {
		sendAdmin(ds, err)
		log.Println("Error parsing date:", err)
		return false
	}
	_, err = db.Exec("INSERT INTO schuzky (nazev, kdy) VALUES (?, ?)", nazev, kdy.Unix())
	if err != nil {
		sendAdmin(ds, err)
		log.Println("Error adding member to meeting:", err)
		return false
	}
	var lastID int
	lastIDR := db.QueryRow("SELECT last_insert_rowid()")
	err = lastIDR.Scan(&lastID)
	if err != nil {
		sendAdmin(ds, err)
		log.Println("Error getting last ID:", err)
		return false
	}
	for _, id := range cleniID {
		_, err = db.Exec("INSERT INTO porada (cleni_id, schuzky_id) VALUES (?, ?)", id, lastID)
		if err != nil {
			sendAdmin(ds, err)
			log.Println("Error adding member to meeting:", err)
			return false
		}
	}
	return true
}

func removeMeeting(ds *discordgo.Session, db *sql.DB, id int) bool {
	_, err := db.Exec("DELETE FROM porada WHERE schuzky_id=?", id)
	if err != nil {
		sendAdmin(ds, err)
		log.Println("Error removing meeting:", err)
		return false
	}
	_, err = db.Exec("DELETE FROM schuzky WHERE id=?", id)
	if err != nil {
		sendAdmin(ds, err)
		log.Println("Error removing meeting:", err)
		return false
	}
	return true
}

func removeMember(ds *discordgo.Session, db *sql.DB, id int) bool {
	_, err := db.Exec("DELETE FROM cleni WHERE id=?", id)
	if err != nil {
		sendAdmin(ds, err)
		log.Println("Error removing member:", err)
		return false
	}
	return true
}

func commandHandler(ds *discordgo.Session, m *discordgo.MessageCreate, db *sql.DB, adminID string) {
	if m.Author.ID == ds.State.User.ID {
		return
	}
	if m.Author.ID != adminID {
		return
	}
	switch {
	case m.Content == "help":
		msg := "Commands: ls, la, lm, ac, as, rm, rc\n"
		msg += "ls - list meetings\n"
		msg += "la - list all meetings\n"
		msg += "lc - list members\n"
		msg += "ac jmeno|discordID - add member\n"
		msg += "as nazev|datum|clenID(odděleno čárkou) - add meeting\n"
		msg += "rm id - remove meeting\n"
		msg += "rc id - remove member\n"
		sendMsg(ds, m.Author.ID, msg)
	case m.Content == "ls":
		sendMsg(ds, m.Author.ID, listMeetings(ds, db, false))
	case m.Content == "la":
		sendMsg(ds, m.Author.ID, listMeetings(ds, db, true))
	case m.Content == "lc":
		sendMsg(ds, m.Author.ID, listMembers(ds, db))
	case strings.HasPrefix(m.Content, "ac "):
		parts := strings.Split(m.Content[3:], "|")
		if len(parts) != 2 {
			sendMsg(ds, m.Author.ID, "Usage: ac jmeno|discordID")
			return
		}
		if addMember(ds, db, parts[0], parts[1]) {
			sendMsg(ds, m.Author.ID, "Member added")
		} else {
			sendMsg(ds, m.Author.ID, "Error adding member")
		}
	case strings.HasPrefix(m.Content, "as "):
		parts := strings.Split(m.Content[3:], "|")
		if len(parts) != 3 {
			sendMsg(ds, m.Author.ID, "Usage: as nazev|datum|clenID")
			return
		}
		cleni := strings.Split(parts[2], ",")
		cleniID := make([]int, 0)
		for _, id := range cleni {
			idI, err := strconv.Atoi(id)
			if err != nil {
				sendMsg(ds, m.Author.ID, "Invalid member ID")
				return
			}
			cleniID = append(cleniID, idI)
		}
		if addMeeting(ds, db, parts[0], parts[1], cleniID) {
			sendMsg(ds, m.Author.ID, "Meeting added")
		} else {
			sendMsg(ds, m.Author.ID, "Error adding meeting")
		}
	case strings.HasPrefix(m.Content, "rm "):
		id, err := strconv.Atoi(m.Content[3:])
		if err != nil {
			sendMsg(ds, m.Author.ID, "Invalid meeting ID")
			return
		}
		if removeMeeting(ds, db, id) {
			sendMsg(ds, m.Author.ID, "Meeting removed")
		} else {
			sendMsg(ds, m.Author.ID, "Error removing meeting")
		}
	case strings.HasPrefix(m.Content, "rc "):
		id, err := strconv.Atoi(m.Content[3:])
		if err != nil {
			sendMsg(ds, m.Author.ID, "Invalid member ID")
			return
		}
		if removeMember(ds, db, id) {
			sendMsg(ds, m.Author.ID, "Member removed")
		} else {
			sendMsg(ds, m.Author.ID, "Error removing member")
		}
	default:
		sendMsg(ds, m.Author.ID, "Unknown command")
	}
}

func sleepNext() <-chan time.Time {
	tn := time.Now()
	t := tn.Truncate(24 * time.Hour).Add((24 + 14) * time.Hour)
	if t.IsDST() {
		t = t.Add(-time.Hour)
	}
	return time.After(t.Sub(tn))
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
	sendAdmin = func(ds *discordgo.Session, err error) {
		sendMsg(ds, conf.Admin, "Error: "+err.Error())
	}
	// Open database
	db, err := sql.Open("sqlite3", conf.Database)
	if err != nil {
		log.Fatalln("Error opening database:", err)
	}
	db.Exec("PRAGMA foreign_keys = ON;") // Enable foreign keys
	// Open Discord session
	ds, err := discordgo.New("Bot " + conf.Token)
	if err != nil {
		log.Fatalln("Error creating Discord session:", err)
	}
	// Register command handler
	ds.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		commandHandler(s, m, db, conf.Admin)
	})
	ds.Identify.Intents = discordgo.IntentsDirectMessages
	// Open discord websocket
	err = ds.Open()
	if err != nil {
		log.Fatalln("Error opening connection:", err)
	}
	defer ds.Close()

	log.Println("Bot is now running.")

	for {
		<-sleepNext()
		sendNotification(ds, db, conf.NotifyChannel)
	}
}
