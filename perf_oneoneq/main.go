package main

import (
	"runtime"
	"time"
	"fmt"
	"os"
	"runtime/pprof"
	"github.com/fmstephe/queues/oneoneq"
)

const chunk = 8
const queue = 1024 * 1024

func main() {
	runtime.GOMAXPROCS(4)
	var itemCount int64 = 100 * 1000 * 1000
	q := oneoneq.New(queue, chunk)
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

func enqueue(num int64, q *oneoneq.Q, done chan bool) {
	runtime.LockOSThread()
	writeBuffer := q.WriteBuffer()
	for i := int64(0); i < num; i++ {
		writeBuffer[0] = byte(i)
		for w := int64(0); w == 0; w = q.Write() {}
	}
	done <- true
}

func dequeue(num int64, q *oneoneq.Q, done chan bool) {
	runtime.LockOSThread()
	start := time.Now().UnixNano()
	readBuffer := q.ReadBuffer()
	sum := int64(0)
	checksum := int64(0)
	for i := int64(0); i < num; i++ {
		for r := int64(0); r == 0; r = q.Read() {}
		sum += int64(readBuffer[0])
		checksum += int64(byte(i))
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
