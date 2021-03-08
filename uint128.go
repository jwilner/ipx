package ipx

import "encoding/binary"

// largely cribbed from https://github.com/davidminor/uint128 and https://github.com/lukechampine/uint128
type Uint128 struct {
	H, L uint64
}

func (u Uint128) And(other Uint128) Uint128 {
	u.H &= other.H
	u.L &= other.L
	return u
}

func (u Uint128) Or(other Uint128) Uint128 {
	u.H |= other.H
	u.L |= other.L
	return u
}

func (u Uint128) Cmp(other Uint128) int {
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

func (u Uint128) Add(addend Uint128) Uint128 {
	old := u.L
	u.H += addend.H
	u.L += addend.L
	if u.L < old { // wrapped
		u.H++
	}
	return u
}

func (u Uint128) Minus(addend Uint128) Uint128 {
	old := u.L
	u.H -= addend.H
	u.L -= addend.L
	if u.L > old { // wrapped
		u.H--
	}
	return u
}

func (u Uint128) Lsh(bits uint) Uint128 {
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

func (u Uint128) Rsh(bits uint) Uint128 {
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

func (u Uint128) Not() Uint128 {
	return Uint128{^u.H, ^u.L}
}

func to128(ip []byte) Uint128 {
	return Uint128{binary.BigEndian.Uint64(ip[:8]), binary.BigEndian.Uint64(ip[8:])}
}

func from128(u Uint128, ip []byte) {
	binary.BigEndian.PutUint64(ip[:8], u.H)
	binary.BigEndian.PutUint64(ip[8:], u.L)
}
