package playlist

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/michaelheyman/spotify-setlist/internal/domain"
	"github.com/michaelheyman/spotify-setlist/internal/domain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_playlistService_CreatePlaylist(t *testing.T) {
	testdata := struct {
		parameters CreatePlaylistParams
		setlists   []domain.Setlist
		setlist    domain.Setlist
	}{
		parameters: CreatePlaylistParams{
			Artist: "artist name",
		},
		setlists: []domain.Setlist{
			{
				Artist:    "artist name",
				Songs:     []string{"foo", "bar", "baz"},
				Venue:     "the venue",
				EventDate: time.Date(2025, time.April, 13, 0, 0, 0, 0, time.UTC),
			},
		},
		setlist: domain.Setlist{
			Artist:    "artist name",
			Songs:     []string{"foo", "bar", "baz"},
			Venue:     "the venue",
			EventDate: time.Date(2025, time.April, 13, 0, 0, 0, 0, time.UTC),
		},
	}

	tests := []struct {
		name            string
		params          CreatePlaylistParams
		setExpectations func(sr *mocks.SetlistRepository, pr *mocks.PlaylistRepository)
		want            CreatePlaylistResult
		wantErr         assert.ErrorAssertionFunc
	}{
		{
			name:   "should create playlist",
			params: testdata.parameters,
			setExpectations: func(sr *mocks.SetlistRepository, pr *mocks.PlaylistRepository) {
				sr.EXPECT().
					GetSetlists(
						mock.Anything,
						testdata.parameters.Artist,
					).
					Return([]domain.Setlist{testdata.setlist}, nil)
				pr.EXPECT().
					CreatePlaylist(
						mock.Anything,
						domain.Playlist{
							Name:        fmt.Sprintf("%s Setlist", testdata.setlist.Artist),
							Description: "at the venue, on 2025-04-13",
							Artist:      testdata.setlist.Artist,
							Songs:       testdata.setlist.Songs,
						},
					).
					Return(
						domain.CreatePlaylistResult{
							Playlist: domain.Playlist{
								Name:        fmt.Sprintf("%s Setlist", testdata.setlist.Artist),
								Description: "at the venue, on 2025-04-13",
								Artist:      testdata.setlist.Artist,
								Songs:       testdata.setlist.Songs,
							},
							AddedSongs: testdata.setlist.Songs,
						},
						nil,
					)
			},
			want: CreatePlaylistResult{
				Playlist: domain.Playlist{
					Name:        fmt.Sprintf("%s Setlist", testdata.setlist.Artist),
					Description: "at the venue, on 2025-04-13",
					Artist:      testdata.setlist.Artist,
					Songs:       testdata.setlist.Songs,
				},
				Setlist: testdata.setlist,
			},
			wantErr: assert.NoError,
		},
		{
			name:            "should return error when playlist creation parameters are invalid",
			params:          CreatePlaylistParams{},
			setExpectations: func(_ *mocks.SetlistRepository, _ *mocks.PlaylistRepository) {},
			wantErr: func(_ assert.TestingT, err error, _ ...any) bool {
				return assert.ErrorContains(t, err, "validating parameters: ")
			},
		},
		{
			name:   "should return error when retrieving setlists fails",
			params: testdata.parameters,
			setExpectations: func(sr *mocks.SetlistRepository, _ *mocks.PlaylistRepository) {
				sr.EXPECT().
					GetSetlists(
						mock.Anything,
						testdata.parameters.Artist,
					).
					Return(nil, assert.AnError)
			},
			wantErr: func(_ assert.TestingT, err error, _ ...any) bool {
				return assert.ErrorContains(t, err, "getting setlists: ")
			},
		},
		{
			// TODO: add more tests for the select best setlist
			name:   "should return error when unable to select best setlist",
			params: testdata.parameters,
			setExpectations: func(sr *mocks.SetlistRepository, _ *mocks.PlaylistRepository) {
				sr.EXPECT().
					GetSetlists(
						mock.Anything,
						testdata.parameters.Artist,
					).
					Return(
						[]domain.Setlist{},
						nil,
					)
			},
			wantErr: func(_ assert.TestingT, err error, _ ...any) bool {
				return assert.EqualError(t, err, "no setlists found matching criteria")
			},
		},
		{
			name:   "should return error when creating playlist fails",
			params: testdata.parameters,
			setExpectations: func(sr *mocks.SetlistRepository, pr *mocks.PlaylistRepository) {
				sr.EXPECT().
					GetSetlists(
						mock.Anything,
						testdata.parameters.Artist,
					).
					Return([]domain.Setlist{testdata.setlist}, nil)
				pr.EXPECT().
					CreatePlaylist(
						mock.Anything,
						domain.Playlist{
							Name:        fmt.Sprintf("%s Setlist", testdata.setlist.Artist),
							Description: "at the venue, on 2025-04-13",
							Artist:      testdata.setlist.Artist,
							Songs:       testdata.setlist.Songs,
						},
					).
					Return(domain.CreatePlaylistResult{}, assert.AnError)
			},
			wantErr: func(_ assert.TestingT, err error, _ ...any) bool {
				require.ErrorIs(t, err, assert.AnError)
				assert.ErrorContains(t, err, "creating playlist: ")
				return true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sr := &mocks.SetlistRepository{}
			pr := &mocks.PlaylistRepository{}
			tt.setExpectations(sr, pr)
			defer sr.AssertExpectations(t)
			defer pr.AssertExpectations(t)

			p := NewPlaylistService(sr, pr)

			got, gotErr := p.CreatePlaylist(context.Background(), tt.params)

			tt.wantErr(t, gotErr, "CreatePlaylist returned error")
			assert.Equal(t, tt.want, got)
		})
	}
}
