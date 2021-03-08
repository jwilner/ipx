package ipx

import (
	"net"
)

// Split splits a subnet into smaller subnets according to the new prefix provided.
func Split(ipNet *net.IPNet, newPrefix int) *NetIter {
	ones, bits := ipNet.Mask.Size()
	if ones > newPrefix || newPrefix > bits {
		return new(NetIter)
	}
	if ipNet.IP.To4() != nil {
		ip := to32(ipNet.IP)
		return &NetIter{
			ips: *iterIPv4(ip, 1<<(bits-newPrefix), ip|(1<<(bits-ones)-1)),
			net: &net.IPNet{Mask: net.CIDRMask(newPrefix, bits)},
		}
	}

	ip := to128(ipNet.IP)

	incr := uint128{0, 1}.Lsh(uint(bits - newPrefix))

	broadCast := uint128{0, 1}.
		Lsh(uint(bits - ones)).
		Minus(uint128{0, 1}).
		Or(ip)

	return &NetIter{
		*iterIPv6(ip, incr, broadCast),
		&net.IPNet{Mask: net.CIDRMask(newPrefix, bits)},
	}
}

// Addresses returns all of the addresses within a network.
func Addresses(ipNet *net.IPNet) *IPIter {
	ones, bits := ipNet.Mask.Size()
	if ipNet.IP.To4() != nil {
		ip := to32(ipNet.IP)
		return iterIPv4(
			ip,
			1,
			ip+(1<<(bits-ones)),
		)
	}
	ip := to128(ipNet.IP)
	return iterIPv6(
		ip,
		uint128{0, 1},
		ip.Add(uint128{0, 1}.Lsh(uint(bits-ones))),
	)
}

// Hosts returns all of the usable addresses within a network except the network itself address and the broadcast address
func Hosts(ipNet *net.IPNet) *IPIter {
	ones, bits := ipNet.Mask.Size()
	if ipNet.IP.To4() != nil {
		ip := to32(ipNet.IP) + 1
		return iterIPv4(
			ip,
			1,
			ip+(1<<(bits-ones))-2,
		)
	}

	ip := to128(ipNet.IP).Add(uint128{0, 1})

	addend := uint128{0, 1}.
		Lsh(uint(bits - ones)).
		Minus(uint128{0, 2})

	return iterIPv6(
		ip,
		uint128{0, 1},
		ip.Add(addend),
	)
}
