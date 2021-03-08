package ipx

import (
	"bytes"
	"math"
	"net"
)

const (
	maxUint32 = math.MaxUint32
	maxUint64 = math.MaxUint64
)

type v4IPIter struct {
	val, incr, limit uint32
}

type v6IPIter struct {
	val, incr, limit uint128
}

const (
	ipIterFlagV6 = 1 << iota
	ipIterFlagNegative
)

// IPIter permits iteration over a series of ips. It is always start inclusive.
type IPIter struct {
	v4    v4IPIter
	v6    v6IPIter
	flags uint8

	ip net.IP
}

// IP returns the most recent IP; the underlying type may be modified on later calls to `Next`.
// It does no allocation.
func (i *IPIter) IP() net.IP {
	return i.ip
}

// Next returns true when the underlying pointer has been successfully updated with the next value.
func (i *IPIter) Next() bool {
	if i.flags&ipIterFlagV6 > 0 {
		if i.flags&ipIterFlagNegative > 0 {
			if i.v6.val.Cmp(i.v6.limit) != 1 {
				return false
			}
			from128(i.v6.val, i.ip)
			old := i.v6.val
			if i.v6.val = i.v6.val.Minus(i.v6.incr); old.Cmp(i.v6.val) == -1 {
				i.v6.val = uint128{0, 0}
			}
			return true
		}
		if i.v6.val.Cmp(i.v6.limit) != -1 {
			return false
		}
		from128(i.v6.val, i.ip)
		i.v6.val = i.v6.val.Add(i.v6.incr)
		return true
	}
	if i.flags&ipIterFlagNegative > 0 {
		if i.v4.val <= i.v4.limit {
			return false
		}
		from32(i.v4.val, i.ip)
		i.v4.val -= i.v4.incr
		return true
	}
	if i.v4.val >= i.v4.limit {
		return false
	}
	from32(i.v4.val, i.ip)
	i.v4.val += i.v4.incr
	return true
}

// IterIP returns an iter for the given step from [start, end). If end is nil, it is set to the maximum type for
// the version. If the step is zero, IP versions mismatch or the sign of the increment doesn't match that of
// end - start, an empty iter is returned.
func IterIP(start net.IP, step int, end net.IP) *IPIter {
	if step == 0 {
		return new(IPIter)
	}

	if start.To4() != nil {
		return resolveIPs4(start, step, end, 0)
	}
	return resolveIPs6(start, step, end, 0)
}

func iterIPv4(val, incr, limit uint32) *IPIter {
	iter := IPIter{ip: make(net.IP, len(net.IPv4zero)), v4: v4IPIter{val, incr, limit}}
	copy(iter.ip, net.IPv4zero)
	if limit < val {
		iter.flags |= ipIterFlagNegative
	}
	return &iter
}

func iterIPv6(val, incr, limit uint128) *IPIter {
	iter := IPIter{
		ip:    make(net.IP, len(net.IPv6zero)),
		v6:    v6IPIter{val, incr, limit},
		flags: ipIterFlagV6,
	}
	copy(iter.ip, net.IPv6zero)
	if limit.Cmp(val) == -1 {
		iter.flags |= ipIterFlagNegative
	}
	return &iter
}

// NetIter permits iteration over a series of IP networks. It is always start inclusive.
type NetIter struct {
	ips IPIter
	net *net.IPNet
}

// Net returns the most recent IPNet; the underlying type may be modified on later calls to `Next`.
// It does no allocation.
func (n *NetIter) Net() *net.IPNet {
	n.net.IP = n.ips.IP()
	return n.net
}

// Next returns true when the underlying pointer has been successfully updated with the next value.
func (n *NetIter) Next() bool {
	return n.ips.Next()
}

// IterNet returns an iterator for the given increment starting with the provided network
func IterNet(start *net.IPNet, step int, end *net.IPNet) *NetIter {
	if step == 0 {
		return new(NetIter)
	}

	var endIP net.IP
	if end != nil {
		if !bytes.Equal(start.Mask, end.Mask) {
			return new(NetIter)
		}
		endIP = end.IP
	}

	mask := make(net.IPMask, len(start.Mask))
	copy(mask, start.Mask)

	ones, bits := mask.Size()
	suffix := uint(bits - ones)

	if start.IP.To4() != nil {
		return &NetIter{*resolveIPs4(start.IP, step, endIP, suffix), &net.IPNet{Mask: mask}}
	}
	return &NetIter{*resolveIPs6(start.IP, step, endIP, suffix), &net.IPNet{Mask: mask}}
}

func resolveIPs4(start net.IP, step int, end net.IP, shift uint) *IPIter {
	sIP := to32(start)
	if step > 0 {
		eIP := uint32(maxUint32)
		if end != nil {
			if end.To4() == nil {
				return new(IPIter)
			}
			eIP = to32(end)
			if eIP <= sIP {
				return new(IPIter)
			}
		}
		return iterIPv4(sIP, uint32(step)<<shift, eIP)
	}
	var eIP uint32
	if end != nil {
		if end.To4() == nil {
			return new(IPIter)
		}
		eIP = to32(end)
		if eIP >= sIP {
			return new(IPIter)
		}
	}
	return iterIPv4(sIP, uint32(step*-1)<<shift, eIP)
}

func resolveIPs6(start net.IP, step int, end net.IP, shift uint) *IPIter {
	sIP := to128(start)
	if step > 0 {
		eIP := uint128{maxUint64, maxUint64}
		if end != nil {
			if end.To4() != nil {
				return new(IPIter)
			}
			eIP = to128(end)
			if eIP.Cmp(sIP) != 1 {
				return new(IPIter)
			}
		}
		return iterIPv6(sIP, uint128{0, uint64(step)}.Lsh(shift), eIP)
	}
	var eIP uint128
	if end != nil {
		if end.To4() != nil {
			return new(IPIter)
		}
		eIP = to128(end)
		if eIP.Cmp(sIP) != -1 {
			return new(IPIter)
		}
	}
	return iterIPv6(sIP, uint128{0, uint64(step * -1)}.Lsh(shift), eIP)
}
