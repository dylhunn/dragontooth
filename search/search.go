package search

import (
	"fmt"
	"github.com/dylhunn/dragontooth/eval"
	"github.com/dylhunn/dragontooth/transtable"
	"github.com/dylhunn/dragontoothmg"
	"math"
)

var DefaultSearchThreads int = 4

// Both constants are negatable and +1able without overflowing
const negInf = math.MinInt16 + 2
const posInf = math.MaxInt16 - 1

func lookupPv(b dragontoothmg.Board, startmove dragontoothmg.Move) string {
	var pv string = startmove.String() + " "
	b.Apply(startmove)
	for {
		found, tableMove, _, _, _ := transtable.Get(&b)
		if !found {
			break
		}
		legalMoves := b.GenerateLegalMoves()
		var isLegal bool = false
		for _, mv := range legalMoves {
			if mv.String() == tableMove.String() {
				isLegal = true
				break
			}
		}
		if isLegal {
			b.Apply(tableMove)
			pv += tableMove.String() + " "
		} else {
			fmt.Println("info string Failed table PV lookup. Table gives invalid next move",
				&tableMove, tableMove, "with PV", pv, "for position", b.ToFen())
			return pv
		}
	}
	return pv
}

func Search(board *dragontoothmg.Board, cm chan dragontoothmg.Move, halt chan bool) {
	var i int8
	stop := false
	for i = 1; ; i++ { // iterative deepening
		eval, move := ab(board, negInf, posInf, i, halt, &stop)
		if stop { // computation was truncated
			return
		} else { // valid results
			fmt.Println("info depth", i, "pv", lookupPv(*board, move), "score", eval)
			cm <- move
		}
	}
}

func ab(b *dragontoothmg.Board, alpha int16, beta int16, depth int8, halt chan bool, stop *bool) (int16, dragontoothmg.Move) {
	select {
	case <-halt:
		*stop = true
		return alpha, 0 // TODO(dylhunn): Is this a reasonable value to return?
	default:
		// continue execution
	}
	if *stop {
		return alpha, 0
	}
	found, tableMove, tableEval, tableDepth, tableNodeType := transtable.Get(b)
	if found && tableDepth >= depth {
		if tableNodeType == transtable.Exact {
			return tableEval, tableMove
		} else if tableNodeType == transtable.LowerBound {
			alpha = max(alpha, tableEval)
		} else { // upperbound
			beta = min(beta, tableEval)
		}
		if alpha >= beta {
			return tableEval, tableMove
		}
	}
	if depth == 0 {
		return eval.Evaluate(b), 0
	}

	alpha0 := alpha
	bestVal := int16(negInf)
	var bestMove dragontoothmg.Move
	moves := b.GenerateLegalMoves()
	for _, move := range moves {
		unapply := b.Apply(move)
		var score int16
		score, _ = ab(b, -beta, -alpha, depth-1, halt, stop)
		score = -score
		unapply()
		if score > bestVal {
			bestMove = move
			bestVal = score
		}
		alpha = max(alpha, score)
		if alpha >= beta {
			break
		}
	}

	select { // No writing to the transposition table after a halt instruction
	case <-halt:
		*stop = true
		return bestVal, bestMove // TODO(dylhunn): Is this a reasonable value to return?
	default:
		// continue execution
	}
	if *stop {
		return bestVal, bestMove
	}

	var nodeType uint8
	if bestVal <= alpha0 {
		nodeType = transtable.UpperBound
	} else if bestVal >= beta {
		nodeType = transtable.LowerBound
	} else {
		nodeType = transtable.Exact
	}
	transtable.Put(b, bestMove, bestVal, depth, nodeType)
	return bestVal, bestMove
}

func min(x, y int16) int16 {
	if x < y {
		return x
	}
	return y
}

func max(x, y int16) int16 {
	if x > y {
		return x
	}
	return y
}
