package ipx

import (
	"net"
)

// Network represents vanilla IP network.
// It's represented as CIDR notation.
type Network = net.IPNet

// Networks is a slice of networks.
type Networks []*Network

// Strings get all networks as string representation.
func (nn Networks) Strings() []string {
	if nn == nil {
		return nil
	}

	out := make([]string, 0, len(nn))
	for _, n := range nn {
		out = append(out, n.String())
	}

	return out
}
