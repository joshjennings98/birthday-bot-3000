package discord_bot

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	commands "github.com/joshjennings98/discord-bot/birthday"
	"github.com/joshjennings98/discord-bot/utils"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	BotConfig  commands.BotConfiguration
	DiscordBot commands.DiscordBot
)

const (
	prefixCmd = "!bd"
	defaultDB = ""
)

func ConnectToMongoDB(ctx context.Context) (c *mongo.Client) {
	client, err := mongo.NewClient(options.Client().ApplyURI(BotConfig.MongoDBURI))
	if err != nil {
		log.Fatal(err)
	}
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}
	databases, err := client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("databases", databases)

	commands.BirthdaysDatabase = client.Database("BirthdaysDatabase")
	return client
}

func StartBot() (err error) {
	// connect to mongodb
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client := ConnectToMongoDB(ctx)
	defer client.Disconnect(ctx)

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + BotConfig.Token)
	if err != nil {
		return fmt.Errorf("error creating Discord session: %w", err)
	}
	defer dg.Close()
	dg.AddHandler(messageCreate)
	dg.AddHandler(onReady)
	// We only care about receiving message events.
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	// Attach DiscordBot to session
	DiscordBot.AttachBotToSession(dg)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		return fmt.Errorf("error opening connection: %w", err)
	}

	// Wait here until CTRL-C or other term signal is received.
	log.Info("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	return
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Check for prefix
	if strings.HasPrefix(m.Content, prefixCmd) {
		DiscordBot.ExecuteCommand(m)
	}
}

func onReady(s *discordgo.Session, _ *discordgo.Ready) {
	ticker := time.NewTicker(1 * time.Hour)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:

				log.Info("Checking for birthdays")

				databases, err := commands.GetServerKeys()
				if err != nil {
					log.Errorf("Could not find databases")
				}
				for _, db := range databases {
					tz, err := commands.GetTimezone(db)
					if err != nil {
						log.Error(fmt.Sprintf("Could not get timezone from database %s", db))
					}
					loc, err := time.LoadLocation(tz)
					if err != nil {
						log.Errorf("Invalid location '%s'", loc)
					}
					interval, err := commands.GetTimeInterval(db)
					if err != nil {
						log.Errorf("Could not get time interval from database '%s'", db)
					}
					i, err := strconv.Atoi(interval)
					if err != nil {
						log.Errorf("Invalid interval '%s'", interval)
					}
					if utils.InHourInterval(i, time.Now().In(loc)) {
						commands.WishTodaysHappyBirthdays(s, db)
					}

				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}
