package domain

import (
	"context"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

// TODO: make this agnostic to the spotify library and create a concrete spotify client
type SpotifyClient interface {
	AddTracksToPlaylist(ctx context.Context, playlistID spotify.ID, trackIDs ...spotify.ID) (snapshotID string, err error)
	CreatePlaylistForUser(ctx context.Context, userID, playlistName, description string, public bool, collaborative bool) (*spotify.FullPlaylist, error)
	CurrentUser(ctx context.Context) (*spotify.PrivateUser, error)
	Search(ctx context.Context, query string, t spotify.SearchType, opts ...spotify.RequestOption) (*spotify.SearchResult, error)
}

type SpotifyClientFactory interface {
	NewClient(httpClient *http.Client) SpotifyClient
}

type SpotifyAuthenticator interface {
	AuthURL(state string) string
	Token(ctx context.Context, state string, r *http.Request) (*oauth2.Token, error)
	Client(ctx context.Context, token *oauth2.Token) *http.Client
}

type AuthenticationFactory interface {
	CreateAuthenticator(ctx context.Context, config SpotifyAuthConfig) (SpotifyAuthenticator, error)
}

type SpotifyAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Scopes       []string
}

func (c SpotifyAuthConfig) Validate() error {
	return validation.ValidateStruct(
		&c,
		validation.Field(&c.ClientID, validation.Required),
		validation.Field(&c.ClientSecret, validation.Required),
		validation.Field(&c.RedirectURI, validation.Required),
		validation.Field(&c.Scopes, validation.Required),
	)
}
