package bitwise

import (
	"math/rand"
	"strconv"
	"testing"
	"unsafe"
)

type referenceSlidingWindowFilter struct {
	last int
	bits []bool
	temp []bool
}

func newReferenceSlidingWindowFilter(size int) *referenceSlidingWindowFilter {
	return &referenceSlidingWindowFilter{
		bits: make([]bool, size),
		temp: make([]bool, size),
	}
}

func (f *referenceSlidingWindowFilter) Reset() {
	f.last = 0
	fillFalse(f.bits)
}

func (f *referenceSlidingWindowFilter) IsOk(counter int) bool {
	diff := f.last - counter

	// Accept counter if it is ahead of window.
	if diff < 0 {
		return true
	}

	// Reject counter if it is behind window.
	if diff >= len(f.bits) {
		return false
	}

	// Within window, accept if not seen before.
	return !f.bits[diff]
}

func (f *referenceSlidingWindowFilter) MustAdd(counter int) {
	diff := f.last - counter

	// Ahead of window, rotate bits.
	if diff < 0 {
		f.moveAhead(-diff)
		f.last = counter
		diff = 0
	}

	f.bits[diff] = true
}

func (f *referenceSlidingWindowFilter) moveAhead(count int) {
	copy(f.temp[count:], f.bits[:len(f.bits)-count])
	f.bits, f.temp = f.temp, f.bits
	fillFalse(f.temp)
}

func fillFalse(bits []bool) {
	for len(bits) >= blockBits/8 {
		*(*uint)(unsafe.Pointer(&bits[0])) = 0
		bits = bits[blockBits/8:]
	}

	for i := range bits {
		bits[i] = false
	}
}

func FuzzIsOkMustAdd(f *testing.F) {
	size := int(rand.Int31())
	filter := NewSlidingWindowFilter(uint64(size))
	ref := newReferenceSlidingWindowFilter(size)

	f.Add(size / 8)
	f.Add(size / 4)
	f.Add(size / 2)
	f.Add(size)

	f.Fuzz(func(t *testing.T, counter int) {
		if ok := ref.IsOk(counter); ok != filter.IsOk(uint64(counter)) {
			t.Error(counter, "should be", ok)
		}
		filter.MustAdd(uint64(counter))
		ref.MustAdd(counter)
	})
}

func FuzzAdd(f *testing.F) {
	size := int(rand.Int31())
	filter := NewSlidingWindowFilter(uint64(size))
	ref := newReferenceSlidingWindowFilter(size)

	f.Add(size / 8)
	f.Add(size / 4)
	f.Add(size / 2)
	f.Add(size)

	f.Fuzz(func(t *testing.T, counter int) {
		if ok := ref.IsOk(counter); ok != filter.Add(uint64(counter)) {
			t.Error(counter, "should be", ok)
		}
		ref.MustAdd(counter)
	})
}

func testIsOkMustAdd(t *testing.T, f *SlidingWindowFilter) {
	f.Reset()
	i := uint64(1)
	n := uint64(len(f.ring)+1) * blockBits

	// Add 1, 3, 5, ..., n-1.
	for ; i < n; i += 2 {
		if !f.IsOk(i) {
			t.Error(i, "should be ok.")
		}
		f.MustAdd(i)
	}

	// Check 0, 2, 4, ..., 126.
	for i = 0; i < n-f.size; i += 2 {
		if f.IsOk(i) {
			t.Error(i, "should not be ok.")
		}
	}

	// Check 128, 130, 132, ..., n-2.
	for ; i < n; i += 2 {
		if !f.IsOk(i) {
			t.Error(i, "should be ok.")
		}
	}

	// Check 1, 3, 5, ..., n-1.
	for i = 1; i < n; i += 2 {
		if f.IsOk(i) {
			t.Error(i, "should not be ok.")
		}
	}

	// Roll over the window.
	n <<= 1
	if !f.IsOk(n) {
		t.Error(n, "should be ok.")
	}
	f.MustAdd(n)

	// Check behind window.
	for i = 0; i < n-f.size+1; i++ {
		if f.IsOk(i) {
			t.Error(i, "should not be ok.")
		}
	}

	// Check within window.
	for ; i < n; i++ {
		if !f.IsOk(i) {
			t.Error(i, "should be ok.")
		}
	}

	// Check n.
	if i == n {
		if f.IsOk(i) {
			t.Error(i, "should not be ok.")
		}
		i++
	}

	// Check after window.
	for ; i < n+f.size; i++ {
		if !f.IsOk(i) {
			t.Error(i, "should be ok.")
		}
	}
}

func testAdd(t *testing.T, f *SlidingWindowFilter) {
	f.Reset()
	i := uint64(1)
	n := uint64(len(f.ring)+1) * blockBits

	// Add 1, 3, 5, ..., n-1.
	for ; i < n; i += 2 {
		if !f.Add(i) {
			t.Error(i, "should succeed.")
		}
	}

	// Check 0, 2, 4, ..., 126.
	for i = 0; i < n-f.size; i += 2 {
		if f.Add(i) {
			t.Error(i, "should fail.")
		}
	}

	// Check 128, 130, 132, ..., n-2.
	for ; i < n; i += 2 {
		if !f.Add(i) {
			t.Error(i, "should succeed.")
		}
	}

	// Check 1, 3, 5, ..., n-1.
	for i = 1; i < n; i += 2 {
		if f.Add(i) {
			t.Error(i, "should fail.")
		}
	}

	// Roll over the window.
	n <<= 1
	if !f.Add(n) {
		t.Error(n, "should succeed.")
	}

	// Check behind window.
	for i = 0; i < n-f.size+1; i++ {
		if f.Add(i) {
			t.Error(i, "should fail.")
		}
	}

	// Check within window.
	for ; i < n; i++ {
		if !f.Add(i) {
			t.Error(i, "should succeed.")
		}
	}

	// Check n.
	if i == n {
		if f.Add(i) {
			t.Error(i, "should fail.")
		}
		i++
	}

	// Check after window.
	for ; i < n+f.size; i++ {
		if !f.Add(i) {
			t.Error(i, "should succeed.")
		}
	}
}

func testReset(t *testing.T, f *SlidingWindowFilter) {
	n := f.Size() * 2

	for i := uint64(0); i < n; i++ {
		f.MustAdd(i)
	}

	f.Reset()

	for i := uint64(0); i < n; i++ {
		if !f.IsOk(i) {
			t.Error(i, "should be ok.")
		}
	}
}

func testSlidingWindowFilter(t *testing.T, f *SlidingWindowFilter) {
	t.Run("IsOkMustAdd", func(t *testing.T) {
		testIsOkMustAdd(t, f)
	})
	t.Run("Add", func(t *testing.T) {
		testAdd(t, f)
	})
	t.Run("Reset", func(t *testing.T) {
		testReset(t, f)
	})
}

func TestSlidingWindowFilter(t *testing.T) {
	sizes := []uint64{0, 1, 2, 31, 32, 33, 63, 64, 65, 127, 128, 129, 255, 256, 257}
	for _, size := range sizes {
		t.Run(strconv.FormatUint(size, 10), func(t *testing.T) {
			f := NewSlidingWindowFilter(size)
			t.Log("ringBlockIndexMask", f.ringBlockIndexMask)
			testSlidingWindowFilter(t, f)
		})
	}
}
