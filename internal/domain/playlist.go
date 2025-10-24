package domain

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type PlaylistRepository interface {
	CreatePlaylist(ctx context.Context, playlist Playlist) error
}

type Playlist struct {
	Name        string
	Description string
	Artist      string
	Songs       []string
}

func (p Playlist) Validate() error {
	return validation.ValidateStruct(
		&p,
		validation.Field(&p.Artist, validation.Required),
		validation.Field(&p.Songs, validation.Required),
	)
}
