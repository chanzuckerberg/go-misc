package runner

import (
	"fmt"
	"io"
	"sync/atomic"
)

type Report struct {
	reader io.Reader

	total uint64
}

func (r *Report) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	atomic.AddUint64(&r.total, uint64(n))

	if err == nil {
		fmt.Println("Read", n, "bytes for a total of", r.total)
	}

	return n, err
}
