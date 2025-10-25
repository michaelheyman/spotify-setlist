package domain

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type PlaylistRepository interface {
	CreatePlaylist(ctx context.Context, playlist Playlist, opts ...CreatePlaylistOption) (CreatePlaylistResult, error)
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

type CreatePlaylistOptions struct {
	// Visibility affects the visibility of the created playlist.
	Visibility bool
	// Collaborative determines whether the created playlist will allow collaborators.
	Collaborative bool
	// MostPopularTrack will disambiguate tracks added to the playlist by the ones that are most popular.
	MostPopularTrack bool
}

type CreatePlaylistOption func(*CreatePlaylistOptions)

// WithPlaylistVisibility sets the playlist to publicly visible.
func WithPlaylistVisibility() CreatePlaylistOption {
	return func(o *CreatePlaylistOptions) {
		o.Visibility = true
	}
}

// WithPlaylistCollaborative sets the playlist to collaborative.
func WithPlaylistCollaborative() CreatePlaylistOption {
	return func(o *CreatePlaylistOptions) {
		o.Collaborative = true
	}
}

// WithPlaylistMostPopularTrack will populate the playlist with the most popular instance of each track.
func WithPlaylistMostPopularTrack() CreatePlaylistOption {
	return func(o *CreatePlaylistOptions) {
		o.MostPopularTrack = true
	}
}
