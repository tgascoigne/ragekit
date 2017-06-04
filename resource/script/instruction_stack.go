package script

type InstructionState struct {
	Instruction
	nodeStack *link
}

type Instructions struct {
	code     []InstructionState
	idx      int
	idxStack *intStack
}

type intStack struct {
	val  int
	next *intStack
}

func (s *Instructions) prevInstructionState() *InstructionState {
	if s.idx <= 0 {
		return nil
	}

	return &s.code[s.idx-1]
}

func (s *Instructions) pushPos() {
	s.idxStack = &intStack{
		val:  s.idx,
		next: s.idxStack,
	}
}

func (s *Instructions) popPos() {
	s.idx = s.idxStack.val
	s.idxStack = s.idxStack.next
}

func (s *Instructions) nextInstruction() Instruction {
	istr := s.peekInstruction()
	s.idx++
	return istr
}

func (s *Instructions) peekInstruction() Instruction {
	if s.idx >= len(s.code) {
		panic("eof when peeking instruction")
	}

	return s.code[s.idx].Instruction
}

func (s *Instructions) reset() {
	s.idx = 0
}

func (s *Instructions) append(istr Instruction) {
	s.code = append(s.code, InstructionState{
		Instruction: istr,
	})
}

func (s *Instructions) isEOF() bool {
	return s.idx >= len(s.code)
}
