package interfaces

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/michaelheyman/spotify-setlist/internal/domain"
	"golang.org/x/oauth2"
)

type SpotifyAuthHandler interface {
	StartAuthFlow(ctx context.Context) (string, error)
	WaitForClient(ctx context.Context) (domain.SpotifyClient, error)
}

type spotifyAuthHandler struct {
	spotify       domain.SpotifyClientFactory
	auth          domain.SpotifyAuthenticator
	state         string
	tokenStore    domain.TokenStore
	clientChan    chan domain.SpotifyClient
	server        *http.Server
	serverErrChan chan error
}

func NewSpotifyAuthHandler(spotify domain.SpotifyClientFactory, auth domain.SpotifyAuthenticator, tokenStore domain.TokenStore) SpotifyAuthHandler {
	return &spotifyAuthHandler{
		spotify:       spotify,
		auth:          auth,
		state:         "abc123",
		tokenStore:    tokenStore,
		clientChan:    make(chan domain.SpotifyClient),
		serverErrChan: make(chan error),
	}
}

func (s *spotifyAuthHandler) StartAuthFlow(ctx context.Context) (string, error) {
	token, err := s.tokenStore.LoadToken(ctx)
	if err != nil {
		return "", fmt.Errorf("loading token: %w", err)
	}
	if token != nil {
		// Token was found
		oauthToken := &oauth2.Token{
			AccessToken:  token.AccessToken,
			TokenType:    token.TokenType,
			RefreshToken: token.RefreshToken,
			Expiry:       token.Expiry,
		}
		client := s.spotify.NewClient(s.auth.Client(ctx, oauthToken))
		// Return the client in a non blocking routine
		go func() {
			s.clientChan <- client
		}()
		return "", nil
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", s.handleAuthCallback)
	s.server = &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.serverErrChan <- err
		}
	}()

	select {
	case err := <-s.serverErrChan:
		return "", fmt.Errorf("server failed to start: %w", err)
	case <-time.After(50 * time.Millisecond):
		// No-op: give the server some time to start
	}

	url := s.auth.AuthURL(s.state)

	return url, nil
}

func (s *spotifyAuthHandler) WaitForClient(ctx context.Context) (domain.SpotifyClient, error) {
	select {
	case client := <-s.clientChan:
		// Server not present indicates the refresh token was retrieved
		if s.server == nil {
			return client, nil
		}

		// Auth completed successfully
		if err := s.server.Shutdown(ctx); err != nil {
			return nil, fmt.Errorf("shutting down server after successful auth: %w", err)
		}
		return client, nil
	case <-ctx.Done():
		// Context timeout
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.server.Shutdown(shutdownCtx); err != nil && err != http.ErrServerClosed {
			return nil, fmt.Errorf("shutting down server on timeout: %w", err)
		}
		return nil, fmt.Errorf("waiting for client: %w", ctx.Err())
	}
}

func (s *spotifyAuthHandler) handleAuthCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	token, err := s.auth.Token(ctx, s.state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		return
	}
	if st := r.FormValue("state"); st != s.state {
		http.Error(w, "OAuth state mismatch", http.StatusUnauthorized)
		return
	}

	// Save new token
	if err := s.tokenStore.SaveToken(ctx, &domain.SpotifyToken{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	}); err != nil {
		http.Error(w, "Failed to save token", http.StatusInternalServerError)
		return
	}

	// use the token to get an authenticated client
	client := s.spotify.NewClient(s.auth.Client(ctx, token))
	fmt.Fprintf(w, "Login Completed!")
	s.clientChan <- client
}
