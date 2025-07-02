//go:build rp

package ringbuffer

import (
	"runtime/volatile"

	"golang.org/x/exp/constraints"
)

const (
	bufferSize = 64
)

type Number interface {
	constraints.Integer | constraints.Float | constraints.Complex
}

type RingBuffer[T Number] struct {
	buffer [bufferSize]T
	head   volatile.Register16
	tail   volatile.Register16
}

// NewRingBuffer returns a new ring buffer.
func NewRingBuffer[T Number]() *RingBuffer[T] {
	return &RingBuffer[T]{}
}

// Used returns how many bytes in buffer have been used.
func (rb *RingBuffer[T]) Used() uint16 {
	return uint16(rb.head.Get() - rb.tail.Get())
}

// Full returns true if the buffer is full.
func (rb *RingBuffer[T]) Full() bool {
	return rb.Used() == bufferSize
}

// Put stores a value in the buffer. If the buffer is already
// full, the method will return false.
func (rb *RingBuffer[T]) Put(val T) bool {
	if !rb.Full() {
		rb.head.Set(rb.head.Get() + 1)
		rb.buffer[rb.head.Get()%bufferSize] = val
		return true
	}
	return false
}

// Get returns a byte from the buffer. If the buffer is empty,
// the method will return a false as the second value.
func (rb *RingBuffer[T]) Get() (T, bool) {
	if rb.Used() != 0 {
		rb.tail.Set(rb.tail.Get() + 1)
		return rb.buffer[rb.tail.Get()%bufferSize], true
	}
	return *new(T), false
}

// Clear resets the head and tail pointer to zero.
func (rb *RingBuffer[T]) Clear() {
	rb.head.Set(0)
	rb.tail.Set(0)
}
