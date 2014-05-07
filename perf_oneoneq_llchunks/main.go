package main

import (
	"fmt"
	"github.com/fmstephe/queues/spscq"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

const chunk = 64
const queue = 1024 * 1024
const itemCount int64 = 100 * 1000 * 1000

func main() {
	runtime.GOMAXPROCS(4)
	q := spscq.NewLLChunkQ(queue, chunk)
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

func enqueue(num int64, q *spscq.LLChunkQ, done chan bool) {
	runtime.LockOSThread()
	for i := int64(0); i < num; i++ {
		writeBuffer := q.WriteBuffer()
		for writeBuffer == nil {
			writeBuffer = q.WriteBuffer()
		}
		writeBuffer[0] = byte(i)
		q.CommitWrite()
	}
	done <- true
}

func dequeue(num int64, q *spscq.LLChunkQ, done chan bool) {
	runtime.LockOSThread()
	start := time.Now().UnixNano()
	sum := int64(0)
	checksum := int64(0)
	for i := int64(0); i < num; i++ {
		readBuffer := q.ReadBuffer()
		for readBuffer == nil {
			readBuffer = q.ReadBuffer()
		}
		sum += int64(readBuffer[0])
		checksum += int64(byte(i))
		q.CommitRead()
	}
	print(fmt.Sprintf("sum      %d\nchecksum %d\n", sum, checksum))
	println()
	total := time.Now().UnixNano() - start
	nanos := total
	micros := total / 1000
	millis := micros / 1000
	seconds := millis / 1000
	print(fmt.Sprintf("\nNanos   %d\nMicros  %d\nMillis  %d\nSeconds %d\n", nanos, micros, millis, seconds))
	done <- true
}
