package ipx

import (
	"net"
	"sort"
)

// Collapse combines subnets into their closest available parent.
func Collapse(toMerge []*net.IPNet) []*net.IPNet {
	var (
		four []ip4Net
		six  []ip6Net
	)
	for _, ipN := range toMerge {
		if ipN.IP.To4() != nil {
			four = append(four, newIP4Net(ipN))
			continue
		}
		six = append(six, newIP6Net(ipN))
	}
	return append(collapse4(four), collapse6(six)...)
}

func collapse4(nets []ip4Net) []*net.IPNet {
	if len(nets) == 0 {
		return nil
	}

	supers := make(map[ip4Net]ip4Net)
	for len(nets) > 0 {
		n := nets[len(nets)-1]
		nets = nets[:len(nets)-1]

		s := n.super()
		other, ok := supers[s]
		if !ok {
			supers[s] = n
			continue
		}
		if other == n {
			continue
		}

		// we have found two nets with same immediate parent -- merge 'em
		delete(supers, s)
		nets = append(nets, s)
	}

	merged := make(ip4Nets, 0, len(supers))
	for _, v := range supers {
		merged = append(merged, v)
	}
	sort.Sort(merged)

	result := []*net.IPNet{merged[0].asNet()}
	lastMask := merged[0].mask()
	lastAddr := merged[0].addr
	for _, m := range merged[1:] {
		if lastAddr == m.addr&lastMask {
			continue
		}
		result = append(result, m.asNet())
		lastMask, lastAddr = m.mask(), m.addr
	}
	return result
}

func collapse6(nets []ip6Net) []*net.IPNet {
	if len(nets) == 0 {
		return nil
	}

	supers := make(map[ip6Net]ip6Net)
	for len(nets) > 0 {
		n := nets[len(nets)-1]
		nets = nets[:len(nets)-1]

		s := n.super()
		other, ok := supers[s]
		if !ok {
			supers[s] = n
			continue
		}
		if other == n {
			continue
		}

		// we have found two nets with same immediate parent -- merge 'em
		delete(supers, s)
		nets = append(nets, s)
	}

	merged := make(ip6Nets, 0, len(supers))
	for _, v := range supers {
		merged = append(merged, v)
	}
	sort.Sort(merged)

	result := []*net.IPNet{merged[0].asNet()}
	lastMask := merged[0].mask()
	lastAddr := merged[0].addr
	for _, m := range merged[1:] {
		if lastAddr == m.addr.And(lastMask) {
			continue
		}
		result = append(result, m.asNet())
		lastMask, lastAddr = m.mask(), m.addr
	}
	return result
}

type ip4Net struct {
	addr   uint32
	prefix uint8
}

func newIP4Net(ipN *net.IPNet) ip4Net {
	ones, _ := ipN.Mask.Size()
	return ip4Net{to32(ipN.IP), uint8(ones)}
}

func (n ip4Net) super() ip4Net {
	n.addr &^= 1 << (32 - n.prefix) // unset last bit of net mask to find supernet address
	n.prefix--
	return n
}

func (n ip4Net) asNet() *net.IPNet {
	r := &net.IPNet{IP: make(net.IP, 4), Mask: make(net.IPMask, 4)}
	from32(n.addr, r.IP)
	from32(n.mask(), r.Mask) // set first eight bits
	return r
}

func (n ip4Net) subnets() (ip4Net, ip4Net) {
	a := ip4Net{addr: n.addr, prefix: n.prefix + 1}
	b := ip4Net{addr: n.addr | 1<<(31-n.prefix), prefix: n.prefix + 1}
	return a, b
}

func (n ip4Net) subnetOf(o ip4Net) bool {
	return n.addr&o.mask() == o.addr
}

func (n ip4Net) mask() uint32 {
	return ^(1<<(32-n.prefix) - 1)
}

type ip4Nets []ip4Net

func (n ip4Nets) Len() int {
	return len(n)
}

func (n ip4Nets) Less(i, j int) bool {
	return n[i].addr < n[j].addr
}

func (n ip4Nets) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

type ip6Net struct {
	addr   uint128
	prefix uint8
}

func newIP6Net(ipN *net.IPNet) ip6Net {
	ones, _ := ipN.Mask.Size()
	return ip6Net{to128(ipN.IP), uint8(ones)}
}

func (n ip6Net) super() ip6Net {
	// unset last bit of net mask to find supernet address
	n.addr = n.addr.And(uint128{0, 1}.Lsh(128 - uint(n.prefix)).Not())
	n.prefix--
	return n
}

func (n ip6Net) subnets() (ip6Net, ip6Net) {
	a := ip6Net{addr: n.addr, prefix: n.prefix + 1}
	b := ip6Net{addr: n.addr.Or(uint128{0, 1}.Lsh(uint(127 - n.prefix))), prefix: n.prefix + 1}
	return a, b
}

func (n ip6Net) asNet() *net.IPNet {
	r := &net.IPNet{IP: make(net.IP, 16), Mask: make(net.IPMask, 16)}
	from128(n.addr, r.IP)
	from128(n.mask(), r.Mask) // set prefix bits
	return r
}

func (n ip6Net) mask() uint128 {
	return uint128{0, 1}.Lsh(128 - uint(n.prefix)).Minus(uint128{0, 1}).Not()
}

func (n ip6Net) subnetOf(o ip6Net) bool {
	return n.addr.And(o.mask()).Cmp(o.addr) == 0
}

type ip6Nets []ip6Net

func (n ip6Nets) Len() int {
	return len(n)
}

func (n ip6Nets) Less(i, j int) bool {
	return n[i].addr.Cmp(n[j].addr) == -1
}

func (n ip6Nets) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}
