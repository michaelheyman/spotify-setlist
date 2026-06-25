package spotify

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cli/browser"
	"github.com/michaelheyman/spotify-setlist/internal/domain"
	"golang.org/x/oauth2"
)

const authState = "abc123"

// URLOpener opens a URL in the user's environment (e.g. their web browser).
type URLOpener interface {
	Open(url string) error
}

// browserOpener is the production URLOpener backed by github.com/cli/browser.
type browserOpener struct{}

func (browserOpener) Open(url string) error {
	return browser.OpenURL(url)
}

// SpotifyAuthProvider drives the interactive OAuth flow. It owns the local
// callback server and the browser hand-off, but not persistence.

type SpotifyAuthProvider struct {
	spotify       domain.SpotifyClientFactory
	auth          domain.SpotifyAuthenticator
	opener        URLOpener
	out           io.Writer
	state         string
	tokenChan     chan *domain.SpotifyToken
	server        *http.Server
	serverErrChan chan error
}

func NewSpotifyAuthProvider(
	spotify domain.SpotifyClientFactory,
	auth domain.SpotifyAuthenticator,
	out io.Writer,
) *SpotifyAuthProvider {
	return &SpotifyAuthProvider{
		spotify:       spotify,
		auth:          auth,
		opener:        browserOpener{},
		out:           out,
		state:         authState,
		tokenChan:     make(chan *domain.SpotifyToken),
		serverErrChan: make(chan error),
	}
}

// ClientFromToken wraps an existing token into an authenticated client without
// any user interaction.
func (p *SpotifyAuthProvider) ClientFromToken(ctx context.Context, token *domain.SpotifyToken) domain.SpotifyClient {
	oauthToken := &oauth2.Token{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	}
	return p.spotify.NewClient(p.auth.Client(ctx, oauthToken))
}

// Authenticate runs the full interactive OAuth flow: it starts a local callback
// server, opens the user's browser to the Spotify authorization page, waits for
// the redirect, and returns the resulting token. It does not persist the token
// or build a client.
func (p *SpotifyAuthProvider) Authenticate(ctx context.Context) (*domain.SpotifyToken, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", p.handleAuthCallback)
	p.server = &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 3 * time.Second,
	}
	go func() {
		if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			p.serverErrChan <- err
		}
	}()

	select {
	case err := <-p.serverErrChan:
		return nil, fmt.Errorf("server failed to start: %w", err)
	case <-time.After(50 * time.Millisecond):
		// No-op: give the server some time to start
	}

	authURL := p.auth.AuthURL(p.state)
	fmt.Fprintf(p.out, "Auth URL: %s\n\n", authURL)
	if err := p.opener.Open(authURL); err != nil {
		_ = p.shutdown(ctx)
		return nil, fmt.Errorf("opening browser: %w", err)
	}
	fmt.Fprintf(p.out, "Your browser has been opened to visit:\n\n\t%s\n\n", authURL)

	return p.waitForToken(ctx)
}

func (p *SpotifyAuthProvider) waitForToken(ctx context.Context) (*domain.SpotifyToken, error) {
	select {
	case token := <-p.tokenChan:
		if err := p.shutdown(ctx); err != nil {
			return nil, fmt.Errorf("shutting down server after successful auth: %w", err)
		}
		return token, nil
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := p.server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return nil, fmt.Errorf("shutting down server on timeout: %w", err)
		}
		return nil, fmt.Errorf("waiting for token: %w", ctx.Err())
	}
}

func (p *SpotifyAuthProvider) shutdown(ctx context.Context) error {
	if err := p.server.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (p *SpotifyAuthProvider) handleAuthCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	token, err := p.auth.Token(ctx, p.state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		return
	}
	if st := r.FormValue("state"); st != p.state {
		http.Error(w, "OAuth state mismatch", http.StatusUnauthorized)
		return
	}

	fmt.Fprintf(w, "Login Completed!")
	p.tokenChan <- &domain.SpotifyToken{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	}
}
