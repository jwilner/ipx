package ipx

import (
	"encoding/binary"
	"errors"
	"net"
)

// IncrIP returns the next IP
func IncrIP(ip net.IP, incr int) {
	if ip == nil {
		panic(errors.New("IP cannot be nil"))
	}
	if ip.To4() != nil {
		n := to32(ip)
		if incr >= 0 {
			n += uint32(incr)
		} else {
			n -= uint32(incr * -1)
		}
		from32(n, ip)
		return
	}

	// ipv6
	u := To128(ip)
	if incr >= 0 {
		u = u.Add(Uint128{0, uint64(incr)})
	} else {
		u = u.Minus(Uint128{0, uint64(incr * -1)})
	}
	From128(u, ip)
}

// IncrNet steps to the next net of the same mask
func IncrNet(ipNet *net.IPNet, incr int) {
	if ipNet.IP == nil {
		panic(errors.New("IP cannot be nil"))
	}
	if ipNet.Mask == nil {
		panic(errors.New("mask cannot be nil"))
	}
	if ipNet.IP.To4() != nil {
		n := to32(ipNet.IP)
		ones, bits := ipNet.Mask.Size()
		suffix := uint32(bits - ones)
		n >>= suffix
		if incr >= 0 {
			n += uint32(incr)
		} else {
			n -= uint32(incr * -1)
		}
		from32(n<<suffix, ipNet.IP)
		return
	}

	b := To128(ipNet.IP)

	ones, bits := ipNet.Mask.Size()
	suffix := uint(bits - ones)

	b = b.Rsh(suffix)
	if incr >= 0 {
		b = b.Add(Uint128{0, uint64(incr)})
	} else {
		b = b.Minus(Uint128{0, uint64(incr * -1)})
	}
	b = b.Lsh(suffix)

	From128(b, ipNet.IP)
}

func to32(ip []byte) uint32 {
	l := len(ip)
	return binary.BigEndian.Uint32(ip[l-4:])
}

func from32(n uint32, ip []byte) {
	l := len(ip)
	binary.BigEndian.PutUint32(ip[l-4:], n)
}
