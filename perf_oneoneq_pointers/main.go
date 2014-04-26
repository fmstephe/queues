package main

import (
	"runtime"
	"time"
	"fmt"
	"os"
	"runtime/pprof"
	"github.com/fmstephe/queues/oneoneq"
	"unsafe"
)

const chunk = 8
const queue = 1024 * 1024

var value int64
var val unsafe.Pointer

func init() {
	value = 777
	val = unsafe.Pointer(&value)
}

func main() {
	runtime.GOMAXPROCS(4)
	var itemCount int64 = 100 * 1000 * 1000
	q := oneoneq.NewPointerQ(queue)
	done := make(chan bool)
	f, err := os.Create("cpu.prof")
	if err != nil {
		panic(err.Error())
	}
	pprof.StartCPUProfile(f)
	go dequeue(itemCount, q, done)
	go enqueue(itemCount, q, done)
	<-done
	<-done
	pprof.StopCPUProfile()
}

func enqueue(num int64, q *oneoneq.PointerQ, done chan bool) {
	runtime.LockOSThread()
	for i := int64(0); i < num; i++ {
		for w := false; w == false; w = q.Write(val) {}
	}
	done <- true
}

func dequeue(num int64, q *oneoneq.PointerQ, done chan bool) {
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
