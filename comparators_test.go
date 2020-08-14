package ipx_test

import (
	"net"
	"testing"

	"github.com/ns1/ipx"
)

func TestCmpIPPanic(t *testing.T) {
	defer func() { recover() }()
	a := net.ParseIP("192.168.0.10")
	b := net.ParseIP("2001:db8::1")

	ipx.CmpIP(a, b)

	t.Errorf("did not panic")
}

func TestCmpIP(t *testing.T) {
	for _, c := range []struct {
		name     string
		a        string
		b        string
		expected int
	}{
		{
			"ipv4 /24 less than",
			"192.168.0.10",
			"192.168.0.20",
			-1,
		},
		{
			"ipv4 /24 greater than",
			"192.168.0.20",
			"192.168.0.10",
			1,
		},
		{
			"ipv4 /24 equal",
			"192.168.0.10",
			"192.168.0.10",
			0,
		},
		{
			"ipv4 /16 less than",
			"192.168.10.20",
			"192.168.20.10",
			-1,
		},
		{
			"ipv4 /16 greater than",
			"192.168.20.10",
			"192.168.10.20",
			1,
		},
		{
			"ipv4 /8 less than",
			"180.254.254.254",
			"190.254.254.254",
			-1,
		},
		{
			"ipv4 /8 greater than",
			"190.254.254.254",
			"180.254.254.254",
			1,
		},
		{
			"ipv6 /16 less than",
			"2001:0db8:85a3:0000:0000:8a2e:0370:6000",
			"2001:0db8:85a3:0000:0000:8a2e:0370:7000",
			-1,
		},
		{
			"ipv6 /16 greater than",
			"2001:0db8:85a3:0000:0000:8a2e:0370:7000",
			"2001:0db8:85a3:0000:0000:8a2e:0370:6000",
			1,
		},
		{
			"ipv6 /16 equal",
			"2001:0db8:85a3:0000:0000:8a2e:0370:6000",
			"2001:0db8:85a3:0000:0000:8a2e:0370:6000",
			0,
		},
		{
			"ipv6 /32 less than",
			"2001:0db8:85a3:0000:0000:8a2e:6000:7000",
			"2001:0db8:85a3:0000:0000:8a2e:7000:6000",
			-1,
		},
		{
			"ipv6 /32 greater than",
			"2001:0db8:85a3:0000:0000:8a2e:7000:6000",
			"2001:0db8:85a3:0000:0000:8a2e:6000:7000",
			1,
		},
		{
			"ipv6 /128 less than",
			"2001:eeee:eeee:eeee:eeee:eeee:eeee:eeee",
			"3001:eeee:eeee:eeee:eeee:eeee:eeee:eeee",
			-1,
		},
		{
			"ipv6 /128 greater than",
			"3001:eeee:eeee:eeee:eeee:eeee:eeee:eeee",
			"2001:eeee:eeee:eeee:eeee:eeee:eeee:eeee",
			1,
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			if result := ipx.CmpIP(net.ParseIP(c.a), net.ParseIP(c.b)); result != c.expected {
				t.Errorf("expected %v but got %v", c.expected, result)
			}
		})
	}
}
