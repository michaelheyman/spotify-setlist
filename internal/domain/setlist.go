package domain

import (
	"context"
	"time"
)

type Setlist struct {
	Artist    string
	Venue     string
	Songs     []string
	EventDate time.Time
	URL       string
}

type GetSetlistOption func(*GetSetlistOptions)

type GetSetlistOptions struct {
	maxSetlists int
}

type SetlistRepository interface {
	GetSetlists(ctx context.Context, artist string, opts ...GetSetlistOption) ([]Setlist, error)
}

func WithMaxSetlists(maxSetlists int) GetSetlistOption {
	return func(o *GetSetlistOptions) {
		o.maxSetlists = maxSetlists
	}
}
