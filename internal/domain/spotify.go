package domain

import (
	"context"
	"net/http"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"golang.org/x/oauth2"
)

type SpotifyClient interface {
	CreatePlaylist(ctx context.Context, user string, playlist SpotifyPlaylist) error
	CurrentUser(ctx context.Context) (string, error)
	SearchTrack(ctx context.Context, artist, title string) (SpotifySearchResult, error)
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

type TokenStore interface {
	SaveToken(ctx context.Context, token *SpotifyToken) error
	LoadToken(ctx context.Context) (*SpotifyToken, error)
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

type SpotifyPlaylist struct {
	Playlist
	Tracks        []SpotifyTrack
	Public        bool
	Collaborative bool
}

type SpotifySearchResult struct {
	Tracks []SpotifyTrack
}

type SpotifyTrack struct {
	Name       string
	ID         string
	Popularity int
}

type SpotifyToken struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expiry"`
}

func (t SpotifyToken) IsExpired() bool {
	return t.Expiry.Before(time.Now())
}
