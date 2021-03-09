package ipx_test

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"testing"

	"github.com/ns1/ipx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ExampleSummarizeRange_v4 is an example of SummarizeRange for IPv4
func ExampleSummarizeRange_v4() {
	networks, _ := ipx.SummarizeRange(
		net.ParseIP("192.0.2.0"),
		net.ParseIP("192.0.2.130"),
	)
	fmt.Println(networks)
	// Output:
	// [192.0.2.0/25 192.0.2.128/31 192.0.2.130/32]
}

// ExampleSummarizeRange_v6 is an example of SummarizeRange for IPv6
func ExampleSummarizeRange_v6() {
	networks, _ := ipx.SummarizeRange(
		net.ParseIP("::"),
		net.ParseIP("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff"),
	)
	fmt.Println(networks)
	// Output:
	// [::/0]
}

// BenchmarkSummarizeRange performance benchmarks for SummarizeRange
func BenchmarkSummarizeRange(bb *testing.B) {
	// helper function to run benchmark
	bench := func(first, last string, expectedLen int) func(*testing.B) {
		return func(b *testing.B) {
			ipFirst := net.ParseIP(first)
			ipLast := net.ParseIP(last)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				nwks, err := ipx.SummarizeRange(ipFirst, ipLast)
				require.NoError(b, err, "failed to summarize range")
				// note, we check only the output length.
				// all corner cases should be covered by
				// corresponding unit tests.
				require.Len(b, nwks, expectedLen)
			}
		}
	}

	// IPv4
	bb.Run("ipv4_all", bench("0.0.0.0", "255.255.255.255", 1))
	bb.Run("ipv4_24", bench("10.10.10.0", "10.10.10.255", 1))
	bb.Run("ipv4_32", bench("0.0.0.1", "255.255.255.255", 32))

	// IPv6
	bb.Run("ipv6_all", bench("::", "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff", 1))
	bb.Run("ipv6_128", bench("::1", "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff", 128))
}

// TestSummarizeRange unit tests for SummarizeRange
func TestSummarizeRange(tt *testing.T) {
	tt.Run("mismatched_versions", func(t *testing.T) {
		_, err := ipx.SummarizeRange(
			net.ParseIP("0.0.0.0"),
			net.ParseIP("::1"),
		)
		require.Error(t, err, "should not summarize range")
		assert.ErrorIs(t, err, ipx.ErrVersionMismatch)
		assert.Contains(t, err.Error(), "IP version mismatch")
	})

	tt.Run("bad_first", func(t *testing.T) {
		_, err := ipx.SummarizeRange(
			net.ParseIP("bad"), // nil
			net.ParseIP("192.168.2.200"),
		)
		require.Error(t, err, "should not summarize range")
		assert.ErrorIs(t, err, ipx.ErrInvalidIP)
		assert.Contains(t, err.Error(), "invalid IP address: first")
	})

	tt.Run("bad_last", func(t *testing.T) {
		_, err := ipx.SummarizeRange(
			net.ParseIP("192.168.2.100"),
			net.ParseIP("bad"), // nil
		)
		require.Error(t, err, "should not summarize range")
		assert.ErrorIs(t, err, ipx.ErrInvalidIP)
		assert.Contains(t, err.Error(), "invalid IP address: last")
	})

	tt.Run("bad_both", func(t *testing.T) {
		_, err := ipx.SummarizeRange(
			net.ParseIP("bad"), // nil
			net.ParseIP("bad"), // nil
		)
		require.Error(t, err, "should not summarize range")
		assert.ErrorIs(t, err, ipx.ErrInvalidIP)
		assert.Contains(t, err.Error(), "invalid IP address: first") // `first` is checked first
	})

	tt.Run("ipv4_no_overlap", func(t *testing.T) {
		nwks, err := ipx.SummarizeRange(
			net.ParseIP("192.168.2.200"),
			net.ParseIP("192.168.2.100"),
		)
		require.NoError(t, err, "failed to summarize range")
		assert.Empty(t, nwks.Strings())
	})

	tt.Run("ipv4_simple", func(t *testing.T) {
		nwks, err := ipx.SummarizeRange(
			net.ParseIP("192.0.2.0"),
			net.ParseIP("192.0.2.130"),
		)
		require.NoError(t, err, "failed to summarize range")
		assert.Equal(t, []string{
			"192.0.2.0/25",
			"192.0.2.128/31",
			"192.0.2.130/32",
		}, nwks.Strings())
	})

	tt.Run("ipv4_32", func(t *testing.T) {
		nwks, err := ipx.SummarizeRange(
			net.ParseIP("192.0.2.100").To4(),
			net.ParseIP("192.0.2.100").To4(),
		)
		require.NoError(t, err, "failed to summarize range")
		assert.Equal(t, []string{
			"192.0.2.100/32",
		}, nwks.Strings())
	})

	tt.Run("ipv4_16", func(t *testing.T) {
		nwks, err := ipx.SummarizeRange(
			net.ParseIP("192.168.0.0"),
			net.ParseIP("192.168.255.255"),
		)
		require.NoError(t, err, "failed to summarize range")
		assert.Equal(t, []string{
			"192.168.0.0/16",
		}, nwks.Strings())
	})

	tt.Run("ipv4_all", func(t *testing.T) {
		nwks, err := ipx.SummarizeRange(
			net.ParseIP("0.0.0.0"),
			net.ParseIP("255.255.255.255"),
		)
		require.NoError(t, err, "failed to summarize range")
		assert.Equal(t, []string{
			"0.0.0.0/0",
		}, nwks.Strings())
	})

	tt.Run("ipv4_odd_start", func(t *testing.T) {
		nwks, err := ipx.SummarizeRange(
			net.ParseIP("192.0.2.101"),
			net.ParseIP("192.0.2.130"),
		)
		require.NoError(t, err, "failed to summarize range")
		assert.Equal(t, []string{
			"192.0.2.101/32",
			"192.0.2.102/31",
			"192.0.2.104/29",
			"192.0.2.112/28",
			"192.0.2.128/31",
			"192.0.2.130/32",
		}, nwks.Strings())
	})

	tt.Run("ipv6_no_overlap", func(t *testing.T) {
		nwks, err := ipx.SummarizeRange(
			net.ParseIP("::200"),
			net.ParseIP("::100"),
		)
		require.NoError(t, err, "failed to summarize range")
		assert.Empty(t, nwks.Strings())
	})

	tt.Run("ipv6_128", func(t *testing.T) {
		nwks, err := ipx.SummarizeRange(
			net.ParseIP("::100").To16(),
			net.ParseIP("::100").To16(),
		)
		require.NoError(t, err, "failed to summarize range")
		assert.Equal(t, []string{
			"::100/128",
		}, nwks.Strings())
	})

	tt.Run("ipv6_16", func(t *testing.T) {
		nwks, err := ipx.SummarizeRange(
			net.ParseIP("1::"),
			net.ParseIP("1:ffff:ffff:ffff:ffff:ffff:ffff:ffff"),
		)
		require.NoError(t, err, "failed to summarize range")
		assert.Equal(t, []string{
			"1::/16",
		}, nwks.Strings())
	})

	tt.Run("ipv6_all", func(t *testing.T) {
		nwks, err := ipx.SummarizeRange(
			net.ParseIP("::"),
			net.ParseIP("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff"),
		)
		require.NoError(t, err, "failed to summarize range")
		assert.Equal(t, []string{
			"::/0",
		}, nwks.Strings())
	})

	tt.Run("ipv6_odd_start", func(t *testing.T) {
		nwks, err := ipx.SummarizeRange(
			net.ParseIP("1::1"),
			net.ParseIP("1::30"),
		)
		require.NoError(t, err, "failed to summarize range")
		assert.Equal(t, []string{
			"1::1/128",
			"1::2/127",
			"1::4/126",
			"1::8/125",
			"1::10/124",
			"1::20/124",
			"1::30/128",
		}, nwks.Strings())
	})
}

func ExampleNetToRange() {
	_, nwk, _ := net.ParseCIDR("192.168.1.100/24")
	fmt.Println(ipx.NetToRange(nwk))
	// Output:
	// 192.168.1.0 192.168.1.255
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
