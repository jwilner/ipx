package ipx

import (
	"errors"
	"net"
)

// CmpIP compares two IPs
func CmpIP(a, b net.IP) int {
	four := a.To4() != nil
	if four != (b.To4() != nil) {
		panic(errors.New("IP versions must be the same"))
	}

	aInt := to128(a.To16())
	return aInt.Cmp(to128(b.To16()))
}

// CmpNet compares two networks, using only the IP, disregarding the mask
func CmpNet(a, b *net.IPNet) int {
	if a == nil || b == nil {
		panic(errors.New("neither net can be nil"))
	}
	return CmpIP(a.IP, b.IP)
}
