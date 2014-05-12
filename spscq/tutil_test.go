package spscq

import (
	"github.com/fmstephe/fstrconv"
	"math/rand"
	"testing"
)

func TestMin(t *testing.T) {
	rand.Seed(1)
	for i := 0; i < 1000*1000; i++ {
		a := rand.Int63n(1000 * 1000 * 1000)
		b := rand.Int63n(1000 * 1000 * 1000)
		m := min(a, b)
		om := simpleMin(a, b)
		if m != om {
			as := fstrconv.Itoa64Comma(a)
			bs := fstrconv.Itoa64Comma(b)
			ms := fstrconv.Itoa64Comma(m)
			t.Errorf("Problem with min of %s, %s - min returned %s", as, bs, ms)
		}
	}
}

func simpleMin(val1, val2 int64) int64 {
	if val1 < val2 {
		return val1
	}
	return val2
}
