package main

import (
	"os"
	"runtime"
	"runtime/pprof"
	"time"
	"unsafe"

	"github.com/fmstephe/queues/spscq"
)

var value int64
var val unsafe.Pointer

func init() {
	value = 777
	val = unsafe.Pointer(&value)
}

func pqTest(msgCount, qSize int64) {
	q := spscq.NewPointerQ(qSize)
	done := make(chan bool)
	f, err := os.Create("pq.prof")
	if err != nil {
		panic(err.Error())
	}
	pprof.StartCPUProfile(f)
	go dequeue(msgCount, q, done)
	go enqueue(msgCount, q, done)
	<-done
	<-done
	pprof.StopCPUProfile()
}

func enqueue(num int64, q *spscq.PointerQ, done chan bool) {
	runtime.LockOSThread()
	for i := int64(0); i < num; i++ {
		for w := false; w == false; w = q.Write(val) {
		}
	}
	done <- true
}

func dequeue(num int64, q *spscq.PointerQ, done chan bool) {
	runtime.LockOSThread()
	start := time.Now().UnixNano()
	sum := int64(0)
	checksum := int64(0)
	for i := int64(0); i < num; i++ {
		r := unsafe.Pointer(nil)
		for r == nil {
			r = q.Read()
		}
		sum += *((*int64)(r))
		checksum += value
	}
	nanos := time.Now().UnixNano() - start
	printTimings(nanos, "pq")
	expect(sum, checksum)
	done <- true
}
