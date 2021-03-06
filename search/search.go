package search

import (
	"fmt"
	"github.com/dylhunn/dragontooth/eval"
	"github.com/dylhunn/dragontooth/transtable"
	"github.com/dylhunn/dragontoothmg"
	"math"
	"math/rand"
	"runtime"
	"time"
)

var DefaultSearchThreads int = runtime.NumCPU()

// Used to keep track of positions that have occurred in the game
var HistoryMap map[uint64]int = make(map[uint64]int)

// Both constants are negatable and +1able without overflowing
const NegInf = math.MinInt16 + 2
const PosInf = math.MaxInt16 - 1

const QuiesceCutoffDepth = 5

// Using the transposition table, attempt to reconstruct the rest of the PV
// (after the first move).
func lookupPv(b dragontoothmg.Board, startmove dragontoothmg.Move, depth int) string {
	var pv string = startmove.String()
	if startmove == 0 {
		fmt.Println("info string looking up PV of null move!")
		return "0000"
	}
	b.Apply(startmove)
	for i := depth; i >= 0; i-- {
		found, tableMove, tableEval, depth, _ := transtable.Get(&b)
		if !found || depth == 0 {
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
			pv += " " + tableMove.String()
		} else if tableMove == 0 && (tableEval <= NegInf || tableEval >= PosInf) {
			// This is a mate
			return pv
		} else {
			fmt.Println("info string Failed table PV lookup. Table gives invalid next move",
				&tableMove, tableMove, "with PV", pv, "for position", b.ToFen())
			return pv
		}
	}
	return pv
}

func CalculateAllowedTime(b *dragontoothmg.Board, ourtime int, opptime int, ourinc int, oppinc int) int {
	result := ourtime / eval.EstimateHalfmovesLeft(b)
	if result <= 0 {
		return 100
	}
	return result
}

// After a certain period of time, sends a signal to halt the search,
// if it has not already been sent.
// Also prints the best move. If the sleep time is 0, does nothing.
// The bool pointer alreadyStopped should be the same as the one given to Search().
func SearchTimeout(halt chan<- bool, ms int, alreadyStopped *bool) {
	if ms == 0 {
		return
	}
	time.Sleep(time.Duration(ms) * time.Millisecond)
	if !(*alreadyStopped) { // don't send the halt signal if the search has already been stopped
		halt <- true
	}
}

var nodeCount int = 0 // Used for search statistics

// The main entrypoint for the search. Spawns the appropriate number of threads,
// and prints the results (including pv and bestmove).
func Search(board *dragontoothmg.Board, halt <-chan bool, stop *bool) {
	var i int8
	var lastMove dragontoothmg.Move = 0
	for i = 1; i < math.MaxInt8; i++ { // iterative deepening
		threadsToSpawn := DefaultSearchThreads
		moves := make([]dragontoothmg.Move, threadsToSpawn)
		evals := make([]int16, threadsToSpawn)
		movesChan := make(chan dragontoothmg.Move)
		evalsChan := make(chan int16)
		start := time.Now()
		nodeCount = 0
		// Start the search threads
		for thread := 0; thread < threadsToSpawn; thread++ {
			boardCopy := *board
			go abWrapper(&boardCopy, NegInf, PosInf, i, halt, stop, movesChan, evalsChan)
		}
		// Block until the search stops, then collect the results
		for thread := 0; thread < threadsToSpawn; thread++ {
			moves[thread], evals[thread] = <-movesChan, <-evalsChan
		}
		// Sanity check: results should be the same
		/*for thread := 0; thread < threadsToSpawn-1; thread++ {
			if moves[thread].String() != moves[thread+1].String() {
				fmt.Println("info string Search threads returned inconsistent results:",
					moves[thread].String(), moves[thread+1].String())
			}
			if evals[thread] != evals[thread+1] {
				fmt.Println("info string Search threads returned inconsistent evals:",
					evals[thread], evals[thread+1])
			}
		}*/
		timeElapsed := time.Since(start)
		eval, move := evals[0], moves[0]
		if lastMove == 0 {
			lastMove = move
		}
		if *stop { // computation was truncated
			fmt.Println("bestmove", &lastMove)
			return
		} else { // valid results
			fmt.Println("info depth", i, "score cp", eval, "pv", lookupPv(*board, move, int(i)), "time",
				timeElapsed.Nanoseconds()/1000000, "hashfull", int(transtable.Load()*1000), "nodes",
				nodeCount, "nps", int(float64(nodeCount)/(timeElapsed.Seconds())))
			lastMove = move
			if eval <= NegInf || eval >= PosInf { // we found a mate; wait for the stop
				mateInPly := i
				if eval <= NegInf { // negate if we are mated
					mateInPly = -mateInPly
				}
				fmt.Println("info score mate", mateInPly/2)
				*stop = <-halt
				fmt.Println("bestmove", &lastMove)
				return
			}
		}
	}
	if i == math.MaxInt8 { // We reached max depth; wait for the stop signal
		*stop = <-halt
		fmt.Println("bestmove", &lastMove)
	}
}

// Use a collection of heuristics to sort the moves in their best order.
func sortMoves(b *dragontoothmg.Board, alpha int16, beta int16, depth int8,
	halt <-chan bool, stop *bool, moves *[]dragontoothmg.Move, history *map[uint64]int) {
	found, tableMove, _, _, _ := transtable.Get(b)
	if !found || tableMove == 0 { // use IID to guess the best move
		var resMove dragontoothmg.Move
		for i := int8(0); i < depth-1; i++ {
			_, resMove = ab(b, alpha, beta, i, halt, stop, history)
		}
		found, tableMove = true, resMove
	}
	if found && tableMove != 0 {
		for i := 0; i < len(*moves); i++ {
			if (*moves)[i].String() == tableMove.String() {
				(*moves)[0], (*moves)[i] = (*moves)[i], (*moves)[0]
				break
			}
		}
	}
}

// Wraps the ab-search function at full-depth, so the return values can be sent over
// the channels for goroutine invocations.
func abWrapper(b *dragontoothmg.Board, alpha int16, beta int16, depth int8, halt <-chan bool,
	stop *bool, moveChan chan<- dragontoothmg.Move, evalChan chan<- int16) {
	localHistory := make(map[uint64]int)
	for k2, v2 := range HistoryMap { // copy the game history map so we can modify it
		localHistory[k2] = v2
	}
	eval, move := ab(b, alpha, beta, depth, halt, stop, &localHistory)
	moveChan <- move
	evalChan <- eval
}

// Perform the alpha-beta search.
// The history parameter is to keep track of positions that have occurred, to identify draw-by-repetition
func ab(b *dragontoothmg.Board, alpha int16, beta int16, depth int8, halt <-chan bool,
	stop *bool, history *map[uint64]int) (int16, dragontoothmg.Move) {
	nodeCount++

	// check for draw by 3-fold repetition
	if (*history)[b.Hash()] >= 3 { // It's a draw
		return eval.DefaultDrawScore, 0
	} else if (*history)[b.Hash()] == 2 { // We are in danger of causing a draw
		transtable.Erase(b) // TODO(dylhunn): Is there a better way to prevent surprise draws?
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
		if alpha >= beta && tableMove != 0 {
			return tableEval, tableMove
		}
	}
	if depth == 0 {
		//return eval.Evaluate(b), 0
		return quiesce(b, alpha, beta, depth, stop), 0
	}

	alpha0 := alpha
	bestVal := int16(NegInf)
	moves := b.GenerateLegalMoves()
	if len(moves) == 0 || b.Halfmoveclock >= 100 {
		if b.OurKingInCheck() { // checkmate
			return NegInf, 0
		} else {
			return eval.DefaultDrawScore, 0 // stalemate
		}
	}
	sortMoves(b, alpha, beta, depth, halt, stop, &moves, history)
	var bestMove dragontoothmg.Move
	if len(moves) > 0 {
		bestMove = moves[0] // randomly pick some move
	} else {
		bestMove = 0
	}

	select {
	case <-halt:
		*stop = true
		return eval.Evaluate(b), bestMove // TODO(dylhunn): Is this a reasonable value to return?
	default: // continue execution
	}
	if *stop {
		return eval.Evaluate(b), bestMove
	}

	for _, move := range moves {
		unapply := b.Apply(move)
		(*history)[b.Hash()]++
		var score int16
		score, _ = ab(b, -beta, -alpha, depth-1, halt, stop, history)
		score = -score
		(*history)[b.Hash()]--
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

// Sort capture moves using MVV-LVA. Remove any non-capture moves.
func sortCaptureMoves(b *dragontoothmg.Board, moves *[]dragontoothmg.Move) {
	// TODO
}

// Quiescence search explores moves until a quiet position is reached, but cutting
// of at a certain depth.
func quiesce(b *dragontoothmg.Board, alpha int16, beta int16, depth int8, stop *bool) int16 {
	nodeCount++
	if *stop {
		return alpha
	}
	var standPat int16
	found, _, evalresult, _, ntype := transtable.Get(b)
	if found && ntype == transtable.Exact {
		standPat = evalresult
	} else {
		standPat = eval.Evaluate(b)
		transtable.Put(b, 0, standPat, 0, transtable.Exact)
	}
	if depth < -QuiesceCutoffDepth {
		return standPat
	}
	if standPat >= beta {
		return beta
	}
	if alpha < standPat {
		alpha = standPat
	}
	moves := b.GenerateLegalMoves()
	if len(moves) == 0 {
		if b.OurKingInCheck() {
			return NegInf
		} else {
			return eval.DefaultDrawScore // stalemate
		}
	}
	sortCaptureMoves(b, &moves)
	for _, move := range moves {
		if !dragontoothmg.IsCapture(move, b) {
			continue
		}
		unapply := b.Apply(move)
		score := -quiesce(b, -beta, -alpha, depth-1, stop)
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
