package main

import "fmt"

type Search struct {
	nodes uint64
	ply      int

	killers  [2][64]Move
	history [12][64]Move

	pv_length [64]int
	pv_table  [64][64]Move
}

const (
	infinity   int = 50000
	mate_value int = 49000
	mate_score int = 48000
)

func (search *Search) quiescence(pos Position, alpha, beta int) int {
	// evaluation
	evaluation := evaluate(pos)

	// increment nodes
	search.nodes++

	// fail-hard beta cutoff
	if evaluation >= beta {
		// node (move) fails high
		return beta 
	}

	// found better move
	if evaluation > alpha {
		// PV node (move)
		alpha = evaluation
	}

	// move list
	moves := pos.generate_moves()

	// sort move list
	search.sort_moves(pos, &moves)

	for i := 0; i < moves.count; i++ {
		move := moves.list[i]

		// preserve board state
		pos.copy_board()

		// increment half move counter
		search.ply++

		// skip if move is ilegal
		if !pos.make_move(move, only_captures) {
			search.ply--
			continue
		} 

		// recursively call quiescence
		score := -search.quiescence(pos, -beta, -alpha)

		// take back move
		pos.take_back()

		// decrement ply
		search.ply--

		// fail-hard beta cutoff
		if score >= beta {
			// node (move) fails high
			return beta 
		}

		// found better move
		if score > alpha {
			// PV node (move)
			alpha = score
		}

	}

	// node fails low
	return alpha
}

func (search *Search) negamax(pos Position, alpha, beta, depth int) int {
	// initialize PV length
	search.pv_length[search.ply] = search.ply
	
	if depth == 0 {
		// search only captures
		return search.quiescence(pos, alpha, beta)
	}

	// increment nodes
	search.nodes++

	// current side to move and opposite side
	var our_side, their_side = pos.side_to_move, other_side(pos.side_to_move)

	// is king in check
	king_square := pos.bitboards[get_piece_type(King, our_side)].bsf()
	in_check := is_square_attacked(king_square, their_side, pos)

	// increase depth if king in check
	if in_check == true {
		depth++
	}

	// legal moves counter
	legal_moves := 0

	// move list
	moves := pos.generate_moves()

	// sort move list
	search.sort_moves(pos, &moves)

	for i := 0; i < moves.count; i++ {
		move := moves.list[i]

		// preserve board state
		pos.copy_board()

		// increment half move counter
		search.ply++

		
		// skip if move is ilegal
		if !pos.make_move(move, all_moves) {
			search.ply--
			continue
		} 

		// increment legal moves
		legal_moves++

		// recursively call negamax
		score := -search.negamax(pos, -beta, -alpha, depth - 1)

		// take back move
		pos.take_back()

		// decrement ply
		search.ply--

		// fail-hard beta cutoff
		if score >= beta {

			// only quiet moves
			if move.get_move_capture() == 0 {
				// store killer moves
				search.killers[1][search.ply] = search.killers[0][search.ply]
				search.killers[0][search.ply] = move
			}

			// node (move) fails high
			return beta 
		}

		// found better move
		if score > alpha {
			
			// only quiet moves
			if move.get_move_capture() == 0 {
				// store history moves
				search.history[move.get_move_piece()][move.get_move_target()] += Move(depth)
			}

			// PV node (move)
			alpha = score

			// write PV move to table
			search.pv_table[search.ply][search.ply] = move

			// copy move from deeper ply into current ply's line
			for next_ply := search.ply + 1; next_ply < search.pv_length[search.ply + 1]; next_ply++ {
				search.pv_table[search.ply][next_ply] = search.pv_table[search.ply + 1][next_ply]
			}

			// adjust PV length
			search.pv_length[search.ply] = search.pv_length[search.ply + 1]
		}

	}

	// no legal moves in current position
	if legal_moves == 0 {
		// king is in check
		if in_check == true {
			return -mate_value + search.ply
		}
		// if not, then statelmate
		return 0
	}

	// node (move) fails low
	return alpha
}

func (search *Search) position(pos Position, depth int) {
	// reset search info
	search.reset_info()

	score := search.negamax(pos, -infinity, infinity, depth)

	fmt.Print("info score cp ", score, " depth ", depth, " nodes ", search.nodes, " pv ")

	// loop over moves within PV line
	for count := 0; count < search.pv_length[0]; count++ {
		// print PV move
		print_move(search.pv_table[0][count])
		fmt.Print(" ")
	}

	fmt.Print("\n")

	fmt.Print("bestmove ")
	print_move(search.pv_table[0][0])
	fmt.Print("\n")
}

func (search *Search) reset_info() {
	// reset search info
	search.ply = 0
	search.nodes = 0

	

	for i := 0; i < 2; i++ {
		for j := 0; j < 64; j++ {
			search.killers[i][j] = Move(0)
		}
	}

	for i := 0; i < 12; i++ {
		for j := 0; j < 64; j++ {
			search.history[i][j] = Move(0)
		}
	}
	
}