package externalsdk

import (
	"context"
	"net/http"

	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

type SpotifyClient interface {
	AddTracksToPlaylist(ctx context.Context, playlistID spotify.ID, trackIDs ...spotify.ID) (snapshotID string, err error)
	CreatePlaylistForUser(ctx context.Context, userID, playlistName, description string, public bool, collaborative bool) (*spotify.FullPlaylist, error)
	CurrentUser(ctx context.Context) (*spotify.PrivateUser, error)
	Search(ctx context.Context, query string, t spotify.SearchType, opts ...spotify.RequestOption) (*spotify.SearchResult, error)
}

type SpotifyAuthenticator interface {
	AuthURL(state string, opts ...oauth2.AuthCodeOption) string
	Token(ctx context.Context, state string, r *http.Request, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
	Client(ctx context.Context, token *oauth2.Token) *http.Client
}
