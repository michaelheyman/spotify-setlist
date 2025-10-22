package domain

type Song struct {
	Artist string
	Name   string
}

type Setlist struct {
	Songs []Song
}

type GetSetlistOption func(*GetSetlistOptions)

type GetSetlistOptions struct {
	maxSetlists int
}

type SetlistRepository interface {
	GetSetlists(artist string, opts ...GetSetlistOption) ([]Setlist, error)
}

func WithMaxSetlists(max int) GetSetlistOption {
	return func(o *GetSetlistOptions) {
		o.maxSetlists = max
	}
}
