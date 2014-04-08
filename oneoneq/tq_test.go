package oneoneq

import (
	"runtime"
	"testing"
	"time"
	"fmt"
)

func TestFoo(t *testing.T) {
	runtime.GOMAXPROCS(4)
	var itemCount int64 = 100 * 1000 * 1000
	q := New(itemCount)
	done := make(chan bool)
	go dequeue(itemCount, q, done)
	go enqueue(itemCount, q, done)
	<-done
	<-done
}

func enqueue(num int64, q *Q, done chan bool) {
	runtime.LockOSThread()
	println("Entering Enqueue")
	for i := int64(0); i < num; i++ {
		for b := false; b == false; b = q.Enqueue(i) {}
	}
	done <- true
}

func dequeue(num int64, q *Q, done chan bool) {
	runtime.LockOSThread()
	start := time.Now().UnixNano()
	println("Entering Dequeue")
	sum := int64(0)
	checksum := int64(0)
	for i := int64(0); i < num; i++ {
		var item int64 = -1
		for ; item == -1; item = q.Dequeue() {}
		sum += item
		checksum += i
	}
	print(fmt.Sprintf("sum      %d\nchecksum %d\n", sum, checksum))
	println()
	total := time.Now().UnixNano() - start
	nanos := total
	micros := total/1000
	millis := micros/1000
	seconds := millis/1000
	print(fmt.Sprintf("\nNanos   %d\nMicros  %d\nMillis  %d\nSeconds %d\n",nanos, micros, millis, seconds))
	done <- true
}
