/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/michaelheyman/spotify-setlist/cmd"
	"github.com/michaelheyman/spotify-setlist/internal/infrastructure"
)

func main() {
	ctx := context.Background()

	deps := &cmd.Dependencies{
		AuthFactory: infrastructure.NewSpotifyAuthFactory(),
		HttpClient:  &http.Client{Timeout: 15 * time.Second},
	}

	rootCmd := cmd.NewRootCmd(deps)
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
