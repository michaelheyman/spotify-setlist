package infrastructure

import (
	"context"
	"fmt"

	"github.com/michaelheyman/spotify-setlist/internal/domain"
	"github.com/michaelheyman/spotify-setlist/internal/infrastructure/externalsdk"
	spotify "github.com/zmb3/spotify/v2"
)

type SpotifyClient struct {
	client externalsdk.SpotifyClient
}

func NewSpotifyClient(client externalsdk.SpotifyClient) *SpotifyClient {
	return &SpotifyClient{
		client: client,
	}
}

func (c SpotifyClient) CreatePlaylistForUser(ctx context.Context, user string, tracks []domain.SpotifyTrack, playlistName, description string, public bool, collaborative bool) error {
	newPlaylist, err := c.client.CreatePlaylistForUser(
		ctx,
		user,
		playlistName,
		description,
		public,
		collaborative,
	)
	if err != nil {
		return fmt.Errorf("creating playlist for user: %w", err)
	}

	trackIDs := make([]spotify.ID, len(tracks))
	for i, track := range tracks {
		trackIDs[i] = spotify.ID(track.ID)
	}

	if _, err := c.client.AddTracksToPlaylist(ctx, newPlaylist.ID, trackIDs...); err != nil {
		return fmt.Errorf("adding tracks to playlist: %w", err)
	}

	return nil
}

func (c SpotifyClient) CreatePlaylist(ctx context.Context, user string, playlist domain.SpotifyPlaylist) error {
	newPlaylist, err := c.client.CreatePlaylistForUser(
		ctx,
		user,
		playlist.Name,
		playlist.Description,
		playlist.Public,
		playlist.Collaborative,
	)
	if err != nil {
		return fmt.Errorf("creating playlist for user: %w", err)
	}

	trackIDs := make([]spotify.ID, len(playlist.Tracks))
	for i, track := range playlist.Tracks {
		trackIDs[i] = spotify.ID(track.ID)
	}

	if _, err := c.client.AddTracksToPlaylist(ctx, newPlaylist.ID, trackIDs...); err != nil {
		return fmt.Errorf("adding tracks to playlist: %w", err)
	}

	return nil
}

func (c SpotifyClient) CurrentUser(ctx context.Context) (string, error) {
	user, err := c.client.CurrentUser(ctx)
	if err != nil {
		return "", fmt.Errorf("getting current user: %w", err)
	}
	return user.ID, nil
}

func (c SpotifyClient) SearchTrack(ctx context.Context, artist, title string) (domain.SpotifySearchResult, error) {
	result, err := c.client.Search(ctx, trackQuery(artist, title), spotify.SearchTypeTrack)
	if err != nil {
		return domain.SpotifySearchResult{}, fmt.Errorf("searching for song '%s': %w", title, err)
	}

	if result.Tracks == nil || len(result.Tracks.Tracks) == 0 {
		return domain.SpotifySearchResult{}, nil
	}

	tracks := make([]domain.SpotifyTrack, len(result.Tracks.Tracks))
	for i, t := range result.Tracks.Tracks {
		tracks[i] = domain.SpotifyTrack{
			Name:       t.Name,
			ID:         string(t.ID),
			Popularity: int(t.Popularity),
		}
	}

	return domain.SpotifySearchResult{
		Tracks: tracks,
	}, nil
}

func trackQuery(artist, title string) string {
	return fmt.Sprintf(`track:"%s" artist:"%s"`, title, artist)
}
