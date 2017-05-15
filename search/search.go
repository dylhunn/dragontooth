package search

import (
	"github.com/dylhunn/dragontoothmg"
	"github.com/dylhunn/dragontooth/eval"
)

var DefaultSearchThreads int = 4

func Search() {

}

func abMax(b *dragontoothmg.Board, alpha int, beta int, depth int) int {
	if (depth == 0) {
		return eval.Evaluate(b)
	}
	moves := b.GenerateLegalMoves()
	for _, move := range moves {
		unapply := b.Apply(move)
		score := abMin(b, alpha, beta, depth - 1)
		unapply()
		if score >= beta {
			return beta
		}
		if score > alpha {
			alpha = score
		}
	}
	return alpha
}

func abMin(b *dragontoothmg.Board, alpha int, beta int, depth int) int {
	if (depth == 0) {
		return eval.Evaluate(b)
	}
	moves := b.GenerateLegalMoves()
	for _, move := range moves {
		unapply := b.Apply(move)
		score := abMax(b, alpha, beta, depth - 1)
		unapply()
		if score <= alpha {
			return alpha
		}
		if score < beta {
			beta = score
		}
	}
	return beta
}