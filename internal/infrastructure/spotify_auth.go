package infrastructure

import (
	"context"
	"net/http"

	spotify "github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

type SpotifyClientFactory interface {
	NewClient(httpClient *http.Client) SpotifyClient
}

type SpotifyAuthenticator interface {
	AuthURL(state string) string
	Token(ctx context.Context, state string, r *http.Request) (*oauth2.Token, error)
	Client(ctx context.Context, token *oauth2.Token) *http.Client
}

type ExternalSpotifyClientFactory struct {
}

func (s ExternalSpotifyClientFactory) NewClient(httpClient *http.Client) SpotifyClient {
	return spotify.New(httpClient)
}

type ExternalSpotifyAuthenticator interface {
	AuthURL(state string, opts ...oauth2.AuthCodeOption) string
	Token(ctx context.Context, state string, r *http.Request, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
	Client(ctx context.Context, token *oauth2.Token) *http.Client
}

type SpotifyAuth struct {
	authenticator ExternalSpotifyAuthenticator
}

func NewSpotifyAuth(authenticator ExternalSpotifyAuthenticator) *SpotifyAuth {
	return &SpotifyAuth{
		authenticator: authenticator,
	}
}

func (a SpotifyAuth) AuthURL(state string) string {
	return a.authenticator.AuthURL(state)
}

func (a SpotifyAuth) Token(ctx context.Context, state string, r *http.Request) (*oauth2.Token, error) {
	return a.authenticator.Token(ctx, state, r)
}

func (a SpotifyAuth) Client(ctx context.Context, token *oauth2.Token) *http.Client {
	return a.authenticator.Client(ctx, token)
}
