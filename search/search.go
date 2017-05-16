package search

import (
	"fmt"
	"github.com/dylhunn/dragontooth/eval"
	"github.com/dylhunn/dragontooth/transtable"
	"github.com/dylhunn/dragontoothmg"
	"math"
	"math/rand"
	"time"
)

var DefaultSearchThreads int = 4

// Both constants are negatable and +1able without overflowing
const negInf = math.MinInt16 + 2
const posInf = math.MaxInt16 - 1

// Using the transposition table, attempt to reconstruct the rest of the PV
// (after the first move).
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

func CalculateAllowedTime(ourtime int, opptime int, ourinc int, oppinc int) int {
	return 5000
}

// After a certain period of time, sends a signal to halt the search,
// if it has not already been sent.
// Also prints the best move. If the sleep time is 0, does nothing.
// The bool pointer alreadyStopped should be the same as the one given to Search().
func SearchTimeout(halt chan bool, ms int, alreadyStopped *bool) {
	if (ms == 0) {
		return
	}
	time.Sleep(time.Duration(ms) * time.Millisecond)
	if !*alreadyStopped { // don't send the halt signal if the search has already been stopped
		halt <- true
	}
}

// The main entrypoint for the search. Spawns the appropriate number of threads,
// and prints the results (including pv and bestmove).
func Search(board *dragontoothmg.Board, halt chan bool, stop *bool) {
	var i int8
	var lastMove dragontoothmg.Move
	for i = 1; ; i++ { // iterative deepening
		threadsToSpawn := DefaultSearchThreads
		moves := make([]dragontoothmg.Move, threadsToSpawn)
		evals := make([]int16, threadsToSpawn)
		movesChan := make(chan dragontoothmg.Move)
		evalsChan := make(chan int16)
		// Start the search threads
		for thread := 0; thread < threadsToSpawn; thread++ {
			boardCopy := *board
			go abWrapper(&boardCopy, negInf, posInf, i, halt, stop, movesChan, evalsChan)
		}
		// Block until the search stops, then collect the results
		for thread := 0; thread < threadsToSpawn; thread++ {
			moves[thread], evals[thread] = <-movesChan, <-evalsChan
		}
		// Sanity check: results should be the same
		for thread := 0; thread < threadsToSpawn - 1; thread++ {
			if moves[thread] != moves[thread+1] || evals[thread] != evals[thread+1] {
				fmt.Println("info string Search threads returned inconsistent results.")
			}
		}
		eval, move := evals[0], moves[0]
		if *stop { // computation was truncated
			fmt.Println("bestmove", &lastMove)
			return
		} else { // valid results
			fmt.Println("info depth", i, "pv", lookupPv(*board, move), "score", eval)
			lastMove = move
		}
	}
}

// Wraps the ab-search function at full-depth, so the return values can be sent over
// the channels.
func abWrapper(b *dragontoothmg.Board, alpha int16, beta int16, depth int8, halt chan bool, 
	stop *bool, moveChan chan dragontoothmg.Move, evalChan chan int16) {
	eval, move := ab(b, alpha, beta, depth, halt, stop)
	moveChan <- move
	evalChan <- eval
}

// Perform the alpha-beta search.
func ab(b *dragontoothmg.Board, alpha int16, beta int16, depth int8, halt chan bool, stop *bool) (int16, dragontoothmg.Move) {
	select {
	case <-halt:
		*stop = true
		return alpha, 0 // TODO(dylhunn): Is this a reasonable value to return?
	default: // continue execution
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

func shuffle(list []dragontoothmg.Move) {
	for i := len(list) - 1; i >= 0; i-- {
		j := rand.Intn(i + 1)
		list[j], list[i] = list[i], list[j]
	}
}