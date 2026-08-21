//go:build windows

package input

import (
	"context"
	"fmt"
	"sync"
)

type PipeReader struct {
	path     string
	once     sync.Once
	callback func(string)
}

func New(path string, callback func(string)) (*PipeReader, error) {
	return nil, fmt.Errorf("named pipes (FIFO) are not supported on Windows")
}

func (p *PipeReader) Start(ctx context.Context) error {
	return fmt.Errorf("named pipes (FIFO) are not supported on Windows")
}

func (p *PipeReader) Path() string {
	return p.path
}

func (p *PipeReader) Stop() {}
