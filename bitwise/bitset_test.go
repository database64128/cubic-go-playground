package bitwise

import "testing"

func TestBitSetSetResetFlip(t *testing.T) {
	var bitSet BitSet

	// Every bit should be unset in an initialized bitset.
	for i := 0; i < bitSetSize; i++ {
		if bitSet.IsSet(i) {
			t.Errorf("Pos %d is set, should be unset.", i)
		}
	}

	// Set pos 128 to true.
	bitSet.Set(128)

	// Flip pos 256 (to true).
	bitSet.Flip(256)

	// 0-127 should be unset.
	for i := 0; i < 128; i++ {
		if bitSet.IsSet(i) {
			t.Errorf("Pos %d is set, should be unset.", i)
		}
	}

	// 128 should be set.
	if !bitSet.IsSet(128) {
		t.Error("Pos 128 is unset, should be set.")
	}

	// 129-255 should be unset.
	for i := 129; i < 256; i++ {
		if bitSet.IsSet(i) {
			t.Errorf("Pos %d is set, should be unset.", i)
		}
	}

	// 256 should be set.
	if !bitSet.IsSet(256) {
		t.Error("Pos 256 is unset, should be set.")
	}

	// 257-1023 should be unset.
	for i := 257; i < bitSetSize; i++ {
		if bitSet.IsSet(i) {
			t.Errorf("Pos %d is set, should be unset.", i)
		}
	}

	// Reset pos 128 and 256.
	bitSet.Reset(128)
	bitSet.Reset(256)

	// All bits should be unset.
	for i := 0; i < bitSetSize; i++ {
		if bitSet.IsSet(i) {
			t.Errorf("Pos %d is set, should be unset.", i)
		}
	}
}

func TestBitSetSetAllResetAllFlipAll(t *testing.T) {
	var bitSet BitSet

	// Set all.
	bitSet.SetAll()

	// All bits should be set.
	for i := 0; i < bitSetSize; i++ {
		if !bitSet.IsSet(i) {
			t.Errorf("Pos %d is unset, should be set.", i)
		}
	}

	// Reset all.
	bitSet.ResetAll()

	// All bits should be unset.
	for i := 0; i < bitSetSize; i++ {
		if bitSet.IsSet(i) {
			t.Errorf("Pos %d is set, should be unset.", i)
		}
	}

	// Flip all.
	bitSet.FlipAll()

	// All bits should be set.
	for i := 0; i < bitSetSize; i++ {
		if !bitSet.IsSet(i) {
			t.Errorf("Pos %d is unset, should be set.", i)
		}
	}
}
