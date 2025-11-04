package spotify

import (
	"context"
	"fmt"

	"github.com/michaelheyman/spotify-setlist/internal/domain"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

type SpotifyAuthFactory struct{}

func NewSpotifyAuthFactory() *SpotifyAuthFactory {
	return &SpotifyAuthFactory{}
}

func (f SpotifyAuthFactory) CreateAuthenticator(ctx context.Context, config domain.SpotifyAuthConfig) (domain.SpotifyAuthenticator, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("validating authenticator config: %w", err)
	}

	authenticator := spotifyauth.New(
		spotifyauth.WithClientID(config.ClientID),
		spotifyauth.WithClientSecret(config.ClientSecret),
		spotifyauth.WithRedirectURL(config.RedirectURI),
		spotifyauth.WithScopes(config.Scopes...),
	)

	return NewSpotifyAuth(authenticator), nil
}
