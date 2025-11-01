package interfaces

import (
	"context"
	"fmt"

	application "github.com/michaelheyman/spotify-setlist/internal/application/playlist"
	"github.com/michaelheyman/spotify-setlist/internal/domain"
)

type CreatePlaylistHandler interface {
	CreatePlaylist(ctx context.Context, req CreatePlaylistRequest) (*CreatePlaylistResponse, error)
}

type CreatePlaylistRequest struct {
	Artist   string
	MinSongs int
}

type CreatePlaylistResponse struct {
	Playlist     domain.Playlist
	MissingSongs []string
	Setlist      domain.Setlist
}

type createPlaylistHandler struct {
	playlist application.PlaylistService
}

func NewCreatePlaylistHandler(playlist application.PlaylistService) CreatePlaylistHandler {
	return &createPlaylistHandler{
		playlist: playlist,
	}
}

func (h createPlaylistHandler) CreatePlaylist(ctx context.Context, req CreatePlaylistRequest) (*CreatePlaylistResponse, error) {
	result, err := h.playlist.CreatePlaylist(ctx, application.CreatePlaylistParams{
		Artist:                  req.Artist,
		MinimumSetlistSongCount: req.MinSongs,
	})
	if err != nil {
		return nil, fmt.Errorf("creating playlist: %w", err)
	}

	response := &CreatePlaylistResponse{
		Playlist:     result.Playlist,
		MissingSongs: result.MissingSongs,
		Setlist:      result.Setlist,
	}

	return response, nil
}
