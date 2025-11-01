package interfaces

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/michaelheyman/spotify-setlist/internal/domain"
	"github.com/michaelheyman/spotify-setlist/internal/domain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2"
)

func Test_spotifyAuthHandler_StartAuthFlow(t *testing.T) {
	state := "abc123"
	authURL := "https://accounts.spotify.com/authorize?client_id=11111111111111111111111111111111&redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Fcallback&response_type=code&scope=playlist-modify-private+playlist-read-private+user-read-private+user-read-email&state=abc123"
	mockClient := &mocks.SpotifyClient{}
	token := &oauth2.Token{
		AccessToken:  "some-access-token",
		RefreshToken: "some-refresh-token",
	}

	tests := []struct {
		name            string
		setExpectations func(c *mocks.SpotifyClientFactory, a *mocks.SpotifyAuthenticator)
		callbackState   string
		wantStatus      int
		want            string
		wantErr         assert.ErrorAssertionFunc
	}{
		{
			name: "successfully starts auth flow",
			setExpectations: func(c *mocks.SpotifyClientFactory, a *mocks.SpotifyAuthenticator) {
				a.EXPECT().
					AuthURL(state).
					Return(authURL)
				a.EXPECT().
					Token(mock.Anything, state, mock.Anything).
					Return(token, nil)
				a.EXPECT().
					Client(mock.Anything, token).
					Return(&http.Client{})
				c.EXPECT().
					NewClient(mock.Anything).
					Return(mockClient)
			},
			callbackState: state,
			wantStatus:    http.StatusOK,
			want:          authURL,
			wantErr:       assert.NoError,
		},
		{
			name: "should return forbidden error when token exchange fails",
			setExpectations: func(c *mocks.SpotifyClientFactory, a *mocks.SpotifyAuthenticator) {
				a.EXPECT().
					AuthURL(state).
					Return(authURL)
				a.EXPECT().
					Token(mock.Anything, state, mock.Anything).
					Return(nil, assert.AnError)
			},
			callbackState: state,
			wantStatus:    http.StatusForbidden,
			want:          authURL,
			wantErr:       assert.NoError,
		},
		{
			name: "should return unauthorized error when there is an OAuth state mismatch",
			setExpectations: func(c *mocks.SpotifyClientFactory, a *mocks.SpotifyAuthenticator) {
				a.EXPECT().
					AuthURL(state).
					Return(authURL)
				a.EXPECT().
					Token(mock.Anything, state, mock.Anything).
					Return(token, nil)
			},
			callbackState: "mismatched-state",
			wantStatus:    http.StatusUnauthorized,
			want:          authURL,
			wantErr:       assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &mocks.SpotifyClientFactory{}
			a := &mocks.SpotifyAuthenticator{}
			tt.setExpectations(c, a)
			defer c.AssertExpectations(t)
			defer a.AssertExpectations(t)

			// Create a context with timeout for the entire test
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			// Don't use the constructor so that you can interact with the private struct fields
			handler := &spotifyAuthHandler{
				spotify:       c,
				auth:          a,
				state:         state,
				clientChan:    make(chan domain.SpotifyClient),
				serverErrChan: make(chan error),
			}

			// Start the auth flow
			gotURL, err := handler.StartAuthFlow(ctx)
			tt.wantErr(t, err)
			assert.Equal(t, tt.want, gotURL)

			callbackURL, err := createQueryURL(tt.callbackState)
			if err != nil {
				t.Fatalf("creating query URL: %v", err)
			}
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, callbackURL, nil)
			assert.NoError(t, err)

			// Make request in a goroutine to avoid blocking
			client := &http.Client{Timeout: 3 * time.Second}
			respChan := make(chan *http.Response)
			errChan := make(chan error)
			go func() {
				resp, err := client.Do(req)
				if err != nil {
					errChan <- err
					return
				}
				respChan <- resp
			}()

			// Wait for either response or client
			select {
			case resp := <-respChan:
				assert.NotNil(t, resp)
				defer resp.Body.Close()
				assert.Equal(t, tt.wantStatus, resp.StatusCode)
			case client := <-handler.clientChan:
				assert.NotNil(t, client)
			case err := <-errChan:
				t.Fatalf("HTTP request failed: %v", err)
			case <-ctx.Done():
				t.Fatal("Test timed out")
			}

			// Cleanup
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer shutdownCancel()
			assert.NoError(t, handler.server.Shutdown(shutdownCtx))
		})
	}
}

func createQueryURL(state string) (string, error) {
	callbackURL := "http://localhost:8080/callback"

	u, err := url.Parse(callbackURL)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Add("state", state)

	u.RawQuery = q.Encode()

	return u.String(), nil
}
