package transtable

import (
	"github.com/dylhunn/dragontoothmg"
)

// Transposition table entry types
const (
	LowerBound = iota // cut-node (fail-high)
	UpperBound = iota // all-node (fail-low)
	Exact      = iota // pv-node
)

// Table data
var keys []uint64
var values []uint64

var DefaultTtableSize int = 512 // 512 MB; can be overridden by user

var slots int
var entries int

func init() {
	Initialize(DefaultTtableSize) // default ttable of 512 MB
}

// Initialize (or reinitialize and clear) the table. Must be called before use.
func Initialize(sizeInMb int) {
	bits := 8 * 1024 * 1024 * sizeInMb
	slots = (bits / 128) + 1 // seek an odd number
	keys = make([]uint64, slots, slots)
	values = make([]uint64, slots, slots)
}

// Key structure, from LSB
// eval result (16 bits)
// move (16 bits)
// age (16 bits): the move of the game on which this position would have occurred
// depth (8 bits)
// node type (8 bits): from the three constants above

// Insert an entry into the transposition table, which requires the board position,
// the best move, the evaluation score, the entry search depth, and the node type.
func Put(b *dragontoothmg.Board, m dragontoothmg.Move, eval int16, depth int8, ntype uint8) {
	var value uint64 = uint64(uint16(eval)) | (uint64(m) << 16) | (uint64(b.Fullmoveno) << 32) |
		(uint64(uint8(depth)) << 48) | (uint64(ntype) << 56)
	hash := b.Hash()
	key := hash ^ value
	index := hash % uint64(len(keys))
	// TODO(dylhunn): Try various probing and replacement strategies
	if keys[index] == 0 {
		entries++
	}
	keys[index] = key
	values[index] = value
}

func Get(b *dragontoothmg.Board) (found bool, move dragontoothmg.Move,
	eval int16, depth int8, ntype uint8) {
	hash := b.Hash()
	index := hash % uint64(len(keys))
	// TODO(dylhunn): Investigate atomics. Atomics might not be necessary on
	// 64-bit platforms, or at all, since we detect corrupted data anyway.
	key := keys[index]
	value := values[index]
	found = (hash == (key ^ value))
	if !found { // TODO(dylhunn): benchmark with and without a branch
		// TODO(dylhunn): Count misses and invalid reads separately
		return false, 0, 0, 0, 0
	}
	eval = int16((value & 0xFFFF))
	move = dragontoothmg.Move((value >> 16) & 0xFFFF)
	depth = int8((value >> 48) & 0xFF)
	ntype = uint8((value >> 56) & 0xFF)
	return
}

func Erase(b *dragontoothmg.Board) {
	hash := b.Hash()
	index := hash % uint64(len(keys))
	keys[index] = 0
	values[index] = 0
}

func Load() float32 {
	return float32(entries) / float32(slots)
}
