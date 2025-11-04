package spotify

import (
	"context"
	"testing"

	"github.com/michaelheyman/spotify-setlist/internal/domain"
	"github.com/michaelheyman/spotify-setlist/internal/domain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_spotifyService_CreatePlaylist(t *testing.T) {
	testdata := struct {
		playlist      domain.Playlist
		spotifyUserID string
		searchResult  domain.SpotifySearchResult
		createResult  domain.CreatePlaylistResult
	}{
		playlist: domain.Playlist{
			Name:        "playlist name",
			Description: "playlist description",
			Artist:      "artist name",
			Songs:       []string{"first", "second", "third"},
		},
		spotifyUserID: "example-user-id",
		createResult: domain.CreatePlaylistResult{
			Playlist: domain.Playlist{
				Name:        "playlist name",
				Description: "playlist description",
				Artist:      "artist name",
				Songs:       []string{"first", "second", "third"},
			},
		},
		searchResult: domain.SpotifySearchResult{
			Tracks: []domain.SpotifyTrack{
				{
					Name:       "first",
					ID:         "track-id",
					Popularity: 50,
				},
				{
					Name:       "first",
					ID:         "track-id-but-most-popular",
					Popularity: 80,
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
					SearchTrack(
						mock.Anything,
						testdata.playlist.Artist,
						mock.Anything,
					).
					Return(testdata.searchResult, nil).
					Times(len(testdata.playlist.Songs))
				c.EXPECT().
					CurrentUser(mock.Anything).
					Return(testdata.spotifyUserID, nil)
				c.EXPECT().
					CreatePlaylist(
						mock.Anything,
						testdata.spotifyUserID,
						domain.SpotifyPlaylist{
							Playlist: testdata.playlist,
							Tracks: []domain.SpotifyTrack{
								{
									Name:       "first",
									ID:         "track-id",
									Popularity: 50,
								},
								{
									Name:       "first",
									ID:         "track-id",
									Popularity: 50,
								},
								{
									Name:       "first",
									ID:         "track-id",
									Popularity: 50,
								},
							},
						},
					).
					Return(nil)
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
					SearchTrack(
						mock.Anything,
						testdata.playlist.Artist,
						mock.Anything,
					).
					Return(testdata.searchResult, nil).
					Times(len(testdata.playlist.Songs))
				c.EXPECT().
					CurrentUser(mock.Anything).
					Return(testdata.spotifyUserID, nil)
				c.EXPECT().
					CreatePlaylist(
						mock.Anything,
						testdata.spotifyUserID,
						domain.SpotifyPlaylist{
							Playlist: testdata.playlist,
							Tracks: []domain.SpotifyTrack{
								{
									Name:       "first",
									ID:         "track-id-but-most-popular",
									Popularity: 80,
								},
								{
									Name:       "first",
									ID:         "track-id-but-most-popular",
									Popularity: 80,
								},
								{
									Name:       "first",
									ID:         "track-id-but-most-popular",
									Popularity: 80,
								},
							},
						},
					).
					Return(nil)
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
					SearchTrack(
						mock.Anything,
						testdata.playlist.Artist,
						mock.Anything,
					).
					Return(testdata.searchResult, nil).
					Times(len(testdata.playlist.Songs))
				c.EXPECT().
					CurrentUser(mock.Anything).
					Return(testdata.spotifyUserID, nil)
				c.EXPECT().
					CreatePlaylist(
						mock.Anything,
						testdata.spotifyUserID,
						domain.SpotifyPlaylist{
							Playlist: testdata.playlist,
							Tracks: []domain.SpotifyTrack{
								{
									Name:       "first",
									ID:         "track-id",
									Popularity: 50,
								},
								{
									Name:       "first",
									ID:         "track-id",
									Popularity: 50,
								},
								{
									Name:       "first",
									ID:         "track-id",
									Popularity: 50,
								},
							},
							Public:        true,
							Collaborative: true,
						},
					).
					Return(nil)
			},
			want:    testdata.createResult,
			wantErr: assert.NoError,
		},
		{
			name:     "should return error when spotify search for track fails",
			playlist: testdata.playlist,
			setExpectations: func(c *mocks.SpotifyClient) {
				c.EXPECT().
					SearchTrack(
						mock.Anything,
						testdata.playlist.Artist,
						mock.Anything,
					).
					Return(domain.SpotifySearchResult{}, assert.AnError)
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
				c.EXPECT().
					SearchTrack(
						mock.Anything,
						testdata.playlist.Artist,
						mock.Anything,
					).
					Return(domain.SpotifySearchResult{}, nil)
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
					SearchTrack(
						mock.Anything,
						testdata.playlist.Artist,
						mock.Anything,
					).
					Return(testdata.searchResult, nil).
					Times(len(testdata.playlist.Songs))
				c.EXPECT().
					CurrentUser(mock.Anything).
					Return("", assert.AnError)
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
					SearchTrack(
						mock.Anything,
						testdata.playlist.Artist,
						mock.Anything,
					).
					Return(testdata.searchResult, nil).
					Times(len(testdata.playlist.Songs))
				c.EXPECT().
					CurrentUser(mock.Anything).
					Return(testdata.spotifyUserID, nil)
				c.EXPECT().
					CreatePlaylist(
						mock.Anything,
						testdata.spotifyUserID,
						domain.SpotifyPlaylist{
							Playlist: testdata.playlist,
							Tracks: []domain.SpotifyTrack{
								{
									Name:       "first",
									ID:         "track-id",
									Popularity: 50,
								},
								{
									Name:       "first",
									ID:         "track-id",
									Popularity: 50,
								},
								{
									Name:       "first",
									ID:         "track-id",
									Popularity: 50,
								},
							},
						},
					).
					Return(assert.AnError)
			},
			wantErr: func(tt assert.TestingT, err error, _ ...any) bool {
				return assert.ErrorContains(t, err, "creating playlist: spotify client: ") &&
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
