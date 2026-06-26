package spotify

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/michaelheyman/spotify-setlist/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileTokenStore_TokenStore(t *testing.T) {
	t.TempDir()
	tests := []struct {
		name    string
		token   *domain.SpotifyToken
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "should save and load token",
			token: &domain.SpotifyToken{
				AccessToken:  "access-token",
				TokenType:    "Bearer",
				RefreshToken: "refresh-token",
				Expiry:       time.Date(2025, time.November, 4, 15, 15, 53, 235880000, time.Local),
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use an ephemeral directory and bypass permissions set in constructor
			s := &FileTokenStore{
				path: filepath.Join(t.TempDir(), tokenFile),
			}

			err := s.SaveToken(context.Background(), tt.token)

			tt.wantErr(t, err, "SaveToken returned error")

			token, found, err := s.LoadToken(context.Background())

			tt.wantErr(t, err, "LoadToken returned error")
			assert.Equal(t, tt.token, token)
			assert.True(t, found)
		})
	}
}

func TestFileTokenStore_DeleteToken(t *testing.T) {
	tests := []struct {
		name      string
		seedToken *domain.SpotifyToken
		wantErr   assert.ErrorAssertionFunc
	}{
		{
			name:      "should delete the token when one exists",
			seedToken: &domain.SpotifyToken{AccessToken: "access-token"},
			wantErr:   assert.NoError,
		},
		{
			name:    "should be a no-op when the token is already absent",
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &FileTokenStore{
				path: filepath.Join(t.TempDir(), tokenFile),
			}
			if tt.seedToken != nil {
				require.NoError(t, s.SaveToken(context.Background(), tt.seedToken))
			}

			err := s.DeleteToken(context.Background())

			tt.wantErr(t, err, "DeleteToken returned error")

			_, found, err := s.LoadToken(context.Background())
			require.NoError(t, err)
			assert.False(t, found)
		})
	}
}
