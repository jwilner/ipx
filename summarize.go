package ipx

import (
	b "math/bits"
	"net"
)

// SummarizeRange returns a series of networks which combined cover the range between the first and last addresses,
// inclusive.
func SummarizeRange(first, last net.IP) []*net.IPNet {
	four := first.To4() != nil
	if four != (last.To4() != nil) {
		panic("addresses must be same version")
	}
	if four {
		return summarizeRange4(to32(first), to32(last))
	}
	return summarizeRange6(to128(first), to128(last))
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

		ipN := net.IPNet{IP: make(net.IP, len(net.IPv4zero)), Mask: net.CIDRMask(int(32-bits), 32)}

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

func summarizeRange6(first, last uint128) (nets []*net.IPNet) {
	for first.Cmp(last) != 1 {
		bits := uint64(128)
		if trailingZeros := countBits128(first.Minus(uint128{0, 1}).And(first.Not())); trailingZeros < bits {
			bits = trailingZeros
		}
		// check extremes to make sure no overflow
		if first.Cmp(uint128{0, 0}) != 0 || last.Cmp(uint128{maxUint64, maxUint64}) != 0 {
			if diffBits := countBits128(last.Minus(first).Add(uint128{0, 1})) - 1; diffBits < bits {
				bits = diffBits
			}
		}

		ipN := net.IPNet{IP: make(net.IP, net.IPv6len), Mask: net.CIDRMask(int(128-bits), 128)}

		from128(first, ipN.IP)
		nets = append(nets, &ipN)

		first = first.Add(uint128{0, 1}.Lsh(uint(bits)))
		if first.Cmp(uint128{0, 0}) == 0 {
			break
		}
	}
	return
}

func countBits(i uint64) (bits uint64) {
	for i>>bits != 0 {
		bits++
	}
	return
}

func countBits128(i uint128) (bits uint64) {
	if highBits := countBits(i.H); highBits > 0 {
		return highBits + 64
	}
	return countBits(i.L)
}
