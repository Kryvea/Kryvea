package safe_test

import (
	"sync"
	"testing"

	"github.com/Kryvea/Kryvea/internal/safe"
)

// TestNew verifies that New creates an independent copy of the provided slice.
func TestNew(t *testing.T) {
	original := []int{1, 2, 3}
	s := safe.New(original)

	// Mutate the original and verify SafeSlice is not affected.
	original[0] = 999
	got := s.GetAll()
	if got[0] != 1 {
		t.Fatalf("expected SafeSlice copy isolation, got %v", got)
	}
}

// TestAppend ensures that Append correctly appends an element.
func TestAppend(t *testing.T) {
	s := safe.New([]string{"a"})
	s.Append("b")
	result := s.GetAll()

	if len(result) != 2 || result[1] != "b" {
		t.Fatalf("expected appended slice ['a','b'], got %v", result)
	}
}

// TestAppendAll validates multiple-element appending.
func TestAppendAll(t *testing.T) {
	s := safe.New([]int{1})
	s.AppendAll(2, 3, 4)
	result := s.GetAll()

	expected := []int{1, 2, 3, 4}
	for i, v := range expected {
		if result[i] != v {
			t.Fatalf("expected %v, got %v", expected, result)
		}
	}
}

// TestLen ensures Len reports accurate length under concurrent use.
func TestLen(t *testing.T) {
	s := safe.New([]int{})
	for i := 0; i < 10; i++ {
		s.Append(i)
	}
	if l := s.Len(); l != 10 {
		t.Fatalf("expected length 10, got %d", l)
	}
}

// TestConcurrentAppend checks thread safety using multiple goroutines.
func TestConcurrentAppend(t *testing.T) {
	s := safe.New([]int{})
	wg := sync.WaitGroup{}
	const goroutines = 100
	const perGoroutine = 100

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < perGoroutine; j++ {
				s.Append(id*perGoroutine + j)
			}
		}(i)
	}
	wg.Wait()

	total := s.Len()
	if total != goroutines*perGoroutine {
		t.Fatalf("expected %d elements, got %d", goroutines*perGoroutine, total)
	}
}

func TestSet(t *testing.T) {
	s := safe.New([]int{10, 20, 30})

	// Replace the middle element
	s.Set(1, 25)
	result := s.GetAll()

	if result[1] != 25 {
		t.Fatalf("expected index 1 to be 25, got %d", result[1])
	}

	// Ensure other elements remain unchanged
	expected := []int{10, 25, 30}
	for i, v := range expected {
		if result[i] != v {
			t.Fatalf("expected %v, got %v", expected, result)
		}
	}

	// Test out-of-bounds (should safely ignore)
	s.Set(-1, 99)
	s.Set(3, 99) // len == 3, last index = 2
	after := s.GetAll()

	if len(after) != 3 {
		t.Fatalf("expected length unchanged at 3, got %d", len(after))
	}
	if after[0] == 99 || after[2] == 99 {
		t.Fatal("Set should not modify slice on out-of-bounds index")
	}
}

// TestSetConcurrent verifies thread safety under concurrent writes.
func TestSetConcurrent(t *testing.T) {
	s := safe.New(make([]int, 100))
	wg := sync.WaitGroup{}
	wg.Add(100)

	for i := 0; i < 100; i++ {
		go func(i int) {
			defer wg.Done()
			s.Set(i, i*10)
		}(i)
	}

	wg.Wait()

	data := s.GetAll()
	for i, v := range data {
		if v != i*10 {
			t.Fatalf("at index %d: expected %d, got %d", i, i*10, v)
		}
	}
}

// TestImmutability ensures GetAll returns a copy, not a shared reference.
func TestImmutability(t *testing.T) {
	s := safe.New([]int{1, 2, 3})
	data := s.GetAll()
	data[0] = 999 // modify external copy

	got := s.GetAll()
	if got[0] == 999 {
		t.Fatal("GetAll should return a copy, not a reference to internal data")
	}
}

// TestTypeSafety verifies generic instantiation for different types.
func TestTypeSafety(t *testing.T) {
	sInt := safe.New([]int{})
	sStr := safe.New([]string{})
	sInt.Append(42)
	sStr.Append("x")

	if sInt.Len() != 1 || sStr.Len() != 1 {
		t.Fatal("expected both SafeSlices to be independent and type-safe")
	}
}
