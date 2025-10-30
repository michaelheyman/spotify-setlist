package playlist

import (
	"context"
	"errors"
	"time"

	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/michaelheyman/spotify-setlist/internal/domain"
)

type CreatePlaylistParams struct {
	Artist                  string
	MinimumSetlistSongCount int
}

func (p CreatePlaylistParams) Validate() error {
	return validation.ValidateStruct(
		&p,
		validation.Field(&p.Artist, validation.Required),
	)
}

type CreatePlaylistResult struct {
	Playlist     domain.Playlist
	MissingSongs []string
	Setlist      domain.Setlist
}

type PlaylistService interface {
	CreatePlaylist(ctx context.Context, params CreatePlaylistParams) (CreatePlaylistResult, error)
}

type playlistService struct {
	setlist  domain.SetlistRepository
	playlist domain.PlaylistRepository
}

func NewPlaylistService(setlist domain.SetlistRepository, playlist domain.PlaylistRepository) playlistService {
	return playlistService{
		setlist:  setlist,
		playlist: playlist,
	}
}

func (p playlistService) CreatePlaylist(ctx context.Context, params CreatePlaylistParams) (CreatePlaylistResult, error) {
	if err := params.Validate(); err != nil {
		return CreatePlaylistResult{}, fmt.Errorf("validating parameters: %w", err)
	}

	setlists, err := p.setlist.GetSetlists(ctx, params.Artist)
	if err != nil {
		return CreatePlaylistResult{}, fmt.Errorf("getting setlists: %w", err)
	}

	setlist := selectBestSetlist(setlists, params.MinimumSetlistSongCount)
	if setlist == nil {
		return CreatePlaylistResult{}, errors.New("no setlists found matching criteria")
	}

	result, err := p.playlist.CreatePlaylist(ctx, domain.Playlist{
		Name:        playlistName(params.Artist),
		Description: playlistDescription(setlist),
		Artist:      setlist.Artist,
		Songs:       setlist.Songs,
	})
	if err != nil {
		return CreatePlaylistResult{}, fmt.Errorf("creating playlist: %w", err)
	}

	return CreatePlaylistResult{
		Playlist:     result.Playlist,
		MissingSongs: result.MissingSongs,
		Setlist:      *setlist,
	}, nil
}

// selectBestSetlist selects the best setlist that fits the selection criteria. By default, the
// most recent playlist is always preferred. If the minSongCount is set, then that will also be
// considered in the selection process, such that the most recent setlist that satisfies the
// minSongCount will be returned.
func selectBestSetlist(setlists []domain.Setlist, minSongCount int) *domain.Setlist {
	if len(setlists) == 0 {
		return nil
	}
	if minSongCount == 0 {
		return &setlists[0]
	}

	var bestSetlist *domain.Setlist
	for _, setlist := range setlists {
		if len(setlist.Songs) < minSongCount {
			continue
		}

		if bestSetlist == nil {
			bestSetlist = &setlist
			continue
		}

		if setlist.EventDate.After(bestSetlist.EventDate) {
			bestSetlist = &setlist
		}
	}

	return bestSetlist
}

func playlistName(artist string) string {
	return fmt.Sprintf("%s Setlist", artist)
}

func playlistDescription(setlist *domain.Setlist) string {
	if setlist == nil {
		return ""
	}
	if setlist.Venue == "" && setlist.EventDate.IsZero() {
		return ""
	}
	if setlist.EventDate.IsZero() {
		return fmt.Sprintf("at %s", setlist.Venue)
	}
	if setlist.Venue == "" {
		return fmt.Sprintf("on %s", setlist.EventDate.Format(time.DateOnly))
	}

	return fmt.Sprintf("at %s, on %s", setlist.Venue, setlist.EventDate.Format(time.DateOnly))
}
