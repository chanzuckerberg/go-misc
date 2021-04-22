package runner

import (
	"io"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type Report struct {
	reader io.Reader

	total        int
	lastReported time.Time

	mu sync.Mutex
}

func (r *Report) Read(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	n, err := r.reader.Read(p)
	r.total += n

	if time.Since(r.lastReported) >= time.Minute {
		r.lastReported = time.Now()
		logrus.Infof("Read %d bytes so far", r.total)
	}

	if err != nil {
		logrus.Error(err)
	}

	return n, err
}
