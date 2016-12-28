package script

type Instructions struct {
	code []Instruction
	idx  int
}

func (s *Instructions) nextInstruction() Instruction {
	istr := s.peekInstruction()
	s.idx++
	return istr
}

func (s *Instructions) peekInstruction() Instruction {
	if s.idx > len(s.code) {
		panic("eof when peeking instruction")
	}

	return s.code[s.idx]
}

func (s *Instructions) reset() {
	s.idx = 0
}

func (s *Instructions) append(istr Instruction) {
	s.code = append(s.code, istr)
}

func (s *Instructions) isEOF() bool {
	return s.idx >= len(s.code)
}
