package main

import (
	"fmt"
	"github.com/fmstephe/queues/spscq"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
	"flag"
)

var (
	llq = flag.Bool("llq", false, "Runs LLChunkQ")
	chunkq = flag.Bool("chunkq", false, "Runs ChunkQ")
	byteq = flag.Bool("byteq", false, "Runs ByteQ")
	millionMsgs = flag.Int64("mm", 66, "The number of messages (in millions) to send")
)

func main() {
	flag.Parse()
	var (
		bytesSize int64 = 63
		chunkSize int64 = 64
		queueSize int64 = 1024 * 1024
		msgCount int64 = (*millionMsgs) * 1000 * 1000
	)
	if *llq {
		runtime.GC()
		llw, llr, name := llchunkTest(msgCount, queueSize, chunkSize)
		perfTest(llw, llr, name)
	}
	if *chunkq {
		runtime.GC()
		cw, cr, name := chunkTest(msgCount, queueSize, chunkSize)
		perfTest(cw, cr, name)
	}
	if *byteq {
		runtime.GC()
		bw, br, name := byteqTest(msgCount, queueSize, bytesSize)
		perfTest(bw, br, name)
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

func llchunkTest(msgCount, qSize, chunkSize int64) (writerf func(), readerf func(), name string) {
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
	return writerf, readerf, "llq"
}

func chunkTest(msgCount, qSize, chunkSize int64) (writerf func(), readerf func(), name string) {
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
	return writerf, readerf, "chunkq"
}

func byteqTest(msgCount, qSize, byteSize int64) (writerf func(), readerf func(), name string) {
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
	return writerf, readerf, "byteq"
}

func expect(sum, checksum int64) {
	if sum != checksum {
		print(fmt.Sprintf("Sum does not match checksum. sum = %d, checksum = %d", sum, checksum))
	}
}
