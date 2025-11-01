package interfaces

import (
	"context"
	"testing"

	"github.com/michaelheyman/spotify-setlist/internal/application/mocks"
	"github.com/michaelheyman/spotify-setlist/internal/domain"
	domainmocks "github.com/michaelheyman/spotify-setlist/internal/domain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_spotifyAuthHandler_Authenticate(t *testing.T) {
	tests := []struct {
		name            string
		setExpectations func(s *mocks.SpotifyAuthService)
		want            domain.SpotifyClient
		wantErr         assert.ErrorAssertionFunc
	}{
		{
			name: "should authenticate and return client",
			setExpectations: func(s *mocks.SpotifyAuthService) {
				s.EXPECT().
					StartAuthFlow(mock.Anything).
					Return("some-auth-url", nil)
				s.EXPECT().
					WaitForClient(mock.Anything).
					Return(&domainmocks.SpotifyClient{}, nil)
			},
			want:    &domainmocks.SpotifyClient{},
			wantErr: assert.NoError,
		},
		{
			name: "should return error when start auth flow fails",
			setExpectations: func(s *mocks.SpotifyAuthService) {
				s.EXPECT().
					StartAuthFlow(mock.Anything).
					Return("", assert.AnError)
			},
			wantErr: func(tt assert.TestingT, err error, _ ...any) bool {
				assert.ErrorIs(t, err, assert.AnError)
				assert.ErrorContains(t, err, "starting auth flow:")
				return true
			},
		},
		{
			name: "should return error when wait for client fails",
			setExpectations: func(s *mocks.SpotifyAuthService) {
				s.EXPECT().
					StartAuthFlow(mock.Anything).
					Return("some-auth-url", nil)
				s.EXPECT().
					WaitForClient(mock.Anything).
					Return(nil, assert.AnError)
			},
			wantErr: func(tt assert.TestingT, err error, _ ...any) bool {
				assert.ErrorIs(t, err, assert.AnError)
				assert.ErrorContains(t, err, "waiting for client:")
				return true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			as := &mocks.SpotifyAuthService{}
			tt.setExpectations(as)
			defer as.AssertExpectations(t)

			s := NewSpotifyAuthHandler(as)
			client, err := s.Authenticate(context.Background())

			tt.wantErr(t, err, "Authenticate returned error")

			assert.Equal(t, tt.want, client)
		})
	}
}
