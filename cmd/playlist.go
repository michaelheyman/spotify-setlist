package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/cli/browser"
	application "github.com/michaelheyman/spotify-setlist/internal/application/playlist"
	"github.com/michaelheyman/spotify-setlist/internal/domain"
	"github.com/michaelheyman/spotify-setlist/internal/infrastructure/setlistfm"
	"github.com/michaelheyman/spotify-setlist/internal/infrastructure/spotify"
	"github.com/michaelheyman/spotify-setlist/internal/interfaces"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

const defaultMinSongs = 5

type CreatePlaylistCmd struct {
	authFactory domain.AuthenticationFactory
	httpClient  *http.Client
	service     application.PlaylistService
	tokenStore  domain.TokenStore
}

func NewCreatePlaylistCmd(
	authFactory domain.AuthenticationFactory,
	httpClient *http.Client,
	tokenStore domain.TokenStore,
) *cobra.Command {
	createCmd := CreatePlaylistCmd{
		authFactory: authFactory,
		httpClient:  httpClient,
		tokenStore:  tokenStore,
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
	cmd.Flags().
		IntP("min-songs", "m", defaultMinSongs, "Sets the minimum number of songs a setlist needs to be included")

	viper.SetEnvPrefix("")
	viper.AutomaticEnv() // Read environment variables

	// Environment variables to look for
	_ = viper.BindEnv("SPOTIFY_CLIENT_ID")
	_ = viper.BindEnv("SPOTIFY_CLIENT_SECRET")
	_ = viper.BindEnv("SPOTIFY_REDIRECT_URI")
	_ = viper.BindEnv("SETLIST_FM_API_KEY")

	// Optional flags
	cmd.Flags().String("spotify-client-id", "", "Spotify client ID (env: SPOTIFY_CLIENT_ID)")
	cmd.Flags().String("spotify-client-secret", "", "Spotify client secret (env: SPOTIFY_CLIENT_SECRET)")
	cmd.Flags().
		String("spotify-redirect-uri", "http://localhost:8080/callback", "Spotify redirect URI (env: SPOTIFY_REDIRECT_URI)")
	cmd.Flags().String("setlistfm-api-key", "", "Setlist.fm API key (env: SETLIST_FM_API_KEY)")

	// Bind flags to viper keys for later lookups
	_ = viper.BindPFlag("SPOTIFY_CLIENT_ID", cmd.Flags().Lookup("spotify-client-id"))
	_ = viper.BindPFlag("SPOTIFY_CLIENT_SECRET", cmd.Flags().Lookup("spotify-client-secret"))
	_ = viper.BindPFlag("SPOTIFY_REDIRECT_URI", cmd.Flags().Lookup("spotify-redirect-uri"))
	_ = viper.BindPFlag("SETLIST_FM_API_KEY", cmd.Flags().Lookup("setlistfm-api-key"))

	return cmd
}

// PreRunE validates requirements and instantiates runtime dependencies before Run().
func (c *CreatePlaylistCmd) PreRunE(cmd *cobra.Command, _ []string) error {
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

	spotifyFactory := spotify.ExternalSpotifyClientFactory{}
	handler := interfaces.NewSpotifyAuthHandler(spotifyFactory, authenticator, c.tokenStore)

	authURL, err := handler.StartAuthFlow(ctx)
	if err != nil {
		return err
	}

	// AuthURL will be empty if a refresh token was found
	if authURL != "" {
		if err := browser.OpenURL(authURL); err != nil {
			return err
		}
		fmt.Fprintf(stderr, "Your browser has been opened to visit:\n\n\t%s\n\n", authURL)
	}

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
		setlistfm.NewSetlistFMService(
			c.httpClient,
			apiKey,
		),
		spotify.NewSpotifyService(client),
	)
	return nil
}

func (c *CreatePlaylistCmd) RunE(cmd *cobra.Command, _ []string) error {
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

	rendered := renderResult(result)

	fmt.Fprintln(cmd.OutOrStdout(), rendered)

	return nil
}

func renderResult(result application.CreatePlaylistResult) string {
	var sections []string

	sections = append(sections, renderHeader(result.Playlist.Artist, result.Setlist))

	sections = append(sections, renderPlaylist(result.Playlist))

	if len(result.MissingSongs) > 0 {
		sections = append(sections, renderMissingSongs(result.MissingSongs))
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func renderHeader(name string, setlist domain.Setlist) string {
	title := lipgloss.NewStyle().
		Render(fmt.Sprintf("%s Playlist", name))

	details := lipgloss.JoinHorizontal(
		lipgloss.Left,
		setlist.Venue, " - ", setlist.EventDate.Format(time.DateOnly),
	)

	div := lipgloss.NewStyle().
		Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	var headerParts []string
	headerParts = append(headerParts, title, details, div)

	if setlist.URL != "" {
		headerParts = append(
			headerParts,
			lipgloss.NewStyle().Render("Source: "+setlist.URL),
		)
	}

	return lipgloss.JoinVertical(lipgloss.Left, append(headerParts, "")...)
}

func renderPlaylist(playlist domain.Playlist) string {
	var rows [][]string
	for i, song := range playlist.Songs {
		rows = append(rows, []string{
			fmt.Sprintf("%d", i+1),
			playlist.Artist,
			song,
		})
	}

	t := table.New().
		Border(lipgloss.HiddenBorder()).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch col {
			case 0:
				return lipgloss.NewStyle().
					PaddingLeft(1).
					PaddingRight(1).
					Align(lipgloss.Right)
			case 1:
				return lipgloss.NewStyle().
					PaddingLeft(1).
					PaddingRight(1).
					Align(lipgloss.Center)
			case 2:
				return lipgloss.NewStyle().
					PaddingLeft(1).
					PaddingRight(1).
					Align(lipgloss.Left)
			default:
				return lipgloss.NewStyle().Align(lipgloss.Left)
			}
		}).
		Headers("#", "ARTIST", "TITLE").
		Rows(rows...)

	return t.String()
}

func renderMissingSongs(missingSongs []string) string {
	var rows [][]string
	for _, song := range missingSongs {
		rows = append(rows, []string{song})
	}

	t := table.New().
		Border(lipgloss.HiddenBorder()).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch col {
			case 0:
				return lipgloss.NewStyle().
					PaddingLeft(1).
					PaddingRight(1).
					Align(lipgloss.Right)
			case 1:
				return lipgloss.NewStyle().
					PaddingLeft(1).
					PaddingRight(1).
					Align(lipgloss.Center)
			case 2:
				return lipgloss.NewStyle().
					PaddingLeft(1).
					PaddingRight(1).
					Align(lipgloss.Left)
			default:
				return lipgloss.NewStyle().Align(lipgloss.Left)
			}
		}).
		Headers("MISSING SONGS").
		Rows(rows...)

	return t.String()
}
