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

	aInt := to128(a)
	return aInt.Cmp(to128(b))
}
