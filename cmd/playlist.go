/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/cli/browser"
	application "github.com/michaelheyman/spotify-setlist/internal/application/playlist"
	"github.com/michaelheyman/spotify-setlist/internal/domain"
	"github.com/michaelheyman/spotify-setlist/internal/infrastructure"
	"github.com/michaelheyman/spotify-setlist/internal/interfaces"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

type CreatePlaylistCmd struct {
	authFactory domain.AuthenticationFactory
	httpClient  *http.Client
	service     application.PlaylistService
}

func NewCreatePlaylistCmd(authFactory domain.AuthenticationFactory, httpClient *http.Client) *cobra.Command {
	createCmd := CreatePlaylistCmd{
		authFactory: authFactory,
		httpClient:  httpClient,
	}

	cmd := &cobra.Command{
		Use:   "create-playlist",
		Short: "Create a Spotify playlist from an artist's setlist",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		PreRunE: createCmd.PreRunE,
		RunE:    createCmd.RunE,
	}

	// Required flags
	cmd.Flags().StringP("artist", "a", "", "Specify the artist that the setlist playlist will be created for")
	cmd.Flags().IntP("min-songs", "m", 5, "Sets the minimum number of songs a setlist needs to be included")

	viper.SetEnvPrefix("")
	viper.AutomaticEnv() // Read environment variables

	// Environment variables to look for
	viper.BindEnv("SPOTIFY_CLIENT_ID")
	viper.BindEnv("SPOTIFY_CLIENT_SECRET")
	viper.BindEnv("SPOTIFY_REDIRECT_URI")
	viper.BindEnv("SETLIST_FM_API_KEY")

	// Optional flags
	cmd.Flags().String("spotify-client-id", "", "Spotify client ID (env: SPOTIFY_CLIENT_ID)")
	cmd.Flags().String("spotify-client-secret", "", "Spotify client secret (env: SPOTIFY_CLIENT_SECRET)")
	cmd.Flags().String("spotify-redirect-uri", "http://localhost:8080/callback", "Spotify redirect URI (env: SPOTIFY_REDIRECT_URI)")
	cmd.Flags().String("setlistfm-api-key", "", "Setlist.fm API key (env: SETLIST_FM_API_KEY)")

	// Bind flags to viper keys for later lookups
	viper.BindPFlag("SPOTIFY_CLIENT_ID", cmd.Flags().Lookup("spotify-client-id"))
	viper.BindPFlag("SPOTIFY_CLIENT_SECRET", cmd.Flags().Lookup("spotify-client-secret"))
	viper.BindPFlag("SPOTIFY_REDIRECT_URI", cmd.Flags().Lookup("spotify-redirect-uri"))
	viper.BindPFlag("SETLIST_FM_API_KEY", cmd.Flags().Lookup("setlistfm-api-key"))

	return cmd
}

// PreRunE validates requirements and instantiates runtime dependencies before Run().
func (c *CreatePlaylistCmd) PreRunE(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	stdout := cmd.OutOrStdout()
	stderr := cmd.ErrOrStderr()

	spotifyAuthConfig := domain.SpotifyAuthConfig{
		ClientID:     viper.GetString("SPOTIFY_CLIENT_ID"),
		ClientSecret: viper.GetString("SPOTIFY_CLIENT_SECRET"),
		RedirectURI:  viper.GetString("SPOTIFY_REDIRECT_URI"),
		Scopes: []string{
			spotifyauth.ScopePlaylistModifyPrivate,
			spotifyauth.ScopePlaylistReadPrivate,
			spotifyauth.ScopeUserReadPrivate,
			spotifyauth.ScopeUserReadEmail,
		},
	}

	apiKey := viper.GetString("SETLIST_FM_API_KEY")
	if apiKey == "" {
		return errors.New("missing Setlist.FM API key")
	}

	if spotifyAuthConfig.ClientID == "" {
		return errors.New("missing Spotify client ID")
	}
	if spotifyAuthConfig.ClientSecret == "" {
		return errors.New("missing Spotify client secret")
	}
	if spotifyAuthConfig.RedirectURI == "" {
		return errors.New("missing Spotify redirect URI")
	}

	authenticator, err := c.authFactory.CreateAuthenticator(ctx, spotifyAuthConfig)
	if err != nil {
		return err
	}

	spotifyFactory := infrastructure.ExternalSpotifyClientFactory{}
	handler := interfaces.NewSpotifyAuthHandler(spotifyFactory, authenticator)

	authURL, err := handler.StartAuthFlow(ctx)
	if err != nil {
		return err
	}

	if err := browser.OpenURL(authURL); err != nil {
		return err
	}
	fmt.Fprintf(stderr, "Your browser has been opened to visit:\n\n\t%s\n\n", authURL)

	client, err := handler.WaitForClient(ctx)
	if err != nil {
		return err
	}

	user, err := client.CurrentUser(ctx)
	if err != nil {
		return err
	}

	fmt.Fprintf(stdout, "logged in as user %s\n", user)
	c.service = application.NewPlaylistService(
		infrastructure.NewSetlistFMService(
			c.httpClient,
			apiKey,
		),
		infrastructure.NewSpotifyService(client),
	)
	return nil
}
func (c *CreatePlaylistCmd) RunE(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	artist, err := cmd.Flags().GetString("artist")
	if err != nil {
		return err
	}
	minSongs, err := cmd.Flags().GetInt("min-songs")
	if err != nil {
		return err
	}

	result, err := c.service.CreatePlaylist(ctx, application.CreatePlaylistParams{
		Artist:                  artist,
		MinimumSetlistSongCount: minSongs,
	})
	if err != nil {
		return fmt.Errorf("creating playlist from setlist: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created playlist for %s\n", result.Playlist.Artist)

	return nil
}
