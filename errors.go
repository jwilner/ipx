package ipx

import "errors"

var (
	// ErrVersionMismatch is IP version mismatch error.
	// When we have both IPv4 and IPv6 in the same request.
	ErrVersionMismatch = errors.New("IP version mismatch")

	// ErrInvalidIP is invalid IP address error.
	// When we pass bad or empty IP address.
	ErrInvalidIP = errors.New("invalid IP address")

	// ErrInvalidNetwork is bad IP network error.
	// When we pass bad or empty IP network.
	ErrInvalidNetwork = errors.New("invalid IP network")
)
