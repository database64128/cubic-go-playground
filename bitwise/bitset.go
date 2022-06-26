package bitwise

const (
	bitSetBlockBitLog = 6                      // 1<<6 == 64 bits
	bitSetBlockBits   = 1 << bitSetBlockBitLog // must be power of 2
	bitSetbitMask     = bitSetBlockBits - 1
	bitSetSize        = 1024
	bitSetBlocks      = bitSetSize / 64
)

// BitSet is a bitset with a capacity of 1024 bits.
type BitSet struct {
	blocks [bitSetBlocks]uint64
}

// IsSet returns whether the bit at position pos is 1.
func (s *BitSet) IsSet(pos int) bool {
	blockIndex := pos >> bitSetBlockBitLog
	bitIndex := pos & bitSetbitMask
	return s.blocks[blockIndex]>>uint64(bitIndex)&1 == 1
}

// Set sets the bit at position pos to 1.
func (s *BitSet) Set(pos int) {
	blockIndex := pos >> bitSetBlockBitLog
	bitIndex := pos & bitSetbitMask
	s.blocks[blockIndex] |= 1 << bitIndex
}

// Reset sets the bit at position pos to 0.
func (s *BitSet) Reset(pos int) {
	blockIndex := pos >> bitSetBlockBitLog
	bitIndex := pos & bitSetbitMask
	s.blocks[blockIndex] &= ^(1 << bitIndex)
}

// Flip flips the bit at position pos.
func (s *BitSet) Flip(pos int) {
	blockIndex := pos >> bitSetBlockBitLog
	bitIndex := pos & bitSetbitMask
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
	var newBlocks [bitSetBlocks]uint64
	s.blocks = newBlocks
}

// FlipAll flips all bits in the bitset.
func (s *BitSet) FlipAll() {
	for i := range s.blocks {
		s.blocks[i] = ^s.blocks[i]
	}
}
