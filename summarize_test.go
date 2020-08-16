package ipx_test

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"testing"

	"github.com/ns1/ipx"
)

func ExampleSummarizeRange() {
	fmt.Println(ipx.SummarizeRange(net.ParseIP("192.0.2.0"), net.ParseIP("192.0.2.130")))
	// Output:
	// [192.0.2.0/25 192.0.2.128/31 192.0.2.130/32]
}

func ExampleSummarizeRange_IP6() {
	fmt.Println(ipx.SummarizeRange(
		net.ParseIP("::"),
		net.ParseIP("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff"),
	))
	// Output:
	// [::/0]
}

func BenchmarkSummarizeRange(b *testing.B) {
	type bench struct {
		name        string
		first, last string
	}
	for _, g := range []struct {
		name    string
		benches []bench
	}{
		{
			"ipv4",
			[]bench{
				{
					"all",
					"0.0.0.0",
					"255.255.255.255",
				},
				{
					"32",
					"0.0.0.1",
					"255.255.255.255",
				},
			},
		},
		{
			"ipv6",
			[]bench{
				{
					"all",
					"::",
					"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
				},
				{
					"128",
					"::1",
					"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
				},
			},
		},
	} {
		b.Run(g.name, func(b *testing.B) {
			for _, c := range g.benches {
				first, last := net.ParseIP(c.first), net.ParseIP(c.last)
				b.Run(c.name, func(b *testing.B) {
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						_ = ipx.SummarizeRange(first, last)
					}
				})
			}
		})
	}
}

func TestSummarizeRange(t *testing.T) {
	for _, c := range []struct {
		name        string
		first, last string
		results     []string
	}{
		{"mismatched versions", "0.0.0.0", "::", []string{}},
		{"no overlap", "192.0.2.1", "192.0.2.0", []string{}},
		{
			"simple",
			"192.0.2.0",
			"192.0.2.130",
			[]string{"192.0.2.0/25", "192.0.2.128/31", "192.0.2.130/32"},
		},
		{
			"single",
			"192.0.2.0",
			"192.0.2.0",
			[]string{"192.0.2.0/32"},
		},
		{
			"all",
			"0.0.0.0",
			"255.255.255.255",
			[]string{"0.0.0.0/0"},
		},
		{
			"odd start",
			"192.0.2.101",
			"192.0.2.130",
			[]string{
				"192.0.2.101/32",
				"192.0.2.102/31",
				"192.0.2.104/29",
				"192.0.2.112/28",
				"192.0.2.128/31",
				"192.0.2.130/32",
			},
		},
		{
			"ipv6",
			"1::",
			"1:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
			[]string{"1::/16"},
		},
		{
			"ipv6 all",
			"::",
			"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
			[]string{"::/0"},
		},
		{
			"ipv6 odd start",
			"1::1",
			"1::30",
			[]string{
				"1::1/128",
				"1::2/127",
				"1::4/126",
				"1::8/125",
				"1::10/124",
				"1::20/124",
				"1::30/128",
			},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			var found []string
			for _, ipN := range ipx.SummarizeRange(net.ParseIP(c.first), net.ParseIP(c.last)) {
				found = append(found, ipN.String())
			}
			if len(c.results) != len(found) {
				t.Fatalf("expected %v elements but got %v: %v", len(c.results), len(found), found)
			}
			for i := range found {
				if c.results[i] != found[i] {
					t.Errorf("position %d: expected %v but got %v", i, c.results[i], found[i])
				}
			}
		})
	}
}

func ExampleNetToRange() {
	fmt.Println(ipx.SummarizeRange(
		net.ParseIP("::"),
		net.ParseIP("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff"),
	))
	// Output:
	// [::/0]
}

func TestNetToRange(t *testing.T) {
	for _, c := range []struct {
		name  string
		cidr  string
		start string
		end   string
	}{
		{
			"ipv4 within subnet",
			"192.168.0.10/29",
			"192.168.0.8",
			"192.168.0.15",
		},
		{
			"ipv4 cross subnets",
			"192.168.0.253/23",
			"192.168.0.0",
			"192.168.1.255",
		},
		{
			"ipv4 mapped ipv6 dot notation",
			"::ffff:192.168.0.10/29",
			"::ffff:192.168.0.8",
			"::ffff:192.168.0.15",
		},
		{
			"ipv4 mapped ipv6",
			"::ffff:c0a8:000A/29",
			"::ffff:c0a8:0008",
			"::ffff:c0a8:000F",
		},
		{
			"ipv6 within subnet",
			"2001:db8::8a2e:370:7334/120",
			"2001:db8::8a2e:370:7300",
			"2001:db8::8a2e:370:73ff",
		},
		{
			"ipv6 cross subnets",
			"2001:db8::8a2e:370:7334/107",
			"2001:db8::8a2e:360:0",
			"2001:db8::8a2e:37f:ffff",
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			parsedCIDR := strings.Split(c.cidr, "/")
			ip := net.ParseIP(parsedCIDR[0])
			mask, _ := strconv.Atoi(parsedCIDR[1])

			maskLen := 32
			if ip.To4() == nil {
				maskLen = 128
			}

			cidr := &net.IPNet{
				IP:   ip,
				Mask: net.CIDRMask(mask, maskLen),
			}

			start, end := ipx.NetToRange(cidr)

			if !net.ParseIP(c.start).Equal(start) {
				t.Errorf("start: expected %v but got %v", c.start, start)
			}

			if !net.ParseIP(c.end).Equal(end) {
				t.Errorf("end: expected %v but got %v", c.end, end)
			}
		})
	}
}

func BenchmarkCIDRtoRange(b *testing.B) {

	ip4Net := &net.IPNet{
		IP:   net.ParseIP("192.168.0.253"),
		Mask: net.CIDRMask(23, 32),
	}

	ip6Net := &net.IPNet{
		IP:   net.ParseIP("2001:db8::8a2e:370:7334"),
		Mask: net.CIDRMask(107, 128),
	}

	b.Run("ipv4", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = ipx.NetToRange(ip4Net)
		}
	})

	b.Run("ipv6", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = ipx.NetToRange(ip6Net)
		}
	})
}
