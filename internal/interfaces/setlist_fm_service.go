package interfaces

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/michaelheyman/spotify-cli/internal/domain"
)

const (
	apiKeyHeader = "x-api-key"
	baseURL      = "https://api.setlist.fm/rest/1.0"
)

type searchArtistResponse struct {
	Artist []struct {
		Mbid           string `json:"mbid"`
		Name           string `json:"name"`
		SortName       string `json:"sortName"`
		Disambiguation string `json:"disambiguation"`
		URL            string `json:"url"`
	} `json:"artist"`
	Total        int `json:"total"`
	Page         int `json:"page"`
	ItemsPerPage int `json:"itemsPerPage"`
}

type getArtistSetlistsResponse struct {
	Type         string `json:"type"`
	ItemsPerPage int    `json:"itemsPerPage"`
	Page         int    `json:"page"`
	Total        int    `json:"total"`
	Setlist      []struct {
		ID          string `json:"id"`
		VersionID   string `json:"versionId"`
		EventDate   string `json:"eventDate"`
		LastUpdated string `json:"lastUpdated"`
		Artist      struct {
			Mbid           string `json:"mbid"`
			Name           string `json:"name"`
			SortName       string `json:"sortName"`
			Disambiguation string `json:"disambiguation"`
			URL            string `json:"url"`
		} `json:"artist"`
		Venue struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			City struct {
				ID        string `json:"id"`
				Name      string `json:"name"`
				State     string `json:"state"`
				StateCode string `json:"stateCode"`
				Coords    struct {
					Lat  float64 `json:"lat"`
					Long float64 `json:"long"`
				} `json:"coords"`
				Country struct {
					Code string `json:"code"`
					Name string `json:"name"`
				} `json:"country"`
			} `json:"city"`
			URL string `json:"url"`
		} `json:"venue"`
		Sets struct {
			Set []struct {
				Song []struct {
					Name string `json:"name"`
					Info string `json:"info,omitempty"`
					With struct {
						Mbid           string `json:"mbid"`
						Name           string `json:"name"`
						SortName       string `json:"sortName"`
						Disambiguation string `json:"disambiguation"`
						URL            string `json:"url"`
					} `json:"with,omitempty"`
				} `json:"song"`
			} `json:"set"`
		} `json:"sets"`
		URL  string `json:"url"`
		Info string `json:"info,omitempty"`
	} `json:"setlist"`
}

type SetlistFMService struct {
	client  *http.Client
	apiKey  string
	baseURL string
}

func NewSetlistFMService(client *http.Client, apiKey string) *SetlistFMService {
	return &SetlistFMService{
		client:  client,
		apiKey:  apiKey,
		baseURL: baseURL,
	}
}

func (s SetlistFMService) GetSetlists(ctx context.Context, artist string, opts ...domain.GetSetlistOption) ([]domain.Setlist, error) {
	// Get the artist mbid (Musicbrainz MBID) from the Setlist.fm API
	mbid, err := s.getArtistMBID(ctx, artist)
	if err != nil {
		return nil, fmt.Errorf("getting artist mbid: %w", err)
	}

	// Search for the artist setlists with the artist's mbid
	setlists, err := s.getArtistSetlists(ctx, mbid)
	if err != nil {
		return nil, fmt.Errorf("getting artist setlists: %w", err)
	}

	return setlists, nil
}

// getArtistMBID retrieves the Musicbrainz MBID for the given artist from Setlist.fm API
func (s SetlistFMService) getArtistMBID(ctx context.Context, artist string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/search/artists", s.baseURL), nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	q := req.URL.Query()
	q.Add("artistName", artist)
	q.Add("sort", "relevance")
	req.URL.RawQuery = q.Encode()

	req.Header.Set(apiKeyHeader, s.apiKey)
	req.Header.Set("Accept", "application/json")

	res, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("executing request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request failed with status %d", res.StatusCode)
	}

	var searchArtistResponse searchArtistResponse
	if err := json.NewDecoder(res.Body).Decode(&searchArtistResponse); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}

	if len(searchArtistResponse.Artist) < 1 {
		return "", fmt.Errorf("no artist found for name: %s", artist)
	}

	// Return the first artist's MBID since we are getting results sorted by relevance
	return searchArtistResponse.Artist[0].Mbid, nil
}

// getArtistSetlists retrieves the setlists for an artist's MBID
func (s SetlistFMService) getArtistSetlists(ctx context.Context, mbid string) ([]domain.Setlist, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/artist/%s/setlists", s.baseURL, mbid), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set(apiKeyHeader, s.apiKey)

	res, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d", res.StatusCode)
	}

	var artistSetlistsResponse getArtistSetlistsResponse

	if err := json.NewDecoder(res.Body).Decode(&artistSetlistsResponse); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return artistSetlistsResponse.toSetlists(), nil
}

func (asr getArtistSetlistsResponse) toSetlists() []domain.Setlist {
	var setlists []domain.Setlist
	for _, s := range asr.Setlist {
		var songs []string
		for _, set := range s.Sets.Set {
			for _, song := range set.Song {
				songs = append(songs, song.Name)
			}
		}
		setlists = append(setlists, domain.Setlist{
			Artist: s.Artist.Name,
			Venue:  s.Venue.Name,
			Songs:  songs,
		})
	}

	return setlists
}
