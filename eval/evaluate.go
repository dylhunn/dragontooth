package eval

import (
	"github.com/dylhunn/dragontoothmg"
	"math/bits"
)

func Evaluate(b *dragontoothmg.Board) int16 {
	var score int
	score += countMaterial(&b.White)
	score -= countMaterial(&b.Black)
	return int16(score)
}

func countMaterial(bb *dragontoothmg.Bitboards) int {
	var score int
	score += bits.OnesCount64(bb.Pawns)
	score += bits.OnesCount64(bb.Knights) * 3
	score += bits.OnesCount64(bb.Bishops) * 3
	score += bits.OnesCount64(bb.Rooks) * 5
	score += bits.OnesCount64(bb.Queens) * 9
	return score
}