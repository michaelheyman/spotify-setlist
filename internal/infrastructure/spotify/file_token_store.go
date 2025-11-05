package spotify

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/michaelheyman/spotify-setlist/internal/domain"
)

const (
	directory = ".spotify-setlist"
	tokenFile = "token.json"
	userRWX   = 0700
	userRW    = 0600
)

type FileTokenStore struct {
	path string
}

func NewFileTokenStore() (*FileTokenStore, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("getting user home directory: %w", err)
	}

	configDir := filepath.Join(home, directory)

	if err := os.MkdirAll(configDir, userRWX); err != nil {
		return nil, fmt.Errorf("getting user home directory: %w", err)
	}

	return &FileTokenStore{
		path: filepath.Join(configDir, tokenFile),
	}, nil
}

func (s FileTokenStore) SaveToken(ctx context.Context, token *domain.SpotifyToken) error {
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("marshaling token: %w", err)
	}

	if err := os.WriteFile(s.path, data, userRW); err != nil {
		return fmt.Errorf("writing token file: %w", err)
	}
	return nil
}

func (s FileTokenStore) LoadToken(ctx context.Context) (*domain.SpotifyToken, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			// Token not found
			return nil, nil
		}
		return nil, fmt.Errorf("reading token file: %w", err)
	}

	var token domain.SpotifyToken
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("unmarshaling token: %w", err)
	}
	return &token, nil
}
