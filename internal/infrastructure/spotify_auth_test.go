package infrastructure_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/michaelheyman/spotify-setlist/internal/infrastructure"
	"github.com/michaelheyman/spotify-setlist/internal/infrastructure/externalsdk/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2"
)

func TestSpotifyAuth_AuthURL(t *testing.T) {
	state := "abc123"
	authURL := "https://accounts.spotify.com/authorize?client_id=11111111111111111111111111111111&redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Fcallback&response_type=code&scope=playlist-modify-private+playlist-read-private+user-read-private+user-read-email&state=abc123"

	tests := []struct {
		name            string
		setExpectations func(a *mocks.SpotifyAuthenticator)
		state           string
		want            string
	}{
		{
			name:  "should return auth url",
			state: state,
			setExpectations: func(a *mocks.SpotifyAuthenticator) {
				a.EXPECT().AuthURL(state).Return(authURL)
			},
			want: authURL,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &mocks.SpotifyAuthenticator{}
			tt.setExpectations(a)
			defer a.AssertExpectations(t)

			auth := infrastructure.NewSpotifyAuth(a)
			url := auth.AuthURL(tt.state)

			assert.Equal(t, tt.want, url)
		})
	}
}

func TestSpotifyAuth_Token(t *testing.T) {
	state := "abc123"
	req, _ := http.NewRequest(http.MethodGet, "some-url", nil)
	token := &oauth2.Token{
		AccessToken:  "some-access-token",
		RefreshToken: "some-refresh-token",
	}

	tests := []struct {
		name            string
		state           string
		req             *http.Request
		setExpectations func(a *mocks.SpotifyAuthenticator)
		want            *oauth2.Token
		wantErr         assert.ErrorAssertionFunc
	}{
		{
			name:  "should return token",
			state: state,
			req:   req,
			setExpectations: func(a *mocks.SpotifyAuthenticator) {
				a.EXPECT().
					Token(
						mock.Anything,
						state,
						req,
					).
					Return(token, nil)
			},
			want:    token,
			wantErr: assert.NoError,
		},
		{
			name:  "should return error when token returns error",
			state: state,
			req:   req,
			setExpectations: func(a *mocks.SpotifyAuthenticator) {
				a.EXPECT().
					Token(
						mock.Anything,
						state,
						req,
					).
					Return(nil, assert.AnError)
			},
			wantErr: func(tt assert.TestingT, err error, _ ...any) bool {
				assert.ErrorIs(t, err, assert.AnError)
				assert.EqualError(t, err, assert.AnError.Error())
				return true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &mocks.SpotifyAuthenticator{}
			tt.setExpectations(a)
			defer a.AssertExpectations(t)

			auth := infrastructure.NewSpotifyAuth(a)
			token, err := auth.Token(context.Background(), tt.state, tt.req)

			tt.wantErr(t, err, "Token returned error")
			assert.Equal(t, tt.want, token)
		})
	}
}

func TestSpotifyAuth_Client(t *testing.T) {
	token := &oauth2.Token{
		AccessToken:  "some-access-token",
		RefreshToken: "some-refresh-token",
	}
	client := &http.Client{}

	tests := []struct {
		name            string
		token           *oauth2.Token
		setExpectations func(a *mocks.SpotifyAuthenticator)
		want            *http.Client
	}{
		{
			name:  "should return token",
			token: token,
			setExpectations: func(a *mocks.SpotifyAuthenticator) {
				a.EXPECT().
					Client(
						mock.Anything,
						token,
					).
					Return(client)
			},
			want: client,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &mocks.SpotifyAuthenticator{}
			tt.setExpectations(a)
			defer a.AssertExpectations(t)

			auth := infrastructure.NewSpotifyAuth(a)
			client := auth.Client(context.Background(), tt.token)

			assert.Equal(t, tt.want, client)
		})
	}
}
