package cmd

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/michaelheyman/spotify-setlist/internal/domain"
	"github.com/michaelheyman/spotify-setlist/internal/domain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

func TestCreatePlaylistCmd_Execute(t *testing.T) {
	testdata := struct {
		clientID     string
		clientSecret string
		redirectURI  string
		apiKey       string
	}{
		clientID:     "11111111111111111111111111111111",
		clientSecret: "ffffffffffffffffffffffffffffffff",
		redirectURI:  "http://localhost:8080/callback",
		apiKey:       "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	validEnv := map[string]string{
		"SPOTIFY_CLIENT_ID":     testdata.clientID,
		"SPOTIFY_CLIENT_SECRET": testdata.clientSecret,
		"SPOTIFY_REDIRECT_URI":  testdata.redirectURI,
		"SETLIST_FM_API_KEY":    testdata.apiKey,
	}

	tests := []struct {
		name            string
		args            []string
		env             map[string]string
		setExpectations func(f *mocks.AuthenticationFactory)
		wantOutput      string
		wantErr         assert.ErrorAssertionFunc
	}{
		{
			name: "should return error when missing Spotify client ID",
			args: []string{"create-playlist", "--artist", "The Beatles"},
			env: map[string]string{
				"SPOTIFY_CLIENT_SECRET": testdata.clientSecret,
				"SPOTIFY_REDIRECT_URI":  testdata.redirectURI,
				"SETLIST_FM_API_KEY":    testdata.apiKey,
			},
			setExpectations: func(_ *mocks.AuthenticationFactory) {},
			wantErr: func(tt assert.TestingT, err error, _ ...any) bool {
				assert.EqualError(t, err, "missing Spotify client ID")
				return true
			},
		},
		{
			name: "should return error when missing Spotify client secret",
			args: []string{"create-playlist", "--artist", "The Beatles"},
			env: map[string]string{
				"SPOTIFY_CLIENT_ID":    testdata.clientID,
				"SPOTIFY_REDIRECT_URI": testdata.redirectURI,
				"SETLIST_FM_API_KEY":   testdata.apiKey,
			},
			setExpectations: func(_ *mocks.AuthenticationFactory) {},
			wantErr: func(tt assert.TestingT, err error, _ ...any) bool {
				assert.EqualError(t, err, "missing Spotify client secret")
				return true
			},
		},
		{
			name: "should return error when missing Setlist.FM API key",
			args: []string{"create-playlist", "--artist", "The Beatles"},
			env: map[string]string{
				"SPOTIFY_CLIENT_ID":     testdata.clientID,
				"SPOTIFY_CLIENT_SECRET": testdata.clientSecret,
				"SPOTIFY_REDIRECT_URI":  testdata.redirectURI,
			},
			setExpectations: func(_ *mocks.AuthenticationFactory) {},
			wantErr: func(tt assert.TestingT, err error, _ ...any) bool {
				assert.EqualError(t, err, "missing Setlist.FM API key")
				return true
			},
		},
		{
			name: "should return error when authenticator fails to create",
			args: []string{"create-playlist", "--artist", "The Beatles"},
			env:  validEnv,
			setExpectations: func(f *mocks.AuthenticationFactory) {
				f.EXPECT().
					CreateAuthenticator(
						mock.Anything,
						domain.SpotifyAuthConfig{
							ClientID:     testdata.clientID,
							ClientSecret: testdata.clientSecret,
							RedirectURI:  testdata.redirectURI,
							Scopes: []string{
								spotifyauth.ScopePlaylistModifyPrivate,
								spotifyauth.ScopePlaylistReadPrivate,
								spotifyauth.ScopeUserReadPrivate,
								spotifyauth.ScopeUserReadEmail,
							},
						},
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
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			f := &mocks.AuthenticationFactory{}
			tt.setExpectations(f)
			defer f.AssertExpectations(t)
			b := bytes.NewBufferString("")

			cmd := NewCreatePlaylistCmd(f, http.DefaultClient)
			cmd.SetArgs(tt.args)
			cmd.SetOut(b)
			cmd.SilenceUsage = true

			err := cmd.Execute()

			tt.wantErr(t, err, "Execute returned error")

			out, err := io.ReadAll(b)
			require.NoError(t, err)
			assert.Equal(t, tt.wantOutput, string(out))
		})
	}
}
