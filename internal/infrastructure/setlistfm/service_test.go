package setlistfm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/michaelheyman/spotify-setlist/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestSetlistFMService_GetSetlists(t *testing.T) {
	testdata := struct {
		artist                string
		apiKey                string
		artistSetlistsSuccess []byte
		searchArtistsSuccess  []byte
	}{
		artist:                "Test Artist",
		apiKey:                "secret_key_8QjS2pZf6XgD7yH9LcV4T3mR0aB1wE5u",
		artistSetlistsSuccess: loadFixture(t, "artist_setlists_success"),
		searchArtistsSuccess:  loadFixture(t, "search_artists_success"),
	}

	tests := []struct {
		name    string
		artist  string
		apiKey  string
		handler http.HandlerFunc
		want    []domain.Setlist
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:   "should return setlists when setlists exist",
			artist: testdata.artist,
			apiKey: testdata.apiKey,
			handler: func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.Path, "/search/artists") {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(testdata.searchArtistsSuccess))
				}
				var res searchArtistResponse
				if err := json.Unmarshal(testdata.searchArtistsSuccess, &res); err != nil {
					panic("test was unable to unmarshal response")
				}
				mbid := res.Artist[0].Mbid
				if strings.Contains(r.URL.Path, fmt.Sprintf("/artist/%s/setlists", mbid)) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(testdata.artistSetlistsSuccess))
				}
			},
			want: []domain.Setlist{
				{
					Artist: "The Midnight",
					Venue:  "Libbey Bowl",
					Songs: []string{
						"America Online",
						"Deep Blue",
						"Monsters",
						"Days of Thunder",
						"Lost Boy",
						"Brooklyn",
						"Dance With Somebody",
						"Prom Night",
						"Jason",
						"Because the Night",
						"Fire in the Sky",
						"Shadows",
						"Vampires",
						"Neon Medusa",
						"Lost and Found",
						"Gloria",
						"Los Angeles",
						"Last Train",
						"Sunset",
					},
					EventDate: time.Date(2020, time.October, 30, 0, 0, 0, 0, time.UTC),
				},
				{
					Artist: "The Midnight",
					Venue:  "Autódromo Hermanos Rodríguez",
					Songs: []string{
						"Days of Thunder",
						"Lost Boy",
						"Vampires",
						"Shadows",
						"Gloria",
						"Sunset",
					},
					EventDate: time.Date(2019, time.November, 17, 0, 0, 0, 0, time.UTC),
				},
			},
			wantErr: assert.NoError,
		},
		{
			name:   "should return error when get artist returns not found",
			artist: testdata.artist,
			apiKey: testdata.apiKey,
			handler: func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.Path, "/search/artists") {
					w.WriteHeader(http.StatusNotFound)
				}
			},
			wantErr: func(tt assert.TestingT, err error, _ ...any) bool {
				return assert.ErrorContains(t, err, "request failed with status 404")
			},
		},
		{
			name:   "should return error when get artist by mbid fails",
			artist: testdata.artist,
			apiKey: testdata.apiKey,
			handler: func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.Path, "/search/artists") {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{}`))
				}
			},
			wantErr: func(tt assert.TestingT, err error, _ ...any) bool {
				return assert.ErrorContains(
					t,
					err,
					fmt.Sprintf("getting artist mbid: no artist found for name: %s", testdata.artist),
				)
			},
		},
		{
			name:   "should return error when get artist setlists returns not found",
			artist: testdata.artist,
			apiKey: testdata.apiKey,
			handler: func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.Path, "/search/artists") {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(testdata.searchArtistsSuccess))
				}
				var res searchArtistResponse
				if err := json.Unmarshal(testdata.searchArtistsSuccess, &res); err != nil {
					panic("test was unable to unmarshal response")
				}
				mbid := res.Artist[0].Mbid
				if strings.Contains(r.URL.Path, fmt.Sprintf("/artist/%s/setlists", mbid)) {
					w.WriteHeader(http.StatusNotFound)
				}
			},
			wantErr: func(tt assert.TestingT, err error, _ ...any) bool {
				return assert.ErrorContains(t, err, "request failed with status 404")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, r.Header.Get(apiKeyHeader), tt.apiKey, "api key header doesn't match submitted api key")
				assert.Equal(t, r.Header.Get("Accept"), "application/json")
				tt.handler(w, r)
			}))
			defer server.Close()
			s := NewSetlistFMService(server.Client(), tt.apiKey, WithBaseURL(server.URL))

			got, gotErr := s.GetSetlists(context.Background(), tt.artist)

			tt.wantErr(t, gotErr, "GetSetlists returned error")

			if gotErr == nil {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func loadFixture(t *testing.T, filename string) []byte {
	t.Helper()

	path := filepath.Join("testdata", fmt.Sprintf("%s.json", filename))

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file path %s", path)
	}

	return data
}
