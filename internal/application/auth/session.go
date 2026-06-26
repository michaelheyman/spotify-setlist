package auth

import (
	"context"
	"errors"
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

// AuthenticatedClient reuses a stored token when present, verifying it is still
// valid. If the stored session has expired it discards the token and
// re-authenticates, so callers always receive a client backed by a live
// session. When no token is stored it authenticates from scratch.
func (s sessionService) AuthenticatedClient(ctx context.Context) (domain.SpotifyClient, error) {
	token, found, err := s.tokenStore.LoadToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading token: %w", err)
	}

	if !found {
		return s.authenticate(ctx)
	}

	client := s.provider.ClientFromToken(ctx, token)
	if _, err := client.CurrentUser(ctx); err != nil {
		if errors.Is(err, domain.ErrSessionExpired) {
			if err := s.tokenStore.DeleteToken(ctx); err != nil {
				return nil, fmt.Errorf("discarding expired token: %w", err)
			}
			return s.authenticate(ctx)
		}
		return nil, fmt.Errorf("verifying session: %w", err)
	}

	return client, nil
}

// authenticate runs the interactive OAuth flow, persists the resulting token,
// and returns a client for it.
func (s sessionService) authenticate(ctx context.Context) (domain.SpotifyClient, error) {
	token, err := s.provider.Authenticate(ctx)
	if err != nil {
		return nil, fmt.Errorf("authenticating: %w", err)
	}
	if err := s.tokenStore.SaveToken(ctx, token); err != nil {
		return nil, fmt.Errorf("saving token: %w", err)
	}

	return s.provider.ClientFromToken(ctx, token), nil
}
