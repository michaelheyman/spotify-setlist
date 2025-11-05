package domain_test

import (
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/michaelheyman/spotify-setlist/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlaylist_Validate(t *testing.T) {
	testdata := struct {
		playlist domain.Playlist
	}{
		playlist: domain.Playlist{
			Name:        "name",
			Description: "description",
			Artist:      "artist",
			Songs:       []string{"foo", "bar", "baz"},
		},
	}
	tests := []struct {
		name     string // description of this test case
		playlist domain.Playlist
		wantErr  assert.ErrorAssertionFunc
	}{
		{
			name:     "should validate when all fields are present",
			playlist: testdata.playlist,
			wantErr:  assert.NoError,
		},
		{
			name: "should return error when playlist is missing artist",
			playlist: domain.Playlist{
				Name:        "name",
				Description: "description",
				Songs:       []string{"foo", "bar", "baz"},
			},
			wantErr: func(_ assert.TestingT, err error, _ ...any) bool {
				return assertValidationField(t, err, "Artist", "cannot be blank")
			},
		},
		{
			name: "should return error when playlist is missing songs",
			playlist: domain.Playlist{
				Name:        "name",
				Description: "description",
				Artist:      "artist",
			},
			wantErr: func(_ assert.TestingT, err error, _ ...any) bool {
				return assertValidationField(t, err, "Songs", "cannot be blank")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.playlist.Validate()

			tt.wantErr(t, err, "Validate returned error")
		})
	}
}

func assertValidationField(t *testing.T, err error, field, expectedMsg string) bool {
	require.Error(t, err, "An error was expected")
	var validationErr validation.Errors
	require.ErrorAs(t, err, &validationErr, "Error must be of type validation.Errors")
	fieldErr := validationErr[field]
	require.Error(t, fieldErr, "Expected error for field: %s", field)
	return assert.EqualError(t, fieldErr, expectedMsg)
}
