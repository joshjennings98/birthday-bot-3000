package cmd

import (
	"context"
	"os"

	bot "github.com/joshjennings98/discord-bot/discord_bot"
	"github.com/joshjennings98/discord-bot/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	app = "discord_bot"
	// CLI flags
	Token       = "token"
	BirthdaysDB = "birthdays_db"
	Channel     = "channel"
	Server      = "server"
)

var (
	viperSession = viper.New()
)

var rootCmd = &cobra.Command{
	Use:   "discord-bot",
	Short: "TODO",
	Long:  `TODO`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		if err := RunCLI(ctx); err != nil {
			return err
		}
		return nil
	},
	SilenceUsage: true, // otherwise 'Usage' is printed after any error
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func initCLI(ctx context.Context) (err error) {
	if err := utils.LoadFromViper(viperSession, app, &bot.BotConfig, bot.DefaultBotConfig()); err != nil {
		return err
	}
	return nil
}

func init() {
	rootCmd.Flags().StringP(Token, "t", "", "Bot token")
	rootCmd.Flags().StringP(BirthdaysDB, "d", "", "Birthdays database")
	rootCmd.Flags().StringP(Channel, "c", "", "Channel to send message on")
	rootCmd.Flags().StringP(Server, "s", "", "Server to send message on")

	_ = utils.BindFlagToEnv(viperSession, app, "DISCORD_BOT_TOKEN", rootCmd.Flags().Lookup(Token))
	_ = utils.BindFlagToEnv(viperSession, app, "DISCORD_BOT_BIRTHDAYS_DB", rootCmd.Flags().Lookup(BirthdaysDB))
	_ = utils.BindFlagToEnv(viperSession, app, "DISCORD_BOT_CHANNEL", rootCmd.Flags().Lookup(Channel))
	_ = utils.BindFlagToEnv(viperSession, app, "DISCORD_BOT_SERVER", rootCmd.Flags().Lookup(Server))
}

func RunCLI(ctx context.Context) error {
	if err := initCLI(ctx); err != nil {
		log.Errorf("Failed to initialise CLI with error: %s", err)
		return err
	}

	return bot.StartBot()
}
