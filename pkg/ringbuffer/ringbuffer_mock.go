//go:build !rp

package ringbuffer

import (
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
	head   uint16
	tail   uint16
}

// NewRingBuffer returns a new ring buffer.
func NewRingBuffer[T Number]() *RingBuffer[T] {
	return &RingBuffer[T]{}
}

// Used returns how many bytes in buffer have been used.
func (rb *RingBuffer[T]) Used() uint16 {
	return uint16(rb.head - rb.tail)
}

// Full returns true if the buffer is full.
func (rb *RingBuffer[T]) Full() bool {
	return rb.Used() == bufferSize
}

// Put stores a value in the buffer. If the buffer is already
// full, the method will return false.
func (rb *RingBuffer[T]) Put(val T) bool {
	if !rb.Full() {
		rb.head++
		rb.buffer[rb.head%bufferSize] = val
		return true
	}
	return false
}

// Get returns a byte from the buffer. If the buffer is empty,
// the method will return a false as the second value.
func (rb *RingBuffer[T]) Get() (T, bool) {
	if rb.Used() != 0 {
		rb.tail++
		return rb.buffer[rb.tail%bufferSize], true
	}
	return *new(T), false
}

// Clear resets the head and tail pointer to zero.
func (rb *RingBuffer[T]) Clear() {
	rb.head = 0
	rb.tail = 0
}
