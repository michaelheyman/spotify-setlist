package domain

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type PlaylistRepository interface {
	CreatePlaylist(ctx context.Context, playlist Playlist) (CreatePlaylistResult, error)
}

type Playlist struct {
	Name        string
	Description string
	Artist      string
	Songs       []string
}

type CreatePlaylistResult struct {
	Playlist     Playlist
	AddedSongs   []string
	MissingSongs []string
}

func (p Playlist) Validate() error {
	return validation.ValidateStruct(
		&p,
		validation.Field(&p.Artist, validation.Required),
		validation.Field(&p.Songs, validation.Required),
	)
}
