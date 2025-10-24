package domain

import (
	"context"
)

type Playlist struct {
	Name        string
	Description string
	Artist      string
	Songs       []string
}

type PlaylistRepository interface {
	CreatePlaylist(ctx context.Context, playlist Playlist) error
}
