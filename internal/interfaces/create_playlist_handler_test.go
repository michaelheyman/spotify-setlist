package interfaces

import (
	"context"
	"testing"

	application "github.com/michaelheyman/spotify-setlist/internal/application/playlist"
	"github.com/michaelheyman/spotify-setlist/internal/application/playlist/mocks"
	"github.com/michaelheyman/spotify-setlist/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_createPlaylistHandler_CreatePlaylist(t *testing.T) {
	testdata := struct {
		request         CreatePlaylistRequest
		createdPlaylist application.CreatePlaylistResult
	}{
		request: CreatePlaylistRequest{
			Artist:   "The Beatles",
			MinSongs: 10,
		},
		createdPlaylist: application.CreatePlaylistResult{
			Playlist: domain.Playlist{},
			Setlist:  domain.Setlist{},
		},
	}
	tests := []struct {
		name            string
		req             CreatePlaylistRequest
		setExpectations func(ps *mocks.PlaylistService)
		want            *CreatePlaylistResponse
		wantErr         assert.ErrorAssertionFunc
	}{
		{
			name: "should create playlist",
			req:  testdata.request,
			setExpectations: func(ps *mocks.PlaylistService) {
				ps.EXPECT().
					CreatePlaylist(
						mock.Anything,
						application.CreatePlaylistParams{
							Artist:                  testdata.request.Artist,
							MinimumSetlistSongCount: testdata.request.MinSongs,
						},
					).
					Return(testdata.createdPlaylist, nil)
			},
			want: &CreatePlaylistResponse{
				Playlist:     testdata.createdPlaylist.Playlist,
				MissingSongs: testdata.createdPlaylist.MissingSongs,
				Setlist:      testdata.createdPlaylist.Setlist,
			},
			wantErr: assert.NoError,
		},
		{
			name: "should return error when create playlist fails",
			req:  testdata.request,
			setExpectations: func(ps *mocks.PlaylistService) {
				ps.EXPECT().
					CreatePlaylist(
						mock.Anything,
						application.CreatePlaylistParams{
							Artist:                  testdata.request.Artist,
							MinimumSetlistSongCount: testdata.request.MinSongs,
						},
					).
					Return(application.CreatePlaylistResult{}, assert.AnError)
			},
			wantErr: func(_ assert.TestingT, err error, _ ...any) bool {
				require.ErrorIs(t, err, assert.AnError)
				assert.ErrorContains(t, err, "creating playlist:")
				return true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &mocks.PlaylistService{}
			tt.setExpectations(ps)
			defer ps.AssertExpectations(t)

			handler := NewCreatePlaylistHandler(ps)
			result, err := handler.CreatePlaylist(context.Background(), tt.req)

			tt.wantErr(t, err, "CreatePlaylist returned error")
			assert.Equal(t, tt.want, result)
		})
	}
}
