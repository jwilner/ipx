package ipx

import (
	"math/big"
	"math/rand"
	"net"
	"testing"
	"time"
)

func BenchmarkUint128(b *testing.B) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	left, right := uint128{r.Uint64(), r.Uint64()}, uint128{r.Uint64(), r.Uint64()}
	shift := uint(r.Intn(128))

	ip6 := make(net.IP, 16)
	_, _ = r.Read(ip6)

	blank := make(net.IP, 16)

	b.Run("new", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = to128(ip6)
		}
	})
	b.Run("from", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			from128(left, blank)
		}
	})
	b.Run("add", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = left.Add(right)
		}
	})
	b.Run("minus", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = left.Minus(right)
		}
	})
	b.Run("lsh", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = left.Lsh(shift)
		}
	})
	b.Run("rsh", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = left.Rsh(shift)
		}
	})
	b.Run("cmp", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = left.Cmp(right)
		}
	})
}

func TestUint128(t *testing.T) {
	b := func() *big.Int { return new(big.Int) }

	maxU64B := b().SetUint64(maxUint64)
	maxU128B := b().Or(b().Lsh(maxU64B, 64), maxU64B)

	for _, c := range []struct {
		name     string
		expr     uint128
		expected *big.Int
	}{

		{
			"add",
			uint128{0, 0}.Add(uint128{0, 1}),
			big.NewInt(1),
		},
		{
			"add overflow",
			uint128{maxUint64, maxUint64}.Add(uint128{0, 1}),
			big.NewInt(0),
		},

		{
			"minus",
			uint128{maxUint64, maxUint64}.Minus(uint128{0, 1}),
			b().Sub(maxU128B, big.NewInt(1)),
		},
		{
			"minus overflow",
			uint128{0, 0}.Minus(uint128{0, 1}),
			b().Or(b().Lsh(maxU64B, 64), maxU64B),
		},

		{
			"lsh max",
			uint128{0, maxUint64}.Lsh(64),
			b().Lsh(maxU64B, 64),
		},
		{
			"lsh one",
			uint128{0, maxUint64}.Lsh(1),
			b().Lsh(maxU64B, 1),
		},
		{
			"lsh zero",
			uint128{0, maxUint64}.Lsh(0),
			b().Lsh(maxU64B, 0),
		},

		{
			"rsh max",
			uint128{0, maxUint64}.Rsh(64),
			b().Rsh(maxU64B, 64),
		},
		{
			"rsh one",
			uint128{0, maxUint64}.Rsh(1),
			b().Rsh(maxU64B, 1),
		},

		{
			"not",
			uint128{0, maxUint64}.Not(),
			b().Lsh(maxU64B, 64),
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			i := b().Or(b().Lsh(b().SetUint64(c.expr.H), 64), b().SetUint64(c.expr.L))
			if i.Cmp(c.expected) != 0 {
				t.Fatalf("expected %v but got %v", c.expected, i)
			}
		})
	}
}
