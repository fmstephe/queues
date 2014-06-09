package main

import (
	"flag"
	"fmt"
	"runtime"

	"github.com/fmstephe/fstrconv"
)

var (
	all         = flag.Bool("all", false, "Runs all queue tests")
	llq         = flag.Bool("llq", false, "Runs LLChunkQ")
	cq          = flag.Bool("cq", false, "Runs ChunkQ")
	bq          = flag.Bool("bq", false, "Runs ByteQ")
	pq          = flag.Bool("pq", false, "Runs PointerQ")
	millionMsgs = flag.Int64("mm", 10, "The number of messages (in millions) to send")
	bytesSize   = flag.Int64("bytesSize", 63, "The number of bytes to read/write in ByteQ")
	chunkSize   = flag.Int64("chunkSize", 64, "The number of bytes to read/write in LLChunkQ and ChunkQ")
	qSize       = flag.Int64("qSize", 1024*1024, "The size of the queue")
)

func main() {
	runtime.GOMAXPROCS(4)
	flag.Parse()
	var msgCount int64 = (*millionMsgs) * 1000 * 1000
	if *llq || *all {
		runtime.GC()
		llqTest(msgCount, *chunkSize, *qSize)
	}
	if *cq || *all {
		runtime.GC()
		cqTest(msgCount, *chunkSize, *qSize)
	}
	if *bq || *all {
		runtime.GC()
		bqTest(msgCount, *bytesSize, *qSize)
	}
	if *pq || *all {
		runtime.GC()
		pqTest(msgCount, *qSize)
	}
}

func printTimings(msgCount, nanos int64, name string) {
	micros := nanos / 1000
	millis := micros / 1000
	seconds := millis / 1000
	print(fmt.Sprintf("\n%s\n%s\nNanos   %d\nMicros  %d\nMillis  %d\nSeconds %d\n", name, fstrconv.Itoa64Comma(msgCount), nanos, micros, millis, seconds))
}

func expect(sum, checksum int64) {
	if sum != checksum {
		print(fmt.Sprintf("Sum does not match checksum. sum = %d, checksum = %d", sum, checksum))
	}
}
