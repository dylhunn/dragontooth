package eval

import (
	"github.com/dylhunn/dragontoothmg"
	"math/bits"
)

const pawnValue = 100
const knightValue = 320
const bishopValue = 330
const rookValue = 500
const queenValue = 900

// Return a static evaluation, relative to the side to move.
func Evaluate(b *dragontoothmg.Board) int16 {
	var score int
	score += countMaterial(&b.White)
	score -= countMaterial(&b.Black)
	if !b.Wtomove {
		score = -score
	}
	return int16(score)
}

func countMaterial(bb *dragontoothmg.Bitboards) int {
	var score int
	score += bits.OnesCount64(bb.Pawns) * pawnValue
	score += bits.OnesCount64(bb.Knights) * knightValue
	score += bits.OnesCount64(bb.Bishops) * bishopValue
	score += bits.OnesCount64(bb.Rooks) * rookValue
	score += bits.OnesCount64(bb.Queens) * queenValue
	return score
}