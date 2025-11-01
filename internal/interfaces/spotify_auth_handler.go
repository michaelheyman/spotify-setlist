package interfaces

import (
	"context"
	"fmt"

	"github.com/michaelheyman/spotify-setlist/internal/application"
	"github.com/michaelheyman/spotify-setlist/internal/domain"
)

type SpotifyAuthHandler interface {
	Authenticate(ctx context.Context) (domain.SpotifyClient, error)
}

type spotifyAuthHandler struct {
	service application.SpotifyAuthService
}

func NewSpotifyAuthHandler(service application.SpotifyAuthService) SpotifyAuthHandler {
	return &spotifyAuthHandler{
		service: service,
	}
}
func (s *spotifyAuthHandler) Authenticate(ctx context.Context) (domain.SpotifyClient, error) {
	authURL, err := s.service.StartAuthFlow(ctx)
	if err != nil {
		return nil, fmt.Errorf("starting auth flow: %w", err)
	}
	fmt.Println("\nAUTH_URL: ", authURL)

	client, err := s.service.WaitForClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("waiting for client: %w", err)
	}
	return client, nil
}
