package domain

import "errors"

// ErrSessionExpired indicates that the stored Spotify session can no longer be
// refreshed (Spotify returned invalid_grant) and the user must re-authenticate.
var ErrSessionExpired = errors.New("spotify session expired")
