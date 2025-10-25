package domain

import (
	"context"
	"time"
)

type SetlistGetter interface {
	GetSetlists(ctx context.Context, artist string, opts ...GetSetlistOption) ([]Setlist, error)
}

type Setlist struct {
	Artist    string
	Venue     string
	Songs     []string
	EventDate time.Time
}

type GetSetlistOption func(*GetSetlistOptions)

type GetSetlistOptions struct {
	maxSetlists int
}

type SetlistRepository interface {
	GetSetlists(ctx context.Context, artist string, opts ...GetSetlistOption) ([]Setlist, error)
}

func WithMaxSetlists(max int) GetSetlistOption {
	return func(o *GetSetlistOptions) {
		o.maxSetlists = max
	}
}
