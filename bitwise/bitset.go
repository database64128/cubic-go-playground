package bitwise

const (
	blockBitLog = 6                // 1<<6 == 64 bits
	blockBits   = 1 << blockBitLog // must be power of 2
	bitMask     = blockBits - 1
	bitSetSize  = 1024
	blocks      = bitSetSize / 64
	blockMask   = blocks - 1
)

// BitSet is a bitset with a capacity of 1024 bits.
type BitSet struct {
	blocks [blocks]uint64
}

// IsSet returns whether the bit at position pos is 1.
func (s *BitSet) IsSet(pos int) bool {
	blockIndex := pos >> blockBitLog
	bitIndex := pos & bitMask
	return s.blocks[blockIndex]>>uint64(bitIndex)&1 == 1
}

// Set sets the bit at position pos to 1.
func (s *BitSet) Set(pos int) {
	blockIndex := pos >> blockBitLog
	bitIndex := pos & bitMask
	s.blocks[blockIndex] |= 1 << bitIndex
}

// Reset sets the bit at position pos to 0.
func (s *BitSet) Reset(pos int) {
	blockIndex := pos >> blockBitLog
	bitIndex := pos & bitMask
	s.blocks[blockIndex] &= ^(1 << bitIndex)
}

// Flip flips the bit at position pos.
func (s *BitSet) Flip(pos int) {
	blockIndex := pos >> blockBitLog
	bitIndex := pos & bitMask
	s.blocks[blockIndex] ^= 1 << bitIndex
}

// SetAll sets all bits in the bitset to 1.
func (s *BitSet) SetAll() {
	for i := range s.blocks {
		s.blocks[i] = 1<<64 - 1
	}
}

// ResetAll sets all bits in the bitset to 0.
func (s *BitSet) ResetAll() {
	var newBlocks [blocks]uint64
	s.blocks = newBlocks
}

// FlipAll flips all bits in the bitset.
func (s *BitSet) FlipAll() {
	for i := range s.blocks {
		s.blocks[i] = ^s.blocks[i]
	}
}
