package main

import (
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/fmstephe/queues/spscq"
)

func llqTest(msgCount, msgSize, qSize int64) {
	q := spscq.NewLLChunkQ(qSize, msgSize)
	done := make(chan bool)
	f, err := os.Create("llq.prof")
	if err != nil {
		panic(err.Error())
	}
	pprof.StartCPUProfile(f)
	go llqDequeue(msgCount, q, done)
	go llqEnqueue(msgCount, q, done)
	<-done
	<-done
	pprof.StopCPUProfile()
}

func llqEnqueue(num int64, q *spscq.LLChunkQ, done chan bool) {
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

func llqDequeue(msgCount int64, q *spscq.LLChunkQ, done chan bool) {
	runtime.LockOSThread()
	start := time.Now().UnixNano()
	sum := int64(0)
	checksum := int64(0)
	for i := int64(0); i < msgCount; i++ {
		readBuffer := q.ReadBuffer()
		for readBuffer == nil {
			readBuffer = q.ReadBuffer()
		}
		sum += int64(readBuffer[0])
		checksum += int64(byte(i))
		q.CommitRead()
	}
	nanos := time.Now().UnixNano() - start
	printTimings(msgCount, nanos, "llq")
	expect(sum, checksum)
	done <- true
}
