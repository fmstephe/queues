package main

import (
	"runtime"
	"time"
	"fmt"
	"os"
	"runtime/pprof"
	"github.com/fmstephe/queues/oneoneq"
)

func main() {
	runtime.GOMAXPROCS(4)
	var itemCount int64 = 100 * 1000 * 1000
	q := oneoneq.New(1024 * 1024 * 1024, 8)
	done := make(chan bool)
	f, err := os.Create("cpu.prof")
	if err != nil {
		panic(err.Error())
	}
	pprof.StartCPUProfile(f)
	go dequeue(itemCount, q, done)
	go enqueue(itemCount, q, done)
	<-done
	println("First finished")
	<-done
	println("Second finished")
	pprof.StopCPUProfile()
}

func enqueue(num int64, q *oneoneq.Q, done chan bool) {
	runtime.LockOSThread()
	println("Entering Enqueue")
	bs := []byte{1,2,3,4,5,6,7,8}
	for i := int64(0); i < num; i++ {
		bs[0] = byte(i)
		var b []byte
		for b == nil {
			b = q.StartWrite()
		}
		copy(b, bs)
		q.FinishWrite()
	}
	done <- true
}

func dequeue(num int64, q *oneoneq.Q, done chan bool) {
	runtime.LockOSThread()
	start := time.Now().UnixNano()
	println("Entering Dequeue")
	sum := int64(0)
	checksum := int64(0)
	for i := int64(0); i < num; i++ {
		var b []byte
		for b == nil {
			b = q.StartRead()
		}
		sum += int64(b[0])
		checksum += int64(byte(i))
		q.FinishRead()
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
