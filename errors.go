package ipx

import "errors"

var (
	// ErrVersionMismatch is IP version mismatch error.
	// When we have both IPv4 and IPv6 in the same request.
	ErrVersionMismatch = errors.New("IP version mismatch")

	// ErrBadIP is bad IP address error.
	// When we pass bad or empty IP address.
	ErrBadIP = errors.New("bad IP address")

	// ErrBadNetwork is bad IP network error.
	// When we pass bad or empty IP network.
	ErrBadNetwork = errors.New("bad IP network")
)
