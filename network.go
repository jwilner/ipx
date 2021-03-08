package ipx

import (
	"net"
)

// Network represents vanilla IP network.
// It's represented by CIDR notation.
type Network = net.IPNet
