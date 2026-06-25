package spotify

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/michaelheyman/spotify-setlist/internal/domain"
	"github.com/michaelheyman/spotify-setlist/internal/domain/mocks"
	spotifymocks "github.com/michaelheyman/spotify-setlist/internal/infrastructure/spotify/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestSpotifyAuthProvider_ClientFromToken(t *testing.T) {
	storedToken := &domain.SpotifyToken{
		AccessToken:  "access-token",
		TokenType:    "Bearer",
		RefreshToken: "refresh-token",
		Expiry:       time.Date(2025, time.November, 4, 15, 15, 53, 0, time.UTC),
	}
	mockClient := mocks.NewSpotifyClient(t)

	a := mocks.NewSpotifyAuthenticator(t)
	c := mocks.NewSpotifyClientFactory(t)
	a.EXPECT().
		Client(mock.Anything, &oauth2.Token{
			AccessToken:  storedToken.AccessToken,
			TokenType:    storedToken.TokenType,
			RefreshToken: storedToken.RefreshToken,
			Expiry:       storedToken.Expiry,
		}).
		Return(&http.Client{})
	c.EXPECT().NewClient(mock.Anything).Return(mockClient)

	provider := NewSpotifyAuthProvider(c, a, io.Discard)
	got := provider.ClientFromToken(context.Background(), storedToken)

	assert.Equal(t, mockClient, got)
}

func TestSpotifyAuthProvider_Authenticate(t *testing.T) {
	state := authState
	authURL := "https://accounts.spotify.com/authorize?state=abc123"
	oauthToken := &oauth2.Token{
		AccessToken:  "some-access-token",
		RefreshToken: "some-refresh-token",
	}
	wantToken := &domain.SpotifyToken{
		AccessToken:  oauthToken.AccessToken,
		RefreshToken: oauthToken.RefreshToken,
	}

	validCallbackURL, err := callbackQueryURL(state)
	require.NoError(t, err)
	mismatchedCallbackURL, err := callbackQueryURL("mismatched-state")
	require.NoError(t, err)

	tests := []struct {
		name string
		// timeout bounds how long Authenticate waits; rejection cases never
		// deliver a token, so they resolve via this deadline.
		timeout         time.Duration
		setExpectations func(a *mocks.SpotifyAuthenticator, opener *spotifymocks.URLOpener)
		wantToken       *domain.SpotifyToken
		wantErr         assert.ErrorAssertionFunc
	}{
		{
			name:    "should return the token when the interactive flow completes",
			timeout: 3 * time.Second,
			setExpectations: func(a *mocks.SpotifyAuthenticator, opener *spotifymocks.URLOpener) {
				a.EXPECT().AuthURL(state).Return(authURL)
				a.EXPECT().Token(mock.Anything, state, mock.Anything).Return(oauthToken, nil)
				// Opening the "browser" stands in for the user completing the
				// login: it fires the OAuth redirect at the local callback server.
				opener.EXPECT().Open(authURL).Run(func(_ string) {
					go fireCallback(validCallbackURL)
				}).Return(nil)
			},
			wantToken: wantToken,
			wantErr:   assert.NoError,
		},
		{
			name:    "should return error when the browser cannot be opened",
			timeout: 3 * time.Second,
			setExpectations: func(a *mocks.SpotifyAuthenticator, opener *spotifymocks.URLOpener) {
				a.EXPECT().AuthURL(state).Return(authURL)
				opener.EXPECT().Open(authURL).Return(assert.AnError)
			},
			wantToken: nil,
			wantErr: func(_ assert.TestingT, err error, _ ...any) bool {
				require.ErrorIs(t, err, assert.AnError)
				require.ErrorContains(t, err, "opening browser:")
				return true
			},
		},
		{
			name:    "should time out when the callback token exchange fails",
			timeout: 500 * time.Millisecond,
			setExpectations: func(a *mocks.SpotifyAuthenticator, opener *spotifymocks.URLOpener) {
				a.EXPECT().AuthURL(state).Return(authURL)
				a.EXPECT().Token(mock.Anything, state, mock.Anything).Return(nil, assert.AnError)
				opener.EXPECT().Open(authURL).Run(func(_ string) {
					go fireCallback(validCallbackURL)
				}).Return(nil)
			},
			wantToken: nil,
			wantErr: func(_ assert.TestingT, err error, _ ...any) bool {
				require.ErrorContains(t, err, "waiting for token:")
				return true
			},
		},
		{
			name:    "should time out when the callback state does not match",
			timeout: 500 * time.Millisecond,
			setExpectations: func(a *mocks.SpotifyAuthenticator, opener *spotifymocks.URLOpener) {
				a.EXPECT().AuthURL(state).Return(authURL)
				a.EXPECT().Token(mock.Anything, state, mock.Anything).Return(oauthToken, nil)
				opener.EXPECT().Open(authURL).Run(func(_ string) {
					go fireCallback(mismatchedCallbackURL)
				}).Return(nil)
			},
			wantToken: nil,
			wantErr: func(_ assert.TestingT, err error, _ ...any) bool {
				require.ErrorContains(t, err, "waiting for token:")
				return true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := mocks.NewSpotifyAuthenticator(t)
			c := mocks.NewSpotifyClientFactory(t)
			opener := spotifymocks.NewURLOpener(t)
			tt.setExpectations(a, opener)

			provider := NewSpotifyAuthProvider(c, a, io.Discard)
			provider.opener = opener

			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			token, err := provider.Authenticate(ctx)

			tt.wantErr(t, err)
			assert.Equal(t, tt.wantToken, token)
		})
	}
}

// fireCallback simulates the Spotify OAuth redirect hitting the local callback
// server once the interactive flow has opened the browser. It runs in a separate
// goroutine and is best-effort: any failure leaves the flow blocked, which the
// caller surfaces via its context timeout.
func fireCallback(callbackURL string) {
	req, err := http.NewRequest(http.MethodGet, callbackURL, nil)
	if err != nil {
		return
	}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	_ = resp.Body.Close()
}

func callbackQueryURL(state string) (string, error) {
	u, err := url.Parse("http://127.0.0.1:8080/callback")
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Add("state", state)
	u.RawQuery = q.Encode()
	return u.String(), nil
}
