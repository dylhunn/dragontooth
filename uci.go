package main

import (
	"bufio"
	"fmt"
	"github.com/dylhunn/dragontooth/search"
	"github.com/dylhunn/dragontooth/transtable"
	"github.com/dylhunn/dragontoothmg"
	"os"
	"strconv"
	"strings"
)

const versionString = "0.1 'Azazel'"

func main() {
	uciLoop()
}

func uciLoop() {
	scanner := bufio.NewScanner(os.Stdin)
	board := dragontoothmg.ParseFen(dragontoothmg.Startpos) // the game board
	// used for communicating with search routine
	haltchannel := make(chan bool)
	for scanner.Scan() {
		line := scanner.Text()
		tokens := strings.Fields(line)
		if len(tokens) == 0 { // ignore blank lines
			continue
		}
		switch strings.ToLower(tokens[0]) {
		case "uci":
			fmt.Println("id name Dragontooth", versionString)
			fmt.Println("id author Dylan D. Hunn (dylhunn)")
			fmt.Println("option name Hash type spin default", transtable.DefaultTtableSize, "min 8 max 65536")
			fmt.Println("option name SearchThreads type spin default", search.DefaultSearchThreads, "min 1 max 128")
			fmt.Println("uciok")
		case "isready":
			fmt.Println("readyok")
		case "ucinewgame":
			transtable.Initialize(transtable.DefaultTtableSize)
			// reset the board, in case the GUI skips 'position' after 'newgame'
			board = dragontoothmg.ParseFen(dragontoothmg.Startpos)
		case "quit":
			return
		case "setoption":
			if len(tokens) != 5 || tokens[1] != "name" || tokens[3] != "value" {
				fmt.Println("info string Malformed setoption command")
				continue
			}
			switch strings.ToLower(tokens[2]) {
			case "hash":
				res, err := strconv.Atoi(tokens[4])
				if err != nil {
					fmt.Println("info string Hash value is not an int (", err, ")")
					continue
				}
				fmt.Println("info string Changed table size. Clearing and reloading table...")
				transtable.DefaultTtableSize = res // reset the size and reload the table
				transtable.Initialize(transtable.DefaultTtableSize)
			case "searchthreads":
				res, err := strconv.Atoi(tokens[4])
				if err != nil {
					fmt.Println("info string Number of threads is not an int (", err, ")")
					continue
				}
				search.DefaultSearchThreads = res
			default:
				fmt.Println("info string Unknown UCI option", tokens[2])
			}
		case "go":
			goScanner := bufio.NewScanner(strings.NewReader(line))
			goScanner.Split(bufio.ScanWords)
			goScanner.Scan() // skip the first token
			var wtime, btime, winc, binc int
			var infinite bool
			var err error
			for goScanner.Scan() {
				nextToken := strings.ToLower(goScanner.Text())
				switch nextToken {
				case "infinite":
					infinite = true
					continue
				case "wtime":
					if !goScanner.Scan() {
						fmt.Println("info string Malformed go command option wtime")
						continue
					}
					wtime, err = strconv.Atoi(goScanner.Text())
					if err != nil {
						fmt.Println("info string Malformed go command option; could not convert wtime")
						continue
					}
				case "btime":
					if !goScanner.Scan() {
						fmt.Println("info string Malformed go command option btime")
						continue
					}
					btime, err = strconv.Atoi(goScanner.Text())
					if err != nil {
						fmt.Println("info string Malformed go command option; could not convert btime")
						continue
					}
				case "winc":
					if !goScanner.Scan() {
						fmt.Println("info string Malformed go command option winc")
						continue
					}
					winc, err = strconv.Atoi(goScanner.Text())
					if err != nil {
						fmt.Println("info string Malformed go command option; could not convert winc")
						continue
					}
				case "binc":
					if !goScanner.Scan() {
						fmt.Println("info string Malformed go command option binc")
						continue
					}
					binc, err = strconv.Atoi(goScanner.Text())
					if err != nil {
						fmt.Println("info string Malformed go command option; could not convert binc")
						continue
					}
				default:
					fmt.Println("info string Unknown go subcommand", nextToken)
					continue
				}
			}
			_, _, _, _ = wtime, btime, winc, binc
			stop := false
			go search.Search(&board, haltchannel, &stop)
			if (wtime != 0 && btime != 0 && !infinite) { // If times are specified
				var ourtime, opptime, ourinc, oppinc int
				if board.Wtomove {
					ourtime, opptime, ourinc, oppinc = wtime, btime, winc, binc
				} else {
					ourtime, opptime, ourinc, oppinc = btime, wtime, binc, winc
				}
				allowedTime := search.CalculateAllowedTime(ourtime, opptime, ourinc, oppinc)
				go search.SearchTimeout(haltchannel, allowedTime, &stop)
			}
		case "stop":
			haltchannel <- true
		case "position":
			posScanner := bufio.NewScanner(strings.NewReader(line))
			posScanner.Split(bufio.ScanWords)
			posScanner.Scan() // skip the first token
			if !posScanner.Scan() {
				fmt.Println("info string Malformed position command")
				continue
			}
			if strings.ToLower(posScanner.Text()) == "startpos" {
				board = dragontoothmg.ParseFen(dragontoothmg.Startpos)
				posScanner.Scan() // advance the scanner to leave it in a consistent state
			} else if strings.ToLower(posScanner.Text()) == "fen" {
				fenstr := ""
				for posScanner.Scan() && strings.ToLower(posScanner.Text()) != "moves" {
					fenstr += posScanner.Text() + " "
				}
				if fenstr == "" {
					fmt.Println("info string Invalid fen position")
					continue
				}
				board = dragontoothmg.ParseFen(fenstr)
			} else {
				fmt.Println("info string Invalid position subcommand")
				continue
			}
			if strings.ToLower(posScanner.Text()) != "moves" {
				continue
			}
			for posScanner.Scan() { // for each move
				moveStr := strings.ToLower(posScanner.Text())
				legalMoves := board.GenerateLegalMoves()
				var nextMove dragontoothmg.Move
				found := false
				for _, mv := range legalMoves {
					if mv.String() == moveStr {
						nextMove = mv
						found = true
						break
					}
				}
				if !found { // we didn't find the move, but we will try to apply it anyway
					fmt.Println("info string Move", moveStr, "not found for position", board.ToFen())
					var err error
					nextMove, err = dragontoothmg.ParseMove(moveStr)
					if err != nil {
						fmt.Println("info string Contingency move parsing failed")
						continue
					}
				}
				board.Apply(nextMove)
			}
		default:
			fmt.Println("info string Unknown command:", line)
		}
	}
}
