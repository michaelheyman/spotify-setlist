package spotify

import (
	"context"
	"fmt"
	"testing"

	"github.com/michaelheyman/spotify-setlist/internal/domain"
	"github.com/michaelheyman/spotify-setlist/internal/infrastructure/externalsdk/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/zmb3/spotify/v2"
)

func Test_spotifyClient_CreatePlaylistForUser(t *testing.T) {
	testdata := struct {
		playlist          domain.SpotifyPlaylist
		spotifyUserID     string
		spotifyPlaylistID string
	}{
		playlist: domain.SpotifyPlaylist{
			Playlist: domain.Playlist{
				Name:        "playlist name",
				Description: "playlist description",
				Artist:      "artist name",
				Songs:       []string{"first", "second", "third"},
			},
			Tracks: []domain.SpotifyTrack{
				{Name: "first", ID: "id1"},
				{Name: "second", ID: "id2"},
				{Name: "third", ID: "id3"},
			},
			Public:        true,
			Collaborative: true,
		},
		spotifyUserID:     "example-user-id",
		spotifyPlaylistID: "37i9dQZF1DWXRqgorJj26U",
	}

	tests := []struct {
		name            string
		playlist        domain.SpotifyPlaylist
		setExpectations func(c *mocks.SpotifyClient)
		wantErr         assert.ErrorAssertionFunc
	}{
		{
			name:     "should create playlist",
			playlist: testdata.playlist,
			setExpectations: func(c *mocks.SpotifyClient) {
				c.EXPECT().
					CreatePlaylistForUser(
						mock.Anything,
						testdata.spotifyUserID,
						testdata.playlist.Name,
						testdata.playlist.Description,
						testdata.playlist.Public,
						testdata.playlist.Collaborative,
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
			wantErr: assert.NoError,
		},
		{
			name:     "should return error when create playlist for user fails",
			playlist: testdata.playlist,
			setExpectations: func(c *mocks.SpotifyClient) {
				c.EXPECT().
					CreatePlaylistForUser(
						mock.Anything,
						testdata.spotifyUserID,
						testdata.playlist.Name,
						testdata.playlist.Description,
						testdata.playlist.Public,
						testdata.playlist.Collaborative,
					).
					Return(nil, assert.AnError)
			},
			wantErr: func(_ assert.TestingT, err error, _ ...any) bool {
				require.ErrorIs(t, err, assert.AnError)
				assert.ErrorContains(t, err, "creating playlist for user:")
				return true
			},
		},
		{
			name:     "should return error when add tracks to playlist fails",
			playlist: testdata.playlist,
			setExpectations: func(c *mocks.SpotifyClient) {
				c.EXPECT().
					CreatePlaylistForUser(
						mock.Anything,
						testdata.spotifyUserID,
						testdata.playlist.Name,
						testdata.playlist.Description,
						testdata.playlist.Public,
						testdata.playlist.Collaborative,
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
			wantErr: func(_ assert.TestingT, err error, _ ...any) bool {
				require.ErrorIs(t, err, assert.AnError)
				assert.ErrorContains(t, err, "adding tracks to playlist:")
				return true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mocks.SpotifyClient{}
			tt.setExpectations(client)
			defer client.AssertExpectations(t)

			s := NewSpotifyClient(client)
			err := s.CreatePlaylist(context.Background(), testdata.spotifyUserID, tt.playlist)

			tt.wantErr(t, err, "CreatePlaylistForUser returned error")
		})
	}
}

func Test_spotifyClient_SearchTrack(t *testing.T) {
	testdata := struct {
		artist                     string
		title                      string
		spotifySearchResultSuccess *spotify.SearchResult
		searchResult               domain.SpotifySearchResult
	}{
		artist: "The Beatles",
		title:  "Hey Jude",
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
		artist          string
		title           string
		setExpectations func(c *mocks.SpotifyClient)
		want            domain.SpotifySearchResult
		wantErr         assert.ErrorAssertionFunc
	}{
		{
			name:   "should return found tracks",
			artist: testdata.artist,
			title:  testdata.title,
			setExpectations: func(c *mocks.SpotifyClient) {
				c.EXPECT().
					Search(
						mock.Anything,
						trackQuery(testdata.artist, testdata.title),
						spotify.SearchType(spotify.SearchTypeTrack),
					).
					Return(testdata.spotifySearchResultSuccess, nil)
			},
			want:    testdata.searchResult,
			wantErr: assert.NoError,
		},
		{
			name:   "should return error when search fails",
			artist: testdata.artist,
			title:  testdata.title,
			setExpectations: func(c *mocks.SpotifyClient) {
				c.EXPECT().
					Search(
						mock.Anything,
						trackQuery(testdata.artist, testdata.title),
						spotify.SearchType(spotify.SearchTypeTrack),
					).
					Return(nil, assert.AnError)
			},
			wantErr: func(_ assert.TestingT, err error, _ ...any) bool {
				require.ErrorIs(t, err, assert.AnError)
				assert.ErrorContains(t, err, fmt.Sprintf("searching for song '%s':", testdata.title))
				return true
			},
		},
		{
			name:   "should return empty result when search returns no tracks",
			artist: testdata.artist,
			title:  testdata.title,
			setExpectations: func(c *mocks.SpotifyClient) {
				c.EXPECT().
					Search(
						mock.Anything,
						trackQuery(testdata.artist, testdata.title),
						spotify.SearchType(spotify.SearchTypeTrack),
					).
					Return(&spotify.SearchResult{}, nil)
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mocks.SpotifyClient{}
			tt.setExpectations(client)
			defer client.AssertExpectations(t)

			s := NewSpotifyClient(client)
			result, err := s.SearchTrack(context.Background(), tt.artist, tt.title)

			tt.wantErr(t, err, "SearchTrack returned error")
			assert.Equal(t, tt.want, result)
		})
	}
}

func Test_spotifyClient_CurrentUser(t *testing.T) {
	testdata := struct {
		spotifyUserID string
	}{
		spotifyUserID: "example-user-id",
	}

	tests := []struct {
		name            string
		setExpectations func(c *mocks.SpotifyClient)
		want            string
		wantErr         assert.ErrorAssertionFunc
	}{
		{
			name: "should return current user",
			setExpectations: func(c *mocks.SpotifyClient) {
				c.EXPECT().
					CurrentUser(mock.Anything).
					Return(
						&spotify.PrivateUser{
							User: spotify.User{
								ID: testdata.spotifyUserID,
							},
						},
						nil,
					)
			},
			want:    testdata.spotifyUserID,
			wantErr: assert.NoError,
		},
		{
			name: "should return error when current user fails",
			setExpectations: func(c *mocks.SpotifyClient) {
				c.EXPECT().
					CurrentUser(mock.Anything).
					Return(nil, assert.AnError)
			},
			wantErr: func(_ assert.TestingT, err error, _ ...any) bool {
				require.ErrorIs(t, err, assert.AnError)
				assert.ErrorContains(t, err, "getting current user:")
				return true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mocks.SpotifyClient{}
			tt.setExpectations(client)
			defer client.AssertExpectations(t)

			s := NewSpotifyClient(client)
			user, err := s.CurrentUser(context.Background())

			tt.wantErr(t, err, "CurrentUser returned error")
			assert.Equal(t, tt.want, user)
		})
	}
}
