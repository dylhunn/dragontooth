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

var pawnTableStart = [64]int{
	 0, 0, 0, 0, 0, 0, 0, 0,
	45,50,50,50,50,50,50,45,
	20,25,30,35,35,30,25,20,
	10,15,15,20,20,15,15,10,
	 3, 5, 5,15,15, 5, 5, 3,
	 2, 5, 3, 6, 6, 3, 5, 2,
	 0, 0, 0, 0, 0, 0, 0, 0,
	 0, 0, 0, 0, 0, 0, 0, 0,
}

var knightTableStart = [64]int{
	 0, 1, 3, 5, 5, 3, 1, 0,
	 1, 7,10,15,15,10, 7, 1,
	 3,10,30,40,40,30,10, 3,
	 5,15,35,50,50,35,15, 5,
	 5,15,30,40,40,30,15, 5,
	 3,10,15,20,20,15,10, 3,
	 1, 7, 9,11,11, 9, 7, 1,
	 0, 1, 3, 5, 5, 3, 1, 0,
}

var kingTableStart = [64]int{
	 0, 0, 0, 0, 0, 0, 0, 0,
	 0, 0, 0, 0, 0, 0, 0, 0,
	 0, 0, 0, 0, 0, 0, 0, 0,
	 0, 0, 0, 0, 0, 0, 0, 0,
	 2, 2, 1, 1, 1, 1, 2, 2,
	 5, 5, 3, 3, 3, 3, 5, 5,
	10,10, 7, 7, 7, 7,10,10,
	20,25,50,20,25,25,50,20,
}

var centralizeTable = [64]int{
	 0, 0, 0, 3, 3, 0, 0, 0,
	 0,20,20,20,20,20,20, 0,
	 0,17,30,30,30,30,17, 0,
	 3,15,25,50,50,25,15, 3,
	 3,15,25,50,50,25,15, 3,
	 0,17,30,30,30,30,17, 0,
	 0,20,20,20,20,20,20, 0,
	 0, 0, 0, 3, 3, 0, 0, 0,
}

// Reflect an index across a horizontal line in the center of the board.
// Used for looking up white pieces in the tables.
func reflect(idx uint8) uint8 {
	return ((7 - (idx / 8)) * 8) + (idx % 8) // todo
}

// Return a static evaluation, relative to the side to move.
func Evaluate(b *dragontoothmg.Board) int16 {
	var score int
	score += CountMaterial(&b.White)
	score -= CountMaterial(&b.Black)
	score += countPieceTables(b.White.Pawns, b.Black.Pawns, &pawnTableStart)
	score += countPieceTables(b.White.Knights, b.Black.Knights, &knightTableStart)
	score += countPieceTables(b.White.Bishops, b.Black.Bishops, &centralizeTable)

	if !b.Wtomove {
		score = -score
	}
	return int16(score)
}

func countPieceTables(bbw uint64, bbb uint64, table *[64]int) int {
	var score int
	for bbw != 0 {
		var idx uint8 = uint8(bits.TrailingZeros64(bbw))
		bbw &= bbw - 1
		score += pawnTableStart[reflect(idx)]
	}
	for bbb != 0 {
		var idx uint8 = uint8(bits.TrailingZeros64(bbb))
		bbb &= bbb - 1
		score -= pawnTableStart[idx]
	}
	return score
}

// This is public so it can also be used for time managment
func CountMaterial(bb *dragontoothmg.Bitboards) int {
	var score int
	score += bits.OnesCount64(bb.Pawns) * pawnValue
	score += bits.OnesCount64(bb.Knights) * knightValue
	score += bits.OnesCount64(bb.Bishops) * bishopValue
	score += bits.OnesCount64(bb.Rooks) * rookValue
	score += bits.OnesCount64(bb.Queens) * queenValue
	return score
}
