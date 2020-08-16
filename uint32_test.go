package ipx

import (
	"math/rand"
	"net"
	"testing"
	"time"
)

func BenchmarkUint32(b *testing.B) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	left, right := r.Uint32(), r.Uint32()
	shift := uint(r.Intn(32))

	ip4 := make(net.IP, 16)
	copy(ip4, net.IPv4zero)
	_, _ = r.Read(ip4[12:])

	blank := make(net.IP, 16)
	copy(blank, net.IPv4zero)

	b.Run("new", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = to32(ip4)
		}
	})
	b.Run("from", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			from32(left, blank)
		}
	})
	b.Run("add", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = left + right
		}
	})
	b.Run("minus", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = left - right
		}
	})
	b.Run("lsh", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = left << shift
		}
	})
	b.Run("rsh", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = left >> shift
		}
	})
	b.Run("cmp", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = left < right
		}
	})
}
