package main

import (
	"os"
	"runtime"
	"runtime/pprof"
	"time"
	"unsafe"

	"github.com/fmstephe/queues/spscq"
)

func batchqTest(msgCount, batchSize, qSize int64) {
	q := spscq.NewBatchQ(qSize)
	done := make(chan bool)
	f, err := os.Create("batchq.prof")
	if err != nil {
		panic(err.Error())
	}
	pprof.StartCPUProfile(f)
	go batchqDequeue(msgCount, q, batchSize, done)
	go batchqEnqueue(msgCount, q, batchSize, done)
	<-done
	<-done
	pprof.StopCPUProfile()
}

func batchqEnqueue(msgCount int64, q *spscq.BatchQ, batchSize int64, done chan bool) {
	runtime.LockOSThread()
	var t int64
	var buffer []unsafe.Pointer
OUTER:
	for {
		buffer = q.WriteBuffer(batchSize)
		for buffer == nil {
			buffer = q.WriteBuffer(batchSize)
		}
		for i := range buffer {
			t++
			if t > msgCount {
				q.CommitWrite(int64(i))
				break OUTER
			}
			buffer[i] = unsafe.Pointer(uintptr(uint(t)))
		}
		q.CommitWrite(int64(len(buffer)))
		if t == msgCount {
			break
		}
	}
	done <- true
}

func batchqDequeue(msgCount int64, q *spscq.BatchQ, batchSize int64, done chan bool) {
	runtime.LockOSThread()
	start := time.Now().UnixNano()
	var sum int64
	var checksum int64
	var t int64
	var buffer []unsafe.Pointer
OUTER:
	for {
		buffer = q.ReadBuffer(batchSize)
		for buffer == nil {
			buffer = q.ReadBuffer(batchSize)
		}
		for i := range buffer {
			t++
			if t > msgCount {
				q.CommitRead(int64(i))
				break OUTER
			}
			sum += int64(uintptr(buffer[i]))
			checksum += t
		}
		q.CommitRead(int64(len(buffer)))
		if t == msgCount {
			break
		}
	}
	nanos := time.Now().UnixNano() - start
	printTimings(msgCount, nanos, "batchq")
	expect(sum, checksum)
	done <- true
}
