package transtable

import (
	"github.com/dylhunn/dragontoothmg"
	"testing"
)

func TestSimpleTt(t *testing.T) {
	// Some example positions taken from apply_test
	movesMap := map[string]dragontoothmg.Move{
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 0": parseMove("e2e4"),
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQK2R w KQkq - 0 0": parseMove("e1g1"),
		"r3kbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQK2R b KQq - 0 0": parseMove("e8c8"),
		"r3k2r/1pppppp1/8/8/8/8/1PPPPPPP/R3K2R w KQq - 0 0": parseMove("a1b1"),
		"r3k2r/1pppppp1/8/8/8/8/1PPPPPPP/R3K2R b KQq - 0 0": parseMove("h8h7"),
		"r3k3/1ppp1ppr/8/3Pp3/8/8/1PP1PPPP/R3K2R w - e6 0 0": parseMove("d5e6"),
		"r3k3/1ppp1ppr/8/8/2Pp4/8/1P2PPPP/R3K2R b - c3 0 0":  parseMove("d4c3"),
		"2kr1bnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQK2R w KQ - 0 0": parseMove("a2a4"),
		"r3k3/1pp3P1/4N3/3b4/8/2p5/1P2PP1P/R3K2R w - - 0 0": parseMove("g7g8q"),
		"r3k1Q1/1pp5/4N3/3b4/8/2p5/1P2PP1p/R3K3 b - - 0 0": parseMove("h2h1n"),
		"r3k1Q1/1pp5/4N3/3br3/8/2p3n1/1p2PP2/R1B1K2n b - - 0 0": parseMove("b2c1b"),
		"r3k1Q1/1pp2p2/4Nk2/3br3/8/2p3n1/4PP2/R1b1K2n b - - 0 0": parseMove("f6e6"),
		"rnbqkbnr/ppp1pppp/8/3p4/8/8/PPP1PPPP/RNBQKBNR w KQkq - 0 2": parseMove("e1d2"),
		"rnbqkbnr/ppp1pppp/8/3p4/8/8/PPP1PPPP/RNBQKBNR w KQkq d6 0 2": parseMove("e1d2"),
		"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 0": parseMove("e1c1"),
		"r3k3/p1ppqpb1/bn2pnpr/3PN3/1p2P3/5Q1p/PPPBBPPP/RN2K2R w KQq - 0 0": parseMove("d2h6"),
		"r3k2r/p1ppqpb1/1n2pnp1/1b1PN3/1p2P3/P1N2Q1p/1PPBBPPP/R3K2R w KQkq - 0 0": parseMove("e1g1"),
		"r3k2r/Pppp1ppp/1b3nbN/nPB5/B1P1P3/q4N2/P2P2PP/r2Q1RK1 w kq - 0 0": parseMove("d1a1"),
		"r3k2r/Pppp1ppp/1b3nbN/nPB5/2P1P3/qB3N2/P2P2PP/r2Q1RK1 b kq - 0 0": parseMove("a1a2"),
	}
	Initialize(1000)
	for k, mv := range movesMap {
		b := dragontoothmg.ParseFen(k)
		Put(&b, mv, -30, -6, LowerBound)
		found, resmove, reseval, resdepth, restype := Get(&b)
		if (!found || resmove != mv || reseval != -30 || resdepth != -6 || restype != LowerBound) {
			t.Error("Simple ttable test failed. \nPut data:", b.ToFen(), &mv,
				-30, 6, Exact, "\n", "Fetched data:", found, &resmove, reseval,
				resdepth, restype)
		}
	}
}


// A testing-use function that ignores the error output
func parseMove(movestr string) dragontoothmg.Move {
	res, _ := dragontoothmg.ParseMove(movestr)
	return res
}
