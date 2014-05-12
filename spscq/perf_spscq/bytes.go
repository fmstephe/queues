package main

import (
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/fmstephe/queues/spscq"
)

func bqTest(msgCount, msgSize, qSize int64) {
	q := spscq.NewByteQ(qSize)
	done := make(chan bool)
	f, err := os.Create("bq.prof")
	if err != nil {
		panic(err.Error())
	}
	pprof.StartCPUProfile(f)
	go bqDequeue(msgCount, msgSize, q, done)
	go bqEnqueue(msgCount, msgSize, q, done)
	<-done
	<-done
	pprof.StopCPUProfile()
}

func bqEnqueue(msgCount, msgSize int64, q *spscq.ByteQ, done chan bool) {
	runtime.LockOSThread()
	writeBuffer := make([]byte, msgSize)
	for i := int64(0); i < msgCount; i++ {
		writeBuffer[0] = byte(i)
		for w := false; w == false; w = q.Write(writeBuffer) {
		}
	}
	done <- true
}

func bqDequeue(msgCount, msgSize int64, q *spscq.ByteQ, done chan bool) {
	runtime.LockOSThread()
	start := time.Now().UnixNano()
	readBuffer := make([]byte, msgSize)
	sum := int64(0)
	checksum := int64(0)
	for i := int64(0); i < msgCount; i++ {
		for r := false; r == false; r = q.Read(readBuffer) {
		}
		sum += int64(readBuffer[0])
		checksum += int64(byte(i))
	}
	nanos := time.Now().UnixNano() - start
	printTimings(nanos, "bq")
	expect(sum, checksum)
	done <- true
}
