package ipx

import (
	"encoding/binary"
	"math/bits"
)

// largely cribbed from https://github.com/davidminor/uint128 and https://github.com/lukechampine/uint128
type uint128 struct {
	H, L uint64
}

func (u uint128) And(other uint128) uint128 {
	u.H &= other.H
	u.L &= other.L
	return u
}

func (u uint128) Or(other uint128) uint128 {
	u.H |= other.H
	u.L |= other.L
	return u
}

func (u uint128) Cmp(other uint128) int {
	switch {
	case u.H > other.H:
		return 1
	case u.H < other.H,
		u.L < other.L:
		return -1
	case u.L > other.L:
		return 1
	default:
		return 0
	}
}

// Equal checks 128-bit values are equal.
// Equivalent of `u.Cmp(other) == 0` but a bit fater.
func (u uint128) Equal(other uint128) bool {
	return (u.H == other.H) &&
		(u.L == other.L)
}

func (u uint128) Add(addend uint128) uint128 {
	old := u.L
	u.H += addend.H
	u.L += addend.L
	if u.L < old { // wrapped
		u.H++
	}
	return u
}

func (u uint128) Minus(addend uint128) uint128 {
	old := u.L
	u.H -= addend.H
	u.L -= addend.L
	if u.L > old { // wrapped
		u.H--
	}
	return u
}

func (u uint128) Lsh(bits uint) uint128 {
	switch {
	case bits >= 128:
		u.H, u.L = 0, 0
	case bits >= 64:
		u.H, u.L = u.L<<(bits-64), 0
	default:
		u.H <<= bits
		u.H |= u.L >> (64 - bits) // set top with prefix that cross from bottom
		u.L <<= bits
	}
	return u
}

func (u uint128) Rsh(bits uint) uint128 {
	switch {
	case bits >= 128:
		u.H, u.L = 0, 0
	case bits >= 64:
		u.H, u.L = 0, u.H>>(bits-64)
	default:
		u.L >>= bits
		u.L |= u.H << (64 - bits) // set bottom with prefix that cross from top
		u.H >>= bits
	}
	return u
}

func (u uint128) Not() uint128 {
	return uint128{^u.H, ^u.L}
}

// TrailingZeros counts the number of trailing zeros.
// Just as standard `bits.TrailingZerosXX` does.
func (u uint128) TrailingZeros() int {
	z := bits.TrailingZeros64(u.L)
	if z == 64 {
		z += bits.TrailingZeros64(u.H)
	}
	return z
}

// LeadingZeros counts the number of leading zeros.
// Just as standard `bits.LeadingZerosXX` does.
func (u uint128) LeadingZeros() int {
	z := bits.LeadingZeros64(u.H)
	if z == 64 {
		z += bits.LeadingZeros64(u.L)
	}
	return z
}

func to128(ip []byte) uint128 {
	return uint128{binary.BigEndian.Uint64(ip[:8]), binary.BigEndian.Uint64(ip[8:])}
}

func from128(u uint128, ip []byte) {
	binary.BigEndian.PutUint64(ip[:8], u.H)
	binary.BigEndian.PutUint64(ip[8:], u.L)
}
