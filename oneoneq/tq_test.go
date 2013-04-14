package oneoneq

import (
	"runtime"
	"testing"
	"time"
)

func TestFoo(t *testing.T) {
	var itemCount int64 = 10 * 1000 * 1000
	q := New(itemCount)
	//var target int64
	done := make(chan bool)
	start := time.Now().UnixNano()
	go dequeue(itemCount, q, done)
	<-done
	go enqueue(itemCount, q, done)
	<-done
	<-done
	total := time.Now().UnixNano() - start
	println(total)
	println(total/1000)
	println(total/(1000*1000))
	println(total/(1000*1000*1000))
}

func enqueue(num int64, q *Q, done chan bool) {
	runtime.LockOSThread()
	for i := int64(0); i < num; i++ {
		for b := false; b == false; b = q.Enqueue(i) {}
	}
	done <- true
}

func dequeue(num int64, q *Q, done chan bool) {
	runtime.LockOSThread()
	done <- true
	sum := int64(0)
	checksum := int64(0)
	for i := int64(0); i < num; i++ {
		var item int64 = -1
		for ; item == -1; item = q.Dequeue() {}
		sum += item
		checksum += i
	}
	println(sum)
	println(checksum)
	println()
	done <- true
}
