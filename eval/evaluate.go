package eval

import (
	"github.com/dylhunn/dragontoothmg"
	"math/bits"
)

var DefaultDrawScore int16 = 0

// Used to estimate the remaining number of moves in the game
const MaxHalfMovesLeft = 75

const pawnValue = 100
const knightValue = 320
const bishopValue = 330
const rookValue = 500
const queenValue = 900

// These are public so that they can be changed during parameter optimization
var BishopPairBonus int = 30
var DiagonalMobilityBonus int = 4
var OrthogonalMobilityBonus int = 4
var DoubledPawnPenalty int = 20
var PassedPawnBonus int = 25
var IsolatedPawnPenalty int = 15

var pawnTableStart = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	45, 50, 50, 50, 50, 50, 50, 45,
	20, 25, 30, 35, 35, 30, 25, 20,
	10, 15, 15, 20, 20, 15, 15, 10,
	3, 5, 5, 15, 15, 5, 5, 3,
	2, 5, 3, 6, 6, 3, 5, 2,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var knightTableStart = [64]int{
	0, 1, 3, 5, 5, 3, 1, 0,
	1, 7, 10, 15, 15, 10, 7, 1,
	3, 10, 30, 40, 40, 30, 10, 3,
	5, 15, 35, 50, 50, 35, 15, 5,
	5, 15, 30, 40, 40, 30, 15, 5,
	3, 10, 15, 20, 20, 15, 10, 3,
	1, 7, 9, 11, 11, 9, 7, 1,
	0, 1, 3, 5, 5, 3, 1, 0,
}

var kingTableStart = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	2, 2, 1, 1, 1, 1, 2, 2,
	5, 5, 3, 3, 3, 3, 5, 5,
	10, 10, 7, 7, 7, 7, 10, 10,
	20, 25, 50, 20, 25, 20, 50, 20,
}

var centralizeTable = [64]int{
	0, 0, 0, 3, 3, 0, 0, 0,
	0, 20, 20, 20, 20, 20, 20, 0,
	0, 17, 30, 30, 30, 30, 17, 0,
	3, 15, 25, 50, 50, 25, 15, 3,
	3, 15, 25, 50, 50, 25, 15, 3,
	0, 17, 30, 30, 30, 30, 17, 0,
	0, 20, 20, 20, 20, 20, 20, 0,
	0, 0, 0, 3, 3, 0, 0, 0,
}

// This table's indices correspond to board indices from types.go.
// Each entry indicates the squares that must be clear, for white, for the
// white pawn at that index to be passed.
var whitePassedPawnTable = [64]uint64{
	0x303030303030300, 0x707070707070700, 0xe0e0e0e0e0e0e00, 0x1c1c1c1c1c1c1c00,
	0x3838383838383800, 0x7070707070707000, 0xe0e0e0e0e0e0e000, 0xc0c0c0c0c0c0c000,
	0x303030303030000, 0x707070707070000, 0xe0e0e0e0e0e0000, 0x1c1c1c1c1c1c0000,
	0x3838383838380000, 0x7070707070700000, 0xe0e0e0e0e0e00000, 0xc0c0c0c0c0c00000,
	0x303030303000000, 0x707070707000000, 0xe0e0e0e0e000000, 0x1c1c1c1c1c000000,
	0x3838383838000000, 0x7070707070000000, 0xe0e0e0e0e0000000, 0xc0c0c0c0c0000000,
	0x303030300000000, 0x707070700000000, 0xe0e0e0e00000000, 0x1c1c1c1c00000000,
	0x3838383800000000, 0x7070707000000000, 0xe0e0e0e000000000, 0xc0c0c0c000000000,
	0x303030000000000, 0x707070000000000, 0xe0e0e0000000000, 0x1c1c1c0000000000,
	0x3838380000000000, 0x7070700000000000, 0xe0e0e00000000000, 0xc0c0c00000000000,
	0x303000000000000, 0x707000000000000, 0xe0e000000000000, 0x1c1c000000000000,
	0x3838000000000000, 0x7070000000000000, 0xe0e0000000000000, 0xc0c0000000000000,
	0x300000000000000, 0x700000000000000, 0xe00000000000000, 0x1c00000000000000,
	0x3800000000000000, 0x7000000000000000, 0xe000000000000000, 0xc000000000000000,
	0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0,
}

// This is the equivalent table for black
var blackPassedPawnTable = [64]uint64{
	0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0,
	0x3, 0x7, 0xe, 0x1c,
	0x38, 0x70, 0xe0, 0xc0,
	0x303, 0x707, 0xe0e, 0x1c1c,
	0x3838, 0x7070, 0xe0e0, 0xc0c0,
	0x30303, 0x70707, 0xe0e0e, 0x1c1c1c,
	0x383838, 0x707070, 0xe0e0e0, 0xc0c0c0,
	0x3030303, 0x7070707, 0xe0e0e0e, 0x1c1c1c1c,
	0x38383838, 0x70707070, 0xe0e0e0e0, 0xc0c0c0c0,
	0x303030303, 0x707070707, 0xe0e0e0e0e, 0x1c1c1c1c1c,
	0x3838383838, 0x7070707070, 0xe0e0e0e0e0, 0xc0c0c0c0c0,
	0x30303030303, 0x70707070707, 0xe0e0e0e0e0e, 0x1c1c1c1c1c1c,
	0x383838383838, 0x707070707070, 0xe0e0e0e0e0e0, 0xc0c0c0c0c0c0,
	0x3030303030303, 0x7070707070707, 0xe0e0e0e0e0e0e, 0x1c1c1c1c1c1c1c,
	0x38383838383838, 0x70707070707070, 0xe0e0e0e0e0e0e0, 0xc0c0c0c0c0c0c0,
}

// For a given file, that file and the adjacent files are activated
var isolatedPawnTable = [8]uint64{
	0x303030303030303, 0x707070707070707, 0xe0e0e0e0e0e0e0e, 0x1c1c1c1c1c1c1c1c, 
	0x3838383838383838, 0x7070707070707070, 0xe0e0e0e0e0e0e0e0, 0xc0c0c0c0c0c0c0c0, 
}

// Only activate one file, A-H (A=0, H=7)
var onlyFile = [8]uint64{
	0x0101010101010101, 0x0202020202020202, 0x0404040404040404, 0x0808080808080808,
	0x1010101010101010, 0x2020202020202020, 0x4040404040404040, 0x8080808080808080}

var onlyRank = [8]uint64{
	0xFF, 0XFF00, 0XFF0000, 0XFF000000,
	0XFF00000000, 0XFF0000000000, 0XFF000000000000, 0XFF00000000000000}

// Reflect an index across a horizontal line in the center of the board.
// Used for looking up white pieces in the tables.
func reflect(idx uint8) uint8 {
	return ((7 - (idx / 8)) * 8) + (idx % 8) // todo
}

// Return a static evaluation, relative to the side to move.
func Evaluate(b *dragontoothmg.Board) int16 {
	if b.Halfmoveclock >= 100 {
		return DefaultDrawScore // It's a draw by 50 move rule.
	}
	var score int

	// Material
	score += CountMaterial(&b.White)
	score -= CountMaterial(&b.Black)

	// Piece-square tables
	score += countPieceTables(b.White.Pawns, b.Black.Pawns, &pawnTableStart)
	score += countPieceTables(b.White.Knights, b.Black.Knights, &knightTableStart)
	score += countPieceTables(b.White.Bishops, b.Black.Bishops, &centralizeTable)
	score += countKingTables(b)

	score += bishopPairBonuses(b)
	score += sliderMobilityBonuses(b)
	score += pawnDoublingPenalties(b)
	score += passedPawnBonuses(b)
	score += isolatedPawnPenalties(b)
	score += kingSafetyBonuses(b)
	score += connectedRookBonus(b)

	if !b.Wtomove {
		score = -score
	}
	return int16(score)
}

func kingSafetyBonuses(b *dragontoothmg.Board) int {
	var score int

	return score
}

func passedPawnBonuses(b *dragontoothmg.Board) int {
	var score int
	whitePawns := b.White.Pawns
	blackPawns := b.Black.Pawns
	for whitePawns != 0 {
		idx := bits.TrailingZeros64(whitePawns)
		whitePawns &= whitePawns - 1
		if whitePassedPawnTable[idx] & b.Black.Pawns == 0 {
			score += PassedPawnBonus
		}
	}
	for blackPawns != 0 {
		idx := bits.TrailingZeros64(blackPawns)
		blackPawns &= blackPawns - 1
		if blackPassedPawnTable[idx] & b.White.Pawns == 0 {
			score -= PassedPawnBonus
		}
	}
	return score
}

func isolatedPawnPenalties(b *dragontoothmg.Board) int {
	var score int
	whitePawns := b.White.Pawns
	blackPawns := b.Black.Pawns
	for whitePawns != 0 {
		idx := bits.TrailingZeros64(whitePawns)
		whitePawns &= whitePawns - 1
		file := idx % 8
		neighbors := bits.OnesCount64(isolatedPawnTable[file] & b.White.Pawns) - 1
		if neighbors == 0 {
			score -= IsolatedPawnPenalty
		}
	}
	for blackPawns != 0 {
		idx := bits.TrailingZeros64(blackPawns)
		blackPawns &= blackPawns - 1
		file := idx % 8
		neighbors := bits.OnesCount64(isolatedPawnTable[file] & b.Black.Pawns) - 1
		if neighbors == 0 {
			score += IsolatedPawnPenalty
		}
		
	}
	return score
}

func pawnDoublingPenalties(b *dragontoothmg.Board) int {
	var score int
	var wDoubledPawnCount int
	var bDoubledPawnCount int
	for i := 0; i < 8; i++ {
		currFile := onlyFile[i]
		wDoubledPawnCount += max(bits.OnesCount64(b.White.Pawns&currFile)-1, 0)
		bDoubledPawnCount += max(bits.OnesCount64(b.Black.Pawns&currFile)-1, 0)
	}
	score -= (wDoubledPawnCount * DoubledPawnPenalty)
	score += (bDoubledPawnCount * DoubledPawnPenalty)
	return score
}

func sliderMobilityBonuses(b *dragontoothmg.Board) int {
	var score int
	whiteBishops := b.White.Bishops | b.White.Queens
	blackBishops := b.Black.Bishops | b.Black.Queens
	whiteRooks := b.White.Rooks | b.White.Queens
	blackRooks := b.Black.Rooks | b.Black.Queens
	allPieces := b.White.All | b.Black.All
	for whiteBishops != 0 {
		idx := uint8(bits.TrailingZeros64(whiteBishops))
		whiteBishops &= whiteBishops - 1
		targets := dragontoothmg.CalculateBishopMoveBitboard(idx, allPieces) & (^b.White.All)
		score += (bits.OnesCount64(targets) * DiagonalMobilityBonus)
	}
	for blackBishops != 0 {
		idx := uint8(bits.TrailingZeros64(blackBishops))
		blackBishops &= blackBishops - 1
		targets := dragontoothmg.CalculateBishopMoveBitboard(idx, allPieces) & (^b.Black.All)
		score -= (bits.OnesCount64(targets) * DiagonalMobilityBonus)
	}
	for whiteRooks != 0 {
		idx := uint8(bits.TrailingZeros64(whiteRooks))
		whiteRooks &= whiteRooks - 1
		targets := dragontoothmg.CalculateRookMoveBitboard(idx, allPieces) & (^b.White.All)
		score += (bits.OnesCount64(targets) * OrthogonalMobilityBonus)
	}
	for blackRooks != 0 {
		idx := uint8(bits.TrailingZeros64(blackRooks))
		blackRooks &= blackRooks - 1
		targets := dragontoothmg.CalculateRookMoveBitboard(idx, allPieces) & (^b.Black.All)
		score -= (bits.OnesCount64(targets) * OrthogonalMobilityBonus)
	}
	return score
}

func bishopPairBonuses(b *dragontoothmg.Board) int {
	var score int
	whiteBishops := bits.OnesCount64(b.White.Bishops)
	blackBishops := bits.OnesCount64(b.White.Bishops)
	if whiteBishops >= 2 {
		score += BishopPairBonus
	}
	if blackBishops >= 2 {
		score -= BishopPairBonus
	}
	return score
}

func connectedRookBonus(b *dragontoothmg.Board) int {
	var score int

	return score
}

func countKingTables(b *dragontoothmg.Board) int {
	// Blend the king tables
	whiteKingIdx := uint8(bits.TrailingZeros64(b.White.Kings))
	blackKingIdx := uint8(bits.TrailingZeros64(b.Black.Kings))
	whiteKingStartScore := kingTableStart[reflect(whiteKingIdx)]
	whiteKingEndScore := centralizeTable[reflect(whiteKingIdx)]
	blackKingStartScore := kingTableStart[blackKingIdx]
	blackKingEndScore := centralizeTable[blackKingIdx]
	startTableWeight := CountPieces(b)
	endTableWeight := 32 - startTableWeight
	whiteKingCumScore := (startTableWeight*whiteKingStartScore +
		(endTableWeight * whiteKingEndScore)) / 32
	blackKingCumScore := (startTableWeight*blackKingStartScore +
		(endTableWeight * blackKingEndScore)) / 32
	return (whiteKingCumScore - blackKingCumScore)
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

func CountPieces(b *dragontoothmg.Board) int {
	return bits.OnesCount64(b.White.All) + bits.OnesCount64(b.Black.All)
}

// Returns an estimate of the total number of halfmoves left in the game
func EstimateHalfmovesLeft(b *dragontoothmg.Board) int {
	// This material counting formula is taken from the research of V. VUČKOVIĆ and R. ŠOLAK
	totalMaterial := CountMaterial(&b.White) + CountMaterial(&b.Black)
	var expectedHalfMovesRemaining int
	if totalMaterial < 2000 {
		expectedHalfMovesRemaining = int(float32(totalMaterial)/100 + 10)
	} else if totalMaterial <= 6000 {
		expectedHalfMovesRemaining = int(((float32(totalMaterial)/100)*3)/8 + 22)
	} else {
		expectedHalfMovesRemaining = int(float32(totalMaterial)/100*5/4 - 30)
	}
	if MaxHalfMovesLeft < expectedHalfMovesRemaining {
		return MaxHalfMovesLeft
	}
	return expectedHalfMovesRemaining
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
