package safe

import (
	"sync"
)

// SafeSlice[T] is a thread-safe slice wrapper.
type SafeSlice[T any] struct {
	mu   sync.Mutex
	data []T
}

// New creates a SafeSlice from an existing slice.
func New[T any](initial []T) *SafeSlice[T] {
	cpy := make([]T, len(initial))
	copy(cpy, initial)
	return &SafeSlice[T]{data: cpy}
}

// Append safely appends an element to the slice.
func (s *SafeSlice[T]) Append(v T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = append(s.data, v)
}

// AppendAll safely appends multiple elements.
func (s *SafeSlice[T]) AppendAll(vals ...T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = append(s.data, vals...)
}

// Set safely replaces the element at the given index.
func (s *SafeSlice[T]) Set(index int, v T) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if index < 0 || index >= len(s.data) {
		return
	}

	s.data[index] = v
}

// GetAll returns a copy of the slice to avoid external mutation.
func (s *SafeSlice[T]) GetAll() []T {
	s.mu.Lock()
	defer s.mu.Unlock()
	cpy := make([]T, len(s.data))
	copy(cpy, s.data)
	return cpy
}

// Len returns the current length of the slice.
func (s *SafeSlice[T]) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.data)
}
