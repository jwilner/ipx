package ipx

import (
	"errors"
	"fmt"
	"math/bits"
	"net"
)

var v4InV6Prefix = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff}

// SummarizeRange returns a series of networks which cover the range
// between the first and last addresses, inclusive.
func SummarizeRange(first, last net.IP) ([]*Network, error) {
	// first IPv4 or IPv6
	var firstV4, firstV6 net.IP
	switch len(first) {
	case net.IPv4len:
		firstV4 = first // first is IPv4

	case net.IPv6len:
		// note, even if IP address length is 128 it still can be IPv4!
		// need to do additional check converting it with `To4()`
		firstV4 = first.To4()
		if firstV4 == nil {
			firstV6 = first
	}

	default:
		// invalid first IP address length
		return nil, fmt.Errorf("%w: first", ErrBadIP)
	}

	// last IPv4 or IPv6
	var lastV4, lastV6 net.IP
	switch len(last) {
	case net.IPv4len:
		lastV4 = last // last is IPv4

	case net.IPv6len:
		// note, even if IP address length is 128 it still can be IPv4!
		// need to do additional check converting it with `To4()`
		lastV4 = last.To4()
		if lastV4 == nil {
			lastV6 = last
}

	default:
		// invalid last IP address length
		return nil, fmt.Errorf("%w: last", ErrBadIP)
	}

	switch {
	case firstV4 != nil && lastV4 != nil:
		return summarizeRange4(to32(firstV4), to32(lastV4)), nil
	case firstV6 != nil && lastV6 != nil:
		return summarizeRange6(to128(firstV6), to128(lastV6)), nil
	}

	return nil, ErrVersionMismatch
}

// summarizeRange4 returns a series of IPv4 networks which cover the range
// between the first and last IPv4 addresses, inclusive.
func summarizeRange4(first, last uint32) (networks []*Network) {
	for first <= last {
		// the network will either be as long as all the trailing zeros of the first address OR the number of bits
		// necessary to cover the distance between first and last address -- whichever is smaller
		nBits := 32
		if z := bits.TrailingZeros32(first); z < nBits {
			nBits = z
		}

		if first != 0 || last != maxUint32 { // guard overflow; this would just be 32 anyway
			d := last - first + 1
			if z := 31 - bits.LeadingZeros32(d); z < nBits {
				nBits = z
			}
		}

		nwkMask := net.CIDRMask(32-nBits, 32)
		nwkIP := make(net.IP, net.IPv4len)
		from32(first, nwkIP)
		networks = append(networks,
			&Network{
				IP:   nwkIP,
				Mask: nwkMask,
			})

		first += 1 << nBits
		if first == 0 {
			break
		}
	}

	return
}

// summarizeRange6 returns a series of IPv6 networks which cover the range
// between the first and last IPv6 addresses, inclusive.
func summarizeRange6(first, last uint128) (networks []*Network) {
	for first.Cmp(last) <= 0 { // first <= last
		// the network will either be as long as all the trailing zeros of the first address OR the number of bits
		// necessary to cover the distance between first and last address -- whichever is smaller
		nBits := 128
		if z := first.TrailingZeros(); z < nBits {
			nBits = z
		}

		// check extremes to make sure no overflow
		if !first.Equal(uint128{0, 0}) || !last.Equal(uint128{maxUint64, maxUint64}) {
			d := last.Minus(first).Add(uint128{0, 1})
			if z := 127 - d.LeadingZeros(); z < nBits {
				nBits = z
			}
		}

		nwkMask := net.CIDRMask(128-nBits, 128)
		nwkIP := make(net.IP, net.IPv6len)
		from128(first, nwkIP)
		networks = append(networks,
			&Network{
				IP:   nwkIP,
				Mask: nwkMask,
			})

		first = first.Add(uint128{0, 1}.Lsh(uint(nBits)))
		if first.Equal(uint128{0, 0}) {
			break
		}
	}

	return
}

func allFF(b []byte) bool {
	for _, c := range b {
		if c != 0xff {
			return false
		}
	}
	return true
}

func bytesEqual(a, b []byte) bool {
	for len(a) != len(b) {
		panic(errors.New("a and b are not equal length"))
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// NetToRange returns the start and end IPs for the given net
func NetToRange(cidr *net.IPNet) (start, end net.IP) {
	// Ripped mostly from net.IP.Mask()
	if cidr == nil {
		panic(errors.New("cidr must not be nil"))
	}

	ip, mask := cidr.IP, cidr.Mask

	if len(mask) == net.IPv6len && len(ip) == net.IPv4len && allFF(mask[:12]) {
		mask = mask[12:]
	}

	// IPv4-mapped IPv6 address
	if len(mask) == net.IPv4len && len(ip) == net.IPv6len && bytesEqual(ip[:12], v4InV6Prefix) {
		ip = ip[12:]
	}

	n := len(ip)
	if n != len(mask) {
		return nil, nil
	}

	start = make(net.IP, n)
	end = make(net.IP, n)
	for i := 0; i < n; i++ {
		start[i] = ip[i] & mask[i]
		end[i] = ip[i] | ^mask[i]
	}

	return
}
