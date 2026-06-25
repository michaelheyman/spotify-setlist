package auth

import (
	"context"
	"fmt"

	"github.com/michaelheyman/spotify-setlist/internal/domain"
)

// SessionService produces a ready-to-use authenticated Spotify client for the
// current invocation, owning token persistence and client construction.
type SessionService interface {
	AuthenticatedClient(ctx context.Context) (domain.SpotifyClient, error)
}

type sessionService struct {
	tokenStore domain.TokenStore
	provider   domain.SpotifyAuthProvider
}

func NewSessionService(tokenStore domain.TokenStore, provider domain.SpotifyAuthProvider) SessionService {
	return sessionService{
		tokenStore: tokenStore,
		provider:   provider,
	}
}

// AuthenticatedClient reuses a stored token when present, otherwise runs the
// interactive OAuth flow and persists the resulting token.
func (s sessionService) AuthenticatedClient(ctx context.Context) (domain.SpotifyClient, error) {
	token, found, err := s.tokenStore.LoadToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading token: %w", err)
	}

	if found {
		return s.provider.ClientFromToken(ctx, token), nil
	}

	token, err = s.provider.Authenticate(ctx)
	if err != nil {
		return nil, fmt.Errorf("authenticating: %w", err)
	}
	if err := s.tokenStore.SaveToken(ctx, token); err != nil {
		return nil, fmt.Errorf("saving token: %w", err)
	}

	return s.provider.ClientFromToken(ctx, token), nil
}
