package domain_test

import (
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/michaelheyman/spotify-cli/internal/domain"
	"github.com/stretchr/testify/assert"
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
			wantErr: func(tt assert.TestingT, err error, _ ...any) bool {
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
			wantErr: func(tt assert.TestingT, err error, _ ...any) bool {
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

func assertValidationField(t assert.TestingT, err error, field, expectedMsg string) bool {
	if !assert.NotNil(t, err, "An error was expected") {
		return false
	}

	validationErr, ok := err.(validation.Errors)
	if !assert.True(t, ok, "Error must be of type validation.Errors") {
		return false
	}

	fieldErr := validationErr[field]
	if !assert.NotNil(t, fieldErr, "Expected error for field: %s", field) {
		return false
	}

	return assert.EqualError(t, fieldErr, expectedMsg)
}
