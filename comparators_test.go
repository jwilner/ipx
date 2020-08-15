package ipx_test

import (
	"fmt"
	"net"
	"testing"

	"github.com/ns1/ipx"
)

func ExampleCmpIP_IP4() {
	a := net.ParseIP("192.168.0.10")
	b := net.ParseIP("192.168.0.20")
	fmt.Println(ipx.CmpIP(a, b))
	// Output:
	// -1
}

func ExampleCmpIP_IP6() {
	a := net.ParseIP("2001:db8::1")
	b := net.ParseIP("2001:db8::cafe")
	fmt.Println(ipx.CmpIP(a, b))
	// Output:
	// -1
}

func ExampleCmpNet_IP4() {
	a := cidr("192.168.0.10/24")
	b := cidr("192.168.0.10/16")
	fmt.Println(ipx.CmpNet(a, b))
	// Output:
	// 0
}

func ExampleCmpNet_IP6() {
	a := cidr("2001:db8::1/128")
	b := cidr("2001:db8::1/64")
	fmt.Println(ipx.CmpNet(a, b))
	// Output:
	// 1
}

func TestCmpIPPanic(t *testing.T) {
	defer func() { recover() }()
	a := net.ParseIP("192.168.0.10")
	b := net.ParseIP("2001:db8::1")

	ipx.CmpIP(a, b)

	t.Errorf("did not panic")
}

func TestCmpNetPanic(t *testing.T) {
	defer func() { recover() }()
	a := cidr("192.168.0.10/24")
	ipx.CmpNet(a, nil)

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

func TestCmpNet(t *testing.T) {
	for _, c := range []struct {
		name     string
		a        string
		b        string
		expected int
	}{
		{
			"ipv4 different masks greater than",
			"192.168.0.20/32",
			"192.168.0.20/24",
			1,
		},
		{
			"ipv4 different masks less than",
			"192.168.0.20/32",
			"192.168.10.20/24",
			-1,
		},
		{
			"ipv4 different masks equal",
			"192.168.0.0/32",
			"192.168.0.0/24",
			0,
		},
		{
			"ipv4 /24 less than",
			"192.168.0.10/24",
			"192.168.10.20/24",
			-1,
		},
		{
			"ipv4 /24 greater than",
			"192.168.10.20/24",
			"192.168.0.10/24",
			1,
		},
		{
			"ipv4 /24 equal",
			"192.168.0.10/24",
			"192.168.0.20/24",
			0,
		},
		{
			"ipv6 /112 less than",
			"2001:0db8:85a3:0000:0000:8a2e:0370:6000/112",
			"2001:0db8:85a3:0000:0000:8a2e:4000:7000/112",
			-1,
		},
		{
			"ipv6 /112 greater than",
			"2001:0db8:85a3:0000:0000:8a2e:4000:7000/112",
			"2001:0db8:85a3:0000:0000:8a2e:0370:6000/112",
			1,
		},
		{
			"ipv6 /112 equal",
			"2001:0db8:85a3:0000:0000:8a2e:0370:6000/112",
			"2001:0db8:85a3:0000:0000:8a2e:0370:6000/112",
			0,
		},
		{
			"ipv6 different masks less than",
			"2001:0db8:85a3:0000:0000:8a2e:0370:6000/128",
			"2001:0db8:85a3:0000:0000:8a2e:4000:6000/112",
			-1,
		},
		{
			"ipv6 different masks greater than",
			"2001:0db8:85a3:0000:0000:8a2e:0370:6000/128",
			"2001:0db8:85a3:0000:0000:8a2e:0370:6000/112",
			1,
		},
		{
			"ipv6 different masks equal",
			"2001:0db8:85a3:0000:0000:8a2e:0370:0000/128",
			"2001:0db8:85a3:0000:0000:8a2e:0370:6000/112",
			0,
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			if result := ipx.CmpNet(cidr(c.a), cidr(c.b)); result != c.expected {
				t.Errorf("expected %v but got %v", c.expected, result)
			}
		})
	}
}
