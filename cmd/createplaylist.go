/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// createPlaylistCmd represents the createPlaylist command
var createPlaylistCmd = &cobra.Command{
	Use:   "create-playlist",
	Short: "Create a Spotify playlist from an artist's setlist",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("create-playlist called")
	},
}

func init() {
	rootCmd.AddCommand(createPlaylistCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createPlaylistCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createPlaylistCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
