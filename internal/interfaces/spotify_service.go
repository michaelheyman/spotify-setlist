package interfaces

import (
	"context"
	"fmt"

	"github.com/michaelheyman/spotify-cli/internal/domain"
	spotify "github.com/zmb3/spotify/v2"
)

const (
	playlistVisibility    = false
	playlistCollaborative = false
)

type SpotifyClient interface {
	AddTracksToPlaylist(ctx context.Context, playlistID spotify.ID, trackIDs ...spotify.ID) (snapshotID string, err error)
	CreatePlaylistForUser(ctx context.Context, userID, playlistName, description string, public bool, collaborative bool) (*spotify.FullPlaylist, error)
	CurrentUser(ctx context.Context) (*spotify.PrivateUser, error)
	// GetTrack(ctx context.Context, id spotify.ID, opts ...spotify.RequestOption) (*spotify.FullTrack, error)
	// GetTracks(ctx context.Context, ids []spotify.ID, opts ...spotify.RequestOption) ([]*spotify.FullTrack, error)
	Search(ctx context.Context, query string, t spotify.SearchType, opts ...spotify.RequestOption) (*spotify.SearchResult, error)
}

type spotifyService struct {
	client SpotifyClient
}

func NewSpotifyService(client SpotifyClient) spotifyService {
	return spotifyService{
		client: client,
	}
}

func (s spotifyService) CreatePlaylist(ctx context.Context, playlist domain.Playlist) error {
	// TODO: add playlist validation

	trackIDs, err := s.getTrackIDs(ctx, playlist.Artist, playlist.Songs)
	if err != nil {
		return fmt.Errorf("getting track IDs: %w", err)
	}
	// TODO: return error if track IDs are empty?

	if err := s.createPlaylist(ctx, playlist, trackIDs); err != nil {
		return fmt.Errorf("creating playlist: %w", err)
	}
	return nil
}

func (s spotifyService) getTrackIDs(ctx context.Context, _ string, songs []string) ([]spotify.ID, error) {
	var trackIDs []spotify.ID
	for _, song := range songs {
		// TODO: figure out how to optimize the query
		query := song
		_, err := s.client.Search(ctx, query, spotify.SearchTypeTrack)
		if err != nil {
			return nil, fmt.Errorf("searching for song '%s': %w", song, err)
		}
		// TODO: figure out how to parse result to get the Track ID
	}

	return trackIDs, nil
}

func (s spotifyService) createPlaylist(ctx context.Context, playlist domain.Playlist, trackIDs []spotify.ID) error {
	user, err := s.client.CurrentUser(ctx)
	if err != nil {
		return fmt.Errorf("getting current user: %w", err)
	}

	newPlaylist, err := s.client.CreatePlaylistForUser(
		ctx,
		user.ID,
		playlist.Name,
		playlist.Description,
		playlistVisibility,
		playlistCollaborative,
	)
	if err != nil {
		return fmt.Errorf("creating: %w", err)
	}

	if _, err := s.client.AddTracksToPlaylist(ctx, newPlaylist.ID, trackIDs...); err != nil {
		return fmt.Errorf("adding tracks: %w", err)
	}

	return nil
}
