package interfaces

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/michaelheyman/spotify-cli/internal/domain"
)

const (
	apiKeyHeader = "x-api-key"
	baseURL      = "https://api.setlist.fm/rest/1.0"
)

type setlistFMService struct {
	client  *http.Client
	apiKey  string
	baseURL string
}

func NewSetlistFMService(client *http.Client, apiKey string) *setlistFMService {
	return &setlistFMService{
		client:  client,
		apiKey:  apiKey,
		baseURL: baseURL,
	}
}

func (s setlistFMService) GetSetlists(ctx context.Context, artist string, opts ...domain.GetSetlistOption) ([]domain.Setlist, error) {
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
func (s setlistFMService) getArtistMBID(ctx context.Context, artist string) (string, error) {
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
func (s setlistFMService) getArtistSetlists(ctx context.Context, mbid string) ([]domain.Setlist, error) {
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

	setlists, err := artistSetlistsResponse.toSetlists()
	if err != nil {
		return nil, fmt.Errorf("converting response to setlists: %w", err)
	}

	return setlists, nil
}

func (asr getArtistSetlistsResponse) toSetlists() ([]domain.Setlist, error) {
	var setlists []domain.Setlist
	for _, s := range asr.Setlist {
		var songs []string
		var setlist domain.Setlist
		for _, set := range s.Sets.Set {
			for _, song := range set.Song {
				songs = append(songs, song.Name)
			}
		}

		setlist = domain.Setlist{
			Artist: s.Artist.Name,
			Venue:  s.Venue.Name,
			Songs:  songs,
		}

		eventDate, err := setlistEventDate(s.EventDate)
		if err == nil {
			// Quiety ignore failures of event date for now
			setlist.EventDate = eventDate
		}

		setlists = append(setlists, setlist)
	}

	return setlists, nil
}

func setlistEventDate(date string) (time.Time, error) {
	if date == "" {
		return time.Time{}, fmt.Errorf("invalid date: %s", date)
	}

	layout := "02-01-2006"

	t, err := time.Parse(layout, date)
	if err != nil {
		return time.Time{}, fmt.Errorf("parsing date '%s': %w", date, err)
	}

	return t, nil
}
