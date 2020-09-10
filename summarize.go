package ipx

import (
	"errors"
	b "math/bits"
	"net"
)

var v4InV6Prefix = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff}

// SummarizeRange returns a series of networks which combined cover the range between the first and last addresses,
// inclusive.
func SummarizeRange(first, last net.IP) []*net.IPNet {
	four := first.To4() != nil
	if four != (last.To4() != nil) {
		return nil // versions must be the same
	}
	if four {
		return summarizeRange4(to32(first), to32(last))
	}
	return summarizeRange6(To128(first), To128(last))
}

func summarizeRange4(first, last uint32) (nets []*net.IPNet) {
	for first <= last {
		// the network will either be as long as all the trailing zeros of the first address OR the number of bits
		// necessary to cover the distance between first and last address -- whichever is smaller
		bits := 32
		if trailingZeros := b.TrailingZeros32(first); trailingZeros < bits {
			bits = trailingZeros
		}

		if first != 0 || last != maxUint32 { // guard overflow; this would just be 32 anyway
			if diffBits := 31 - b.LeadingZeros32(last-first+1); diffBits < bits {
				bits = diffBits
			}
		}

		ipN := net.IPNet{IP: make(net.IP, len(net.IPv4zero)), Mask: net.CIDRMask(32-bits, 32)}

		ipN.IP = ipN.IP[:4]
		from32(first, ipN.IP)
		nets = append(nets, &ipN)

		first += 1 << bits
		if first == 0 {
			break
		}
	}
	return
}

func summarizeRange6(first, last Uint128) (nets []*net.IPNet) {
	for first.Cmp(last) != 1 {
		bits := 128
		if trailingZeros := trailingZeros128(first); trailingZeros < bits {
			bits = trailingZeros
		}
		// check extremes to make sure no overflow
		if first.Cmp(Uint128{0, 0}) != 0 || last.Cmp(Uint128{maxUint64, maxUint64}) != 0 {
			if diffBits := 127 - leadingZeros128(last.Minus(first).Add(Uint128{0, 1})); diffBits < bits {
				bits = diffBits
			}
		}

		ipN := net.IPNet{IP: make(net.IP, net.IPv6len), Mask: net.CIDRMask(128-bits, 128)}

		From128(first, ipN.IP)
		nets = append(nets, &ipN)

		first = first.Add(Uint128{0, 1}.Lsh(uint(bits)))
		if first.Cmp(Uint128{0, 0}) == 0 {
			break
		}
	}
	return
}

func trailingZeros128(i Uint128) int {
	trailingZeros := b.TrailingZeros64(i.L)
	if trailingZeros == 64 {
		trailingZeros += b.TrailingZeros64(i.H)
	}
	return trailingZeros
}

func leadingZeros128(i Uint128) int {
	leadingZeros := b.LeadingZeros64(i.H)
	if leadingZeros == 64 {
		leadingZeros += b.LeadingZeros64(i.L)
	}
	return leadingZeros
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
