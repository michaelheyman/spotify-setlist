package infrastructure

import (
	"context"
	"errors"
	"fmt"

	"github.com/michaelheyman/spotify-setlist/internal/domain"
	spotify "github.com/zmb3/spotify/v2"
)

type spotifyService struct {
	client domain.SpotifyClient
}

func NewSpotifyService(client domain.SpotifyClient) spotifyService {
	return spotifyService{
		client: client,
	}
}

func (s spotifyService) CreatePlaylist(ctx context.Context, playlist domain.Playlist, opts ...domain.CreatePlaylistOption) (domain.CreatePlaylistResult, error) {
	if err := playlist.Validate(); err != nil {
		return domain.CreatePlaylistResult{}, fmt.Errorf("validating playlist parameter: %w", err)
	}

	var options domain.CreatePlaylistOptions
	for _, opt := range opts {
		opt(&options)
	}

	trackIDs, missingSongs, err := s.findTracks(ctx, playlist.Artist, playlist.Songs, options.MostPopularTrack)
	if err != nil {
		return domain.CreatePlaylistResult{}, fmt.Errorf("getting track IDs: %w", err)
	}

	if err := s.createPlaylist(ctx, playlist, trackIDs, options.Visibility, options.Collaborative); err != nil {
		return domain.CreatePlaylistResult{}, fmt.Errorf("creating playlist: %w", err)
	}
	return domain.CreatePlaylistResult{
		Playlist:     playlist,
		MissingSongs: missingSongs,
	}, nil
}

func (s spotifyService) findTracks(ctx context.Context, artist string, songs []string, mostPopular bool) ([]spotify.ID, []string, error) {
	var trackIDs []spotify.ID
	var missing []string

	for _, song := range songs {
		query := fmt.Sprintf(`track:"%s" artist:"%s"`, song, artist)
		result, err := s.client.Search(ctx, query, spotify.SearchTypeTrack)
		if err != nil {
			return nil, nil, fmt.Errorf("searching for song '%s': %w", song, err)
		}

		if result.Tracks == nil || len(result.Tracks.Tracks) == 0 {
			missing = append(missing, song)
			continue
		}

		trackID, err := extractTrackID(result.Tracks.Tracks, mostPopular)
		if err != nil {
			return nil, nil, fmt.Errorf("extracting track ID from search result: %w", err)
		}
		trackIDs = append(trackIDs, trackID)
	}

	if len(trackIDs) == 0 {
		return nil, nil, errors.New("no tracks found")
	}

	return trackIDs, missing, nil
}

func (s spotifyService) createPlaylist(
	ctx context.Context,
	playlist domain.Playlist,
	trackIDs []spotify.ID,
	visibility bool,
	collaborative bool,
) error {
	user, err := s.client.CurrentUser(ctx)
	if err != nil {
		return fmt.Errorf("getting current user: %w", err)
	}

	newPlaylist, err := s.client.CreatePlaylistForUser(
		ctx,
		user.ID,
		playlist.Name,
		playlist.Description,
		visibility,
		collaborative,
	)
	if err != nil {
		return fmt.Errorf("creating: %w", err)
	}

	if _, err := s.client.AddTracksToPlaylist(ctx, newPlaylist.ID, trackIDs...); err != nil {
		return fmt.Errorf("adding tracks: %w", err)
	}

	return nil
}

func extractTrackID(tracks []spotify.FullTrack, mostPopular bool) (spotify.ID, error) {
	if len(tracks) == 0 {
		return "", errors.New("no tracks found")
	}

	if !mostPopular {
		return tracks[0].ID, nil
	}

	var highestPopularity spotify.Numeric
	var mostPopularIndex int

	for i, track := range tracks {
		if track.Popularity > highestPopularity {
			highestPopularity = track.Popularity
			mostPopularIndex = i
		}
	}

	return tracks[mostPopularIndex].ID, nil
}
