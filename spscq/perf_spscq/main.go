package main

import (
	"fmt"
	"github.com/fmstephe/queues/spscq"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
	"flag"
	"unsafe"
)

var (
	llq = flag.Bool("llq", false, "Runs LLChunkQ")
	cq = flag.Bool("cq", false, "Runs ChunkQ")
	bq = flag.Bool("bq", false, "Runs ByteQ")
	pq = flag.Bool("pq", false, "Runs PointerQ")
	millionMsgs = flag.Int64("mm", 10, "The number of messages (in millions) to send")
	bytesSize = flag.Int64("bytesSize", 63, "The number of bytes to read/write in ByteQ")
	chunkSize = flag.Int64("chunkSize", 64, "The number of bytes to read/write in LLChunkQ and ChunkQ")
	qSize = flag.Int64("qSize", 1024 * 1024, "The size of the queue")
)

func main() {
	flag.Parse()
	var msgCount int64 = (*millionMsgs) * 1000 * 1000
	if *llq {
		runtime.GC()
		w, r := llchunkTest(msgCount, *qSize, *chunkSize)
		perfTest(w, r, "llq")
	}
	if *cq {
		runtime.GC()
		w, r := chunkTest(msgCount, *qSize, *chunkSize)
		perfTest(w, r, "cq")
	}
	if *bq {
		runtime.GC()
		w, r := byteqTest(msgCount, *qSize, *bytesSize)
		perfTest(w, r, "bq")
	}
	if *pq {
		runtime.GC()
		w, r := pqTest(msgCount, *qSize)
		perfTest(w, r, "pq")
	}
}

func perfTest(writef, readf func(), name string) {
	runtime.GOMAXPROCS(4)
	done := make(chan bool)
	f, err := os.Create(name + ".cpu.prof")
	if err != nil {
		panic(err.Error())
	}
	pprof.StartCPUProfile(f)
	go writer(writef, done)
	go reader(readf, name, done)
	<-done
	<-done
	pprof.StopCPUProfile()
}

func writer(writef func(), done chan bool) {
	runtime.LockOSThread()
	writef()
	done <- true
}

func reader(readf func(), name string, done chan bool) {
	runtime.LockOSThread()
	start := time.Now().UnixNano()
	readf()
	total := time.Now().UnixNano() - start
	nanos := total
	micros := total / 1000
	millis := micros / 1000
	seconds := millis / 1000
	print(fmt.Sprintf("\n%s\nNanos   %d\nMicros  %d\nMillis  %d\nSeconds %d\n", name, nanos, micros, millis, seconds))
	done <- true
}

func llchunkTest(msgCount, qSize, chunkSize int64) (writerf func(), readerf func()) {
	q := spscq.NewLLChunkQ(qSize, chunkSize)
	writerf = func() {
		for i := int64(0); i < msgCount; i++ {
			writeBuffer := q.WriteBuffer()
			for writeBuffer == nil {
				writeBuffer = q.WriteBuffer()
			}
			writeBuffer[0] = byte(i)
			q.CommitWrite()
		}
	}
	readerf = func() {
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
		expect(sum, checksum)
	}
	return writerf, readerf
}

func chunkTest(msgCount, qSize, chunkSize int64) (writerf func(), readerf func()) {
	q := spscq.NewChunkQ(qSize, chunkSize)
	writerf = func() {
		writeBuffer := q.WriteBuffer()
		for i := int64(0); i < msgCount; i++ {
			writeBuffer[0] = byte(i)
			for w := false; w == false; w = q.Write() {}
		}
	}
	readerf = func() {
		sum := int64(0)
		checksum := int64(0)
		readBuffer := q.ReadBuffer()
		for i := int64(0); i < msgCount; i++ {
			for r := false; r == false; r = q.Read() {}
			sum += int64(readBuffer[0])
			checksum += int64(byte(i))
		}
		expect(sum, checksum)
	}
	return writerf, readerf
}

func byteqTest(msgCount, qSize, byteSize int64) (writerf func(), readerf func()) {
	q := spscq.NewByteQ(qSize)
	writerf = func() {
		writeBuffer := make([]byte, byteSize)
		for i := int64(0); i < msgCount; i++ {
			writeBuffer[0] = byte(i)
			for w := false; w == false; w = q.Write(writeBuffer) {}
		}
	}
	readerf = func() {
		sum := int64(0)
		checksum := int64(0)
		readBuffer := make([]byte, byteSize)
		for i := int64(0); i < msgCount; i++ {
			for r := false; r == false; r = q.Read(readBuffer) {}
			sum += int64(readBuffer[0])
			checksum += int64(byte(i))
		}
		expect(sum, checksum)
	}
	return writerf, readerf
}

func pqTest(msgCount, qSize int64) (writerf func(), readerf func()) {
	q := spscq.NewPointerQ(qSize)
	value := int64(777)
	val := unsafe.Pointer(&value)
	writerf = func() {
		for i := int64(0); i < msgCount; i++ {
			for w := false; w == false; w = q.Write(val) {
			}
		}
	}
	readerf = func() {
		sum := int64(0)
		checksum := int64(0)
		for i := int64(0); i < msgCount; i++ {
			r := unsafe.Pointer(nil)
			for r == nil {
				r = q.Read()
			}
			sum += *((*int64)(r))
			checksum += value
		}
		expect(sum, checksum)
	}
	return writerf, readerf
}

func expect(sum, checksum int64) {
	if sum != checksum {
		print(fmt.Sprintf("Sum does not match checksum. sum = %d, checksum = %d", sum, checksum))
	}
}
