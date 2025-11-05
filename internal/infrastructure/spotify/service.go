package spotify

import (
	"context"
	"errors"
	"fmt"

	"github.com/michaelheyman/spotify-setlist/internal/domain"
)

type spotifyService struct {
	client domain.SpotifyClient
}

func NewSpotifyService(client domain.SpotifyClient) spotifyService {
	return spotifyService{
		client: client,
	}
}

func (s spotifyService) CreatePlaylist(
	ctx context.Context,
	playlist domain.Playlist,
	opts ...domain.CreatePlaylistOption,
) (domain.CreatePlaylistResult, error) {
	if err := playlist.Validate(); err != nil {
		return domain.CreatePlaylistResult{}, fmt.Errorf("validating playlist parameter: %w", err)
	}

	var options domain.CreatePlaylistOptions
	for _, opt := range opts {
		opt(&options)
	}

	tracks, missingSongs, err := s.findTracks(ctx, playlist.Artist, playlist.Songs, options.MostPopularTrack)
	if err != nil {
		return domain.CreatePlaylistResult{}, fmt.Errorf("getting track IDs: %w", err)
	}

	if err := s.createPlaylist(ctx, playlist, tracks, options.Visibility, options.Collaborative); err != nil {
		return domain.CreatePlaylistResult{}, fmt.Errorf("creating playlist: %w", err)
	}
	return domain.CreatePlaylistResult{
		Playlist:     playlist,
		MissingSongs: missingSongs,
	}, nil
}

func (s spotifyService) findTracks(
	ctx context.Context,
	artist string,
	songs []string,
	mostPopular bool,
) ([]domain.SpotifyTrack, []string, error) {
	var tracks []domain.SpotifyTrack
	var missing []string

	for _, song := range songs {
		result, err := s.client.SearchTrack(ctx, artist, song)
		if err != nil {
			return nil, nil, fmt.Errorf("searching for song '%s': %w", song, err)
		}

		if len(result.Tracks) == 0 {
			missing = append(missing, song)
			continue
		}

		track, err := extractTrack(result.Tracks, mostPopular)
		if err != nil {
			return nil, nil, fmt.Errorf("extracting track ID from search result: %w", err)
		}
		tracks = append(tracks, track)
	}

	if len(tracks) == 0 {
		return nil, nil, errors.New("no tracks found")
	}

	return tracks, missing, nil
}

func (s spotifyService) createPlaylist(
	ctx context.Context,
	playlist domain.Playlist,
	tracks []domain.SpotifyTrack,
	visibility bool,
	collaborative bool,
) error {
	userID, err := s.client.CurrentUser(ctx)
	if err != nil {
		return fmt.Errorf("getting current user: %w", err)
	}

	pl := domain.SpotifyPlaylist{
		Playlist:      playlist,
		Tracks:        tracks,
		Public:        visibility,
		Collaborative: collaborative,
	}
	if err := s.client.CreatePlaylist(
		ctx,
		userID,
		pl,
	); err != nil {
		return fmt.Errorf("spotify client: %w", err)
	}

	return nil
}

func extractTrack(tracks []domain.SpotifyTrack, mostPopular bool) (domain.SpotifyTrack, error) {
	if len(tracks) == 0 {
		return domain.SpotifyTrack{}, errors.New("no tracks found")
	}

	if !mostPopular {
		return tracks[0], nil
	}

	var highestPopularity int
	var mostPopularIndex int

	for i, track := range tracks {
		if track.Popularity > highestPopularity {
			highestPopularity = track.Popularity
			mostPopularIndex = i
		}
	}

	return tracks[mostPopularIndex], nil
}
