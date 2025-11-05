/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/michaelheyman/spotify-setlist/cmd"
	"github.com/michaelheyman/spotify-setlist/internal/infrastructure/spotify"
)

func main() {
	ctx := context.Background()

	tokenStore, err := spotify.NewFileTokenStore()
	if err != nil {
		fmt.Printf("creating file token store: %v", err)
		os.Exit(1)
	}

	deps := &cmd.Dependencies{
		AuthFactory: spotify.NewSpotifyAuthFactory(),
		TokenStore:  tokenStore,
		HTTPClient:  &http.Client{Timeout: 15 * time.Second},
	}

	rootCmd := cmd.NewRootCmd(deps)
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
