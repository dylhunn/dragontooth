package main

import (
	"fmt"
	//"math/bits"
)

// Only activate one file, A-H (A=0, H=7)
var onlyFile = [8]uint64{
	0x0101010101010101, 0x0202020202020202, 0x0404040404040404, 0x0808080808080808,
	0x1010101010101010, 0x2020202020202020, 0x4040404040404040, 0x8080808080808080}

var onlyRank = [8]uint64{
	0xFF, 0XFF00, 0XFF0000, 0XFF000000,
	0XFF00000000, 0XFF0000000000, 0XFF000000000000, 0XFF00000000000000}

func main() {
	// Generate isolated pawn tables
	resarr := make([]uint64, 64, 64)
	for i := 0; i < 64; i++ { // for every index
		var clear uint64
		for j := 7; j >= 0; j-- {
			clear |= uint64(1) << uint(j * 8 + (i % 8))
		}
		if i % 8 != 0 {
			clear |= clear >> 1
		}
		if i % 8 != 7 {
			clear |= clear << 1
		}
		fmt.Printf("%d: %x\n", i, clear)
		resarr[i] = clear
		printBitboard(clear)
	}
	for i, val := range resarr {
		if i % 4 == 0 {
			fmt.Println()
		}
		fmt.Printf("0x%x, ", val)
	}
}

func blackpassedpawns() {
	// Generate passed pawn tables
	resarr := make([]uint64, 64, 64)
	for i := 0; i < 64; i++ { // for every index
		var clear uint64
		for j := i/8 - 1; j >= 0; j-- {
			clear |= uint64(1) << uint(j * 8 + (i % 8))
		}
		if i % 8 != 0 {
			clear |= clear >> 1
		}
		if i % 8 != 7 {
			clear |= clear << 1
		}
		fmt.Printf("%d: %x\n", i, clear)
		resarr[i] = clear
		printBitboard(clear)
	}
	for i, val := range resarr {
		if i % 4 == 0 {
			fmt.Println()
		}
		fmt.Printf("0x%x, ", val)
	}
}

func whitepassedpawns() {
	// Generate passed pawn tables
	resarr := make([]uint64, 64, 64)
	for i := 0; i < 64; i++ { // for every index
		var clear uint64
		for j := i/8 + 1; j < 8; j++ {
			clear |= uint64(1) << uint(j * 8 + (i % 8))
		}
		if i % 8 != 0 {
			clear |= clear >> 1
		}
		if i % 8 != 7 {
			clear |= clear << 1
		}
		fmt.Printf("%d: %x\n", i, clear)
		resarr[i] = clear
		printBitboard(clear)
	}
	for i, val := range resarr {
		if i % 4 == 0 {
			fmt.Println()
		}
		fmt.Printf("0x%x, ", val)
	}
}

func printBitboard(bitboard uint64) {
	for i := 63; i >= 0; i-- {
		j := (i/8)*8 + (7 - (i % 8))
		if bitboard&(uint64(1)<<uint8(j)) == 0 {
			fmt.Print("-")
		} else {
			fmt.Print("X")
		}
		if i%8 == 0 {
			fmt.Println()
		}
	}
	fmt.Println()
}