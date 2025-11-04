package infrastructure

import (
	"context"
	"net/http"

	"github.com/michaelheyman/spotify-setlist/internal/domain"
	"github.com/michaelheyman/spotify-setlist/internal/infrastructure/externalsdk"
	spotify "github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

type ExternalSpotifyClientFactory struct {
}

func (s ExternalSpotifyClientFactory) NewClient(httpClient *http.Client) domain.SpotifyClient {
	return NewSpotifyClient(spotify.New(httpClient))
}

type SpotifyAuth struct {
	authenticator externalsdk.SpotifyAuthenticator
}

func NewSpotifyAuth(authenticator externalsdk.SpotifyAuthenticator) *SpotifyAuth {
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
