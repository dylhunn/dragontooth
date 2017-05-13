package transtable

import "github.com/dylhunn/dragontoothmg"

// Transposition table entry types
const (
	LowerBound = iota // cut-node (fail-high)
	UpperBound = iota // all-node (fail-low)
	Exact      = iota // pv-node
)

// Key structure, from LSB
// eval result (16 bits)
// move (16 bits)
// age (16 bits): the move of the game on which this position would have occurred
// depth (8 bits)
// node type (8 bits): from the three constants above

// Insert an entry into the transposition table, which requires the board position,
// the best move, the evaluation score, the entry search depth, and the node type.
func Put(b *dragontoothmg.Board, m dragontoothmg.Move, eval int16, depth uint8, ntype uint8) {
	var value uint64 = uint64(eval) & (uint64(m) << 16) & (uint64(b.Fullmoveno) << 32) &
		(uint64(depth) << 48) & (uint64(ntype) << 56)
	key := b.Hash() & value
	_ = key
}
