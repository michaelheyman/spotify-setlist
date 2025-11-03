/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/michaelheyman/spotify-setlist/internal/domain"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

type Dependencies struct {
	AuthFactory domain.AuthenticationFactory
	HttpClient  *http.Client
}

func NewRootCmd(deps *Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spotify-setlist",
		Short: "A brief description of your application",
		Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		// Uncomment the following line if your bare application
		// has an action associated with it:
		// Run: func(cmd *cobra.Command, args []string) { },
	}

	cobra.OnInitialize(initConfig)

	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.spotify-setlist.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	cmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output")

	cmd.AddCommand(NewCreatePlaylistCmd(deps.AuthFactory, deps.HttpClient))

	return cmd
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".spotify-setlist" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".spotify-setlist")
	}

	if err := godotenv.Load(); err != nil {
		// It's ok if .env doesn't exist. Add a debug/verbose message here later.
	}
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
