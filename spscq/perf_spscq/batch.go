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
	if batchSize > 1 {
		batchqBatchEnqueue(msgCount, q, batchSize, done)
	} else {
		batchqSingleEnqueue(msgCount, q, batchSize, done)
	}
}

func batchqBatchEnqueue(msgCount int64, q *spscq.BatchQ, batchSize int64, done chan bool) {
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
				q.CommitWriteBuffer(int64(i))
				break OUTER
			}
			buffer[i] = unsafe.Pointer(uintptr(uint(t)))
		}
		q.CommitWriteBuffer(int64(len(buffer)))
		if t == msgCount {
			break
		}
	}
	done <- true
}

func batchqSingleEnqueue(msgCount int64, q *spscq.BatchQ, batchSize int64, done chan bool) {
	runtime.LockOSThread()
	t := 1
	var v unsafe.Pointer
	for i := int64(0); i < msgCount; i++ {
		v = unsafe.Pointer(uintptr(uint(t)))
		w := q.WriteSingle(v)
		for w == false {
			w = q.WriteSingle(v)
		}
		t++
	}
	done <- true
}

func batchqDequeue(msgCount int64, q *spscq.BatchQ, batchSize int64, done chan bool) {
	if batchSize > 1 {
		batchqBatchDequeue(msgCount, q, batchSize, done)
	} else {
		batchqSingleDequeue(msgCount, q, batchSize, done)
	}
}

func batchqBatchDequeue(msgCount int64, q *spscq.BatchQ, batchSize int64, done chan bool) {
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
				q.CommitReadBuffer(int64(i))
				break OUTER
			}
			sum += int64(uintptr(buffer[i]))
			checksum += t
		}
		q.CommitReadBuffer(int64(len(buffer)))
		if t == msgCount {
			break
		}
	}
	nanos := time.Now().UnixNano() - start
	printTimings(msgCount, nanos, "batchq")
	expect(sum, checksum)
	done <- true
}

func batchqSingleDequeue(msgCount int64, q *spscq.BatchQ, batchSize int64, done chan bool) {
	runtime.LockOSThread()
	start := time.Now().UnixNano()
	sum := int64(0)
	checksum := int64(0)
	var v unsafe.Pointer
	for i := int64(0); i < msgCount; i++ {
		v = q.ReadSingle()
		for v == nil {
			v = q.ReadSingle()
		}
		sum += int64(uintptr(v))
		checksum += i + 1
	}
	nanos := time.Now().UnixNano() - start
	printTimings(msgCount, nanos, "batchq")
	expect(sum, checksum)
	done <- true
}
