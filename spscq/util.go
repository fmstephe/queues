package spscq

import ()

func divisble(num, denom int64) bool {
	return num%denom == 0
}

func powerOfTwo(val int64) bool {
	pow := int64(2)
	for i := 0; i < 64; i++ {
		if val == pow {
			return true
		}
		pow *= 2
	}
	return false
}

// NB: Only valid if math.MinInt64 <= x-y <= math.MaxInt64
// This is valid for these queues because x and y will always be positive
func min(x, y int64) int64 {
	return y + ((x - y) & ((x - y) >> 63))
}
