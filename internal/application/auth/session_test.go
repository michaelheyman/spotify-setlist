package auth

import (
	"context"
	"testing"

	"github.com/michaelheyman/spotify-setlist/internal/domain"
	"github.com/michaelheyman/spotify-setlist/internal/domain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_sessionService_AuthenticatedClient(t *testing.T) {
	token := &domain.SpotifyToken{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
	}

	tests := []struct {
		name            string
		setExpectations func(s *mocks.TokenStore, p *mocks.SpotifyAuthProvider, c *mocks.SpotifyClient)
		wantErr         assert.ErrorAssertionFunc
	}{
		{
			name: "should reuse the stored token when it is still valid",
			setExpectations: func(s *mocks.TokenStore, p *mocks.SpotifyAuthProvider, c *mocks.SpotifyClient) {
				s.EXPECT().LoadToken(mock.Anything).Return(token, true, nil)
				p.EXPECT().ClientFromToken(mock.Anything, token).Return(c)
				c.EXPECT().CurrentUser(mock.Anything).Return("user", nil)
			},
			wantErr: assert.NoError,
		},
		{
			name: "should authenticate and persist the token when none is stored",
			setExpectations: func(s *mocks.TokenStore, p *mocks.SpotifyAuthProvider, c *mocks.SpotifyClient) {
				s.EXPECT().LoadToken(mock.Anything).Return(nil, false, nil)
				p.EXPECT().Authenticate(mock.Anything).Return(token, nil)
				s.EXPECT().SaveToken(mock.Anything, token).Return(nil)
				p.EXPECT().ClientFromToken(mock.Anything, token).Return(c)
			},
			wantErr: assert.NoError,
		},
		{
			name: "should discard the token and re-authenticate when the session has expired",
			setExpectations: func(s *mocks.TokenStore, p *mocks.SpotifyAuthProvider, c *mocks.SpotifyClient) {
				s.EXPECT().LoadToken(mock.Anything).Return(token, true, nil)
				p.EXPECT().ClientFromToken(mock.Anything, token).Return(c)
				c.EXPECT().CurrentUser(mock.Anything).Return("", domain.ErrSessionExpired)
				s.EXPECT().DeleteToken(mock.Anything).Return(nil)
				p.EXPECT().Authenticate(mock.Anything).Return(token, nil)
				s.EXPECT().SaveToken(mock.Anything, token).Return(nil)
			},
			wantErr: assert.NoError,
		},
		{
			name: "should return error when loading the token fails",
			setExpectations: func(s *mocks.TokenStore, _ *mocks.SpotifyAuthProvider, _ *mocks.SpotifyClient) {
				s.EXPECT().LoadToken(mock.Anything).Return(nil, false, assert.AnError)
			},
			wantErr: func(_ assert.TestingT, err error, _ ...any) bool {
				require.ErrorIs(t, err, assert.AnError)
				assert.ErrorContains(t, err, "loading token:")
				return true
			},
		},
		{
			name: "should return error when discarding the expired token fails",
			setExpectations: func(s *mocks.TokenStore, p *mocks.SpotifyAuthProvider, c *mocks.SpotifyClient) {
				s.EXPECT().LoadToken(mock.Anything).Return(token, true, nil)
				p.EXPECT().ClientFromToken(mock.Anything, token).Return(c)
				c.EXPECT().CurrentUser(mock.Anything).Return("", domain.ErrSessionExpired)
				s.EXPECT().DeleteToken(mock.Anything).Return(assert.AnError)
			},
			wantErr: func(_ assert.TestingT, err error, _ ...any) bool {
				require.ErrorIs(t, err, assert.AnError)
				assert.ErrorContains(t, err, "discarding expired token:")
				return true
			},
		},
		{
			name: "should return error when verifying the session fails for a non-expiry reason",
			setExpectations: func(s *mocks.TokenStore, p *mocks.SpotifyAuthProvider, c *mocks.SpotifyClient) {
				s.EXPECT().LoadToken(mock.Anything).Return(token, true, nil)
				p.EXPECT().ClientFromToken(mock.Anything, token).Return(c)
				c.EXPECT().CurrentUser(mock.Anything).Return("", assert.AnError)
			},
			wantErr: func(_ assert.TestingT, err error, _ ...any) bool {
				require.ErrorIs(t, err, assert.AnError)
				assert.ErrorContains(t, err, "verifying session:")
				return true
			},
		},
		{
			name: "should return error when authentication fails",
			setExpectations: func(s *mocks.TokenStore, p *mocks.SpotifyAuthProvider, _ *mocks.SpotifyClient) {
				s.EXPECT().LoadToken(mock.Anything).Return(nil, false, nil)
				p.EXPECT().Authenticate(mock.Anything).Return(nil, assert.AnError)
			},
			wantErr: func(_ assert.TestingT, err error, _ ...any) bool {
				require.ErrorIs(t, err, assert.AnError)
				assert.ErrorContains(t, err, "authenticating:")
				return true
			},
		},
		{
			name: "should return error when saving the token fails",
			setExpectations: func(s *mocks.TokenStore, p *mocks.SpotifyAuthProvider, _ *mocks.SpotifyClient) {
				s.EXPECT().LoadToken(mock.Anything).Return(nil, false, nil)
				p.EXPECT().Authenticate(mock.Anything).Return(token, nil)
				s.EXPECT().SaveToken(mock.Anything, token).Return(assert.AnError)
			},
			wantErr: func(_ assert.TestingT, err error, _ ...any) bool {
				require.ErrorIs(t, err, assert.AnError)
				assert.ErrorContains(t, err, "saving token:")
				return true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := mocks.NewTokenStore(t)
			p := mocks.NewSpotifyAuthProvider(t)
			c := mocks.NewSpotifyClient(t)
			tt.setExpectations(s, p, c)

			service := NewSessionService(s, p)
			client, err := service.AuthenticatedClient(context.Background())

			tt.wantErr(t, err)
			if err == nil {
				assert.NotNil(t, client)
			}
		})
	}
}
