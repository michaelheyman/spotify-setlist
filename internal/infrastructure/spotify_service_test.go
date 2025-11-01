package infrastructure

import (
	"context"
	"fmt"
	"testing"

	"github.com/michaelheyman/spotify-setlist/internal/domain"
	"github.com/michaelheyman/spotify-setlist/internal/domain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zmb3/spotify/v2"
)

func Test_spotifyService_CreatePlaylist(t *testing.T) {
	defaultPlaylistVisibility := false
	defaultPlaylistCollaborative := false

	testdata := struct {
		playlist                   domain.Playlist
		spotifyUserID              string
		spotifyPlaylistID          string
		spotifySearchResultSuccess *spotify.SearchResult
		createResult               domain.CreatePlaylistResult
	}{
		playlist: domain.Playlist{
			Name:        "playlist name",
			Description: "playlist description",
			Artist:      "artist name",
			Songs:       []string{"first", "second", "third"},
		},
		spotifyUserID:     "example-user-id",
		spotifyPlaylistID: "37i9dQZF1DWXRqgorJj26U",
		createResult: domain.CreatePlaylistResult{
			Playlist: domain.Playlist{
				Name:        "playlist name",
				Description: "playlist description",
				Artist:      "artist name",
				Songs:       []string{"first", "second", "third"},
			},
		},
		spotifySearchResultSuccess: &spotify.SearchResult{
			Tracks: &spotify.FullTrackPage{
				Tracks: []spotify.FullTrack{
					{
						SimpleTrack: spotify.SimpleTrack{
							Name: "first",
							ID:   "track-id",
						},
						Popularity: 50,
					},
					{
						SimpleTrack: spotify.SimpleTrack{
							Name: "first",
							ID:   "track-id-but-most-popular",
						},
						Popularity: 80,
					},
				},
			},
		},
	}

	tests := []struct {
		name            string
		playlist        domain.Playlist
		options         []domain.CreatePlaylistOption
		setExpectations func(c *mocks.SpotifyClient)
		want            domain.CreatePlaylistResult
		wantErr         assert.ErrorAssertionFunc
	}{
		{
			name:     "should create playlist",
			playlist: testdata.playlist,
			setExpectations: func(c *mocks.SpotifyClient) {
				c.EXPECT().
					Search(
						mock.Anything,
						mock.Anything,
						spotify.SearchType(spotify.SearchTypeTrack),
					).
					Return(testdata.spotifySearchResultSuccess, nil).
					Times(len(testdata.playlist.Songs))
				c.EXPECT().
					CurrentUser(mock.Anything).
					Return(
						&spotify.PrivateUser{
							User: spotify.User{
								ID: testdata.spotifyUserID,
							},
						}, nil,
					)
				c.EXPECT().
					CreatePlaylistForUser(
						mock.Anything,
						testdata.spotifyUserID,
						testdata.playlist.Name,
						testdata.playlist.Description,
						defaultPlaylistVisibility,
						defaultPlaylistCollaborative,
					).
					Return(
						&spotify.FullPlaylist{
							SimplePlaylist: spotify.SimplePlaylist{
								ID: spotify.ID(testdata.spotifyPlaylistID),
							},
						},
						nil,
					)
				c.EXPECT().
					AddTracksToPlaylist(
						mock.Anything,
						spotify.ID(testdata.spotifyPlaylistID),
						mock.Anything,
					).
					Return("some-snapshot-id", nil)
			},
			want:    testdata.createResult,
			wantErr: assert.NoError,
		},
		{
			name:     "should create playlist with most popular tracks",
			playlist: testdata.playlist,
			options: []domain.CreatePlaylistOption{
				domain.WithPlaylistMostPopularTrack(),
			},
			setExpectations: func(c *mocks.SpotifyClient) {
				c.EXPECT().
					Search(
						mock.Anything,
						mock.Anything,
						spotify.SearchType(spotify.SearchTypeTrack),
					).
					Return(testdata.spotifySearchResultSuccess, nil).
					Times(len(testdata.playlist.Songs))
				c.EXPECT().
					CurrentUser(mock.Anything).
					Return(
						&spotify.PrivateUser{
							User: spotify.User{
								ID: testdata.spotifyUserID,
							},
						}, nil,
					)
				c.EXPECT().
					CreatePlaylistForUser(
						mock.Anything,
						testdata.spotifyUserID,
						testdata.playlist.Name,
						testdata.playlist.Description,
						defaultPlaylistVisibility,
						defaultPlaylistCollaborative,
					).
					Return(
						&spotify.FullPlaylist{
							SimplePlaylist: spotify.SimplePlaylist{
								ID: spotify.ID(testdata.spotifyPlaylistID),
							},
						},
						nil,
					)
				c.EXPECT().
					AddTracksToPlaylist(
						mock.Anything,
						spotify.ID(testdata.spotifyPlaylistID),
						[]spotify.ID{
							"track-id-but-most-popular",
							"track-id-but-most-popular",
							"track-id-but-most-popular",
						},
					).
					Return("some-snapshot-id", nil)
			},
			want:    testdata.createResult,
			wantErr: assert.NoError,
		},
		{
			name:     "should create visible and collaborative playlist",
			playlist: testdata.playlist,
			options: []domain.CreatePlaylistOption{
				domain.WithPlaylistCollaborative(),
				domain.WithPlaylistVisibility(),
			},
			setExpectations: func(c *mocks.SpotifyClient) {
				c.EXPECT().
					Search(
						mock.Anything,
						mock.Anything,
						spotify.SearchType(spotify.SearchTypeTrack),
					).
					Return(testdata.spotifySearchResultSuccess, nil).
					Times(len(testdata.playlist.Songs))
				c.EXPECT().
					CurrentUser(mock.Anything).
					Return(
						&spotify.PrivateUser{
							User: spotify.User{
								ID: testdata.spotifyUserID,
							},
						}, nil,
					)
				c.EXPECT().
					CreatePlaylistForUser(
						mock.Anything,
						testdata.spotifyUserID,
						testdata.playlist.Name,
						testdata.playlist.Description,
						true,
						true,
					).
					Return(
						&spotify.FullPlaylist{
							SimplePlaylist: spotify.SimplePlaylist{
								ID: spotify.ID(testdata.spotifyPlaylistID),
							},
						},
						nil,
					)
				c.EXPECT().
					AddTracksToPlaylist(
						mock.Anything,
						spotify.ID(testdata.spotifyPlaylistID),
						mock.Anything,
					).
					Return("some-snapshot-id", nil)
			},
			want:    testdata.createResult,
			wantErr: assert.NoError,
		},
		{
			name:     "should return error when spotify search for track fails",
			playlist: testdata.playlist,
			setExpectations: func(c *mocks.SpotifyClient) {
				query := fmt.Sprintf(`track:"%s" artist:"%s"`, "first", testdata.playlist.Artist)
				c.EXPECT().
					Search(
						mock.Anything,
						query,
						spotify.SearchType(spotify.SearchTypeTrack),
					).
					Return(nil, assert.AnError)
			},
			wantErr: func(tt assert.TestingT, err error, _ ...any) bool {
				return assert.ErrorContains(t, err, "getting track IDs: searching for song 'first'") &&
					assert.ErrorIs(t, err, assert.AnError)
			},
		},
		{
			name: "should return error when spotify search for track returns no tracks",
			playlist: domain.Playlist{
				Name:        "playlist name",
				Description: "playlist description",
				Artist:      "artist name",
				Songs:       []string{"first"},
			},
			setExpectations: func(c *mocks.SpotifyClient) {
				query := fmt.Sprintf(`track:"%s" artist:"%s"`, "first", testdata.playlist.Artist)
				c.EXPECT().
					Search(
						mock.Anything,
						query,
						spotify.SearchType(spotify.SearchTypeTrack),
					).
					Return(&spotify.SearchResult{}, nil)
			},
			wantErr: func(tt assert.TestingT, err error, _ ...any) bool {
				return assert.Equal(t, err.Error(), "getting track IDs: no tracks found")
			},
		},
		{
			name: "should return error when spotify search for track returns empty tracks",
			playlist: domain.Playlist{
				Name:        "playlist name",
				Description: "playlist description",
				Artist:      "artist name",
				Songs:       []string{"first"},
			},
			setExpectations: func(c *mocks.SpotifyClient) {
				query := fmt.Sprintf(`track:"%s" artist:"%s"`, "first", testdata.playlist.Artist)
				c.EXPECT().
					Search(
						mock.Anything,
						query,
						spotify.SearchType(spotify.SearchTypeTrack),
					).
					Return(
						&spotify.SearchResult{
							Tracks: &spotify.FullTrackPage{
								Tracks: []spotify.FullTrack{},
							},
						},
						nil,
					)
			},
			wantErr: func(tt assert.TestingT, err error, _ ...any) bool {
				return assert.Equal(t, err.Error(), "getting track IDs: no tracks found")
			},
		},
		{
			name:     "should return error when spotify current user fails",
			playlist: testdata.playlist,
			setExpectations: func(c *mocks.SpotifyClient) {
				c.EXPECT().
					Search(
						mock.Anything,
						mock.Anything,
						spotify.SearchType(spotify.SearchTypeTrack),
					).
					Return(testdata.spotifySearchResultSuccess, nil).
					Times(len(testdata.playlist.Songs))
				c.EXPECT().
					CurrentUser(mock.Anything).
					Return(nil, assert.AnError)
			},
			wantErr: func(tt assert.TestingT, err error, _ ...any) bool {
				return assert.ErrorContains(t, err, "creating playlist: getting current user: ") &&
					assert.ErrorIs(t, err, assert.AnError)
			},
		},
		{
			name:     "should return error when spotify create playlist fails",
			playlist: testdata.playlist,
			setExpectations: func(c *mocks.SpotifyClient) {
				c.EXPECT().
					Search(
						mock.Anything,
						mock.Anything,
						spotify.SearchType(spotify.SearchTypeTrack),
					).
					Return(testdata.spotifySearchResultSuccess, nil).
					Times(len(testdata.playlist.Songs))
				c.EXPECT().
					CurrentUser(mock.Anything).
					Return(
						&spotify.PrivateUser{
							User: spotify.User{
								ID: testdata.spotifyUserID,
							},
						}, nil,
					)
				c.EXPECT().
					CreatePlaylistForUser(
						mock.Anything,
						testdata.spotifyUserID,
						testdata.playlist.Name,
						testdata.playlist.Description,
						defaultPlaylistVisibility,
						defaultPlaylistCollaborative,
					).
					Return(nil, assert.AnError)
			},
			wantErr: func(tt assert.TestingT, err error, _ ...any) bool {
				return assert.ErrorContains(t, err, "creating playlist: creating: ") &&
					assert.ErrorIs(t, err, assert.AnError)
			},
		},
		{
			name:     "should return error when spotify add tracks to playlist fails",
			playlist: testdata.playlist,
			setExpectations: func(c *mocks.SpotifyClient) {
				c.EXPECT().
					Search(
						mock.Anything,
						mock.Anything,
						spotify.SearchType(spotify.SearchTypeTrack),
					).
					Return(testdata.spotifySearchResultSuccess, nil).
					Times(len(testdata.playlist.Songs))
				c.EXPECT().
					CurrentUser(mock.Anything).
					Return(
						&spotify.PrivateUser{
							User: spotify.User{
								ID: testdata.spotifyUserID,
							},
						}, nil,
					)
				c.EXPECT().
					CreatePlaylistForUser(
						mock.Anything,
						testdata.spotifyUserID,
						testdata.playlist.Name,
						testdata.playlist.Description,
						defaultPlaylistVisibility,
						defaultPlaylistCollaborative,
					).
					Return(
						&spotify.FullPlaylist{
							SimplePlaylist: spotify.SimplePlaylist{
								ID: spotify.ID(testdata.spotifyPlaylistID),
							},
						},
						nil,
					)
				c.EXPECT().
					AddTracksToPlaylist(
						mock.Anything,
						spotify.ID(testdata.spotifyPlaylistID),
						mock.Anything,
					).
					Return("", assert.AnError)
			},
			wantErr: func(tt assert.TestingT, err error, _ ...any) bool {
				return assert.ErrorContains(t, err, "creating playlist: adding tracks: ") &&
					assert.ErrorIs(t, err, assert.AnError)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mocks.SpotifyClient{}
			tt.setExpectations(client)
			defer client.AssertExpectations(t)

			s := NewSpotifyService(client)
			got, err := s.CreatePlaylist(context.Background(), tt.playlist, tt.options...)

			tt.wantErr(t, err, "CreatePlaylist returned error")
			if err == nil {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
