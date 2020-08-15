package ipx_test

import (
	"fmt"
	"net"
	"testing"

	"github.com/ns1/ipx"
)

func ExampleReversePointer_IPv4() {
	fmt.Println(ipx.ReversePointer(net.ParseIP("192.168.0.10")))
	// Output:
	// 10.0.168.192.in-addr.arpa
}

func ExampleReversePointer_IPv6() {
	fmt.Println(ipx.ReversePointer(net.ParseIP("2001:db8::1")))
	// Output:
	// 1.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.8.b.d.0.1.0.0.2.ip6.arpa
}

func TestReversePointer(t *testing.T) {
	for _, c := range []struct {
		name     string
		input    string
		expected string
	}{
		{
			"IPv4",
			"192.168.0.10",
			"10.0.168.192.in-addr.arpa",
		},
		{
			"IPv6 full",
			"2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			"4.3.3.7.0.7.3.0.e.2.a.8.0.0.0.0.0.0.0.0.3.a.5.8.8.b.d.0.1.0.0.2.ip6.arpa",
		},
		{
			"IPv6 short",
			"2001:db8::1",
			"1.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.8.b.d.0.1.0.0.2.ip6.arpa",
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			result := ipx.ReversePointer(net.ParseIP(c.input))
			if result != c.expected {
				t.Errorf("expected %v but got %v", c.expected, result)
			}
		})
	}
}
