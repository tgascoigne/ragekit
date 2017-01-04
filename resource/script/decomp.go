package script

import "fmt"

type Machine struct {
	file   *File
	script *Script
	instrs *Instructions

	identifierIdx int
}

func NewMachine(script *Script, code []Instruction) *Machine {
	m := &Machine{
		instrs: &Instructions{
			code: make([]InstructionState, len(code)),
			idx:  0,
		},
		script: script,
		file:   &File{},
	}

	for i := range code {
		m.instrs.code[i] = InstructionState{
			Instruction: code[i],
		}
	}

	return m
}

func (m *Machine) createStaticDecls() {
	statics := m.script.StaticValues
	for i, v := range statics {
		staticVar := Variable{
			Index:      i,
			Identifier: fmt.Sprintf("static_%v", i),
			Type:       IntType,
		}

		var staticDecl *VariableDeclaration
		if v != 0 {
			staticDecl = staticVar.DeclarationWithValue(v)
		} else {
			staticDecl = staticVar.Declaration()
		}
		staticDecl.Scope = StaticToken

		m.file.Decls.AddVariable(staticDecl)
	}
}

func (m *Machine) Decompile() File {
	m.createStaticDecls()

	m.file.Functions = make([]*Function, 0)

	for !m.instrs.isEOF() {
		istr := m.instrs.peekInstruction()

		switch istr.Operation {
		case OpEnter:
			newFunc := m.scanFunction()
			m.file.Functions = append(m.file.Functions, newFunc)
		case OpNop:

		default:
			panic(fmt.Sprintf("unexpected instruction %v\n", istr))
		}
	}

	for i := range m.file.Functions {
		m.decompileFunction(m.file.Functions[i])
		m.file.Nodes = append(m.file.Nodes, m.file.Functions[i])
	}

	return *m.file
}

func (m *Machine) scanFunction() *Function {
	function := NewFunction()

	enterIstr := m.instrs.nextInstruction()
	operands := enterIstr.Operands.(*EnterOperands)

	function.Address = enterIstr.Address
	function.Identifier = operands.Name
	function.Out = &Variable{
		Identifier: "ERROR",
		Type:       VoidType,
	}

	// Create arg decls
	for i := 0; i < int(operands.NumArgs); i++ {
		arg := &Variable{
			Index:      i,
			Identifier: fmt.Sprintf("local_%v", i),
			Type:       UnknownType,
			IsArgument: true,
		}
		function.In.AddVariable(arg.Declaration())
		function.Decls.AddVariable(arg.Declaration())
	}

	// Create local decls
	for i := int(operands.NumArgs); i < int(operands.NumLocals); i++ {
		local := &Variable{
			Index:      i,
			Identifier: fmt.Sprintf("local_%v", i),
			Type:       UnknownType,
		}
		function.Decls.AddVariable(local.Declaration())
	}

	entryBlock := newBlock(function)
	firstIstr := m.instrs.peekInstruction()
	function.blocks[firstIstr.Address] = entryBlock
	function.BasicBlock = entryBlock

	var currentBlock *BasicBlock

	// Scan for RETs and copy the instructions in this function into a new buffer
	for !m.instrs.isEOF() {
		nextIstr := m.instrs.peekInstruction()
		if nextIstr.Operation == OpEnter {
			break
		}

		if block, ok := function.blocks[nextIstr.Address]; ok {
			currentBlock = block
		}

		currentBlock.instrs.append(nextIstr)
		m.instrs.nextInstruction()

		if nextIstr.Operation > OpBranchStart && nextIstr.Operation < OpBranchEnd {
			m.scanBranch(function, nextIstr, currentBlock)
		}
	}

	return function
}

func defineFlowPath(from, to *BasicBlock) {
	from.Outs = append(from.Outs, to)
	to.Ins = append(to.Ins, from)
}

func (m *Machine) scanBranch(fn *Function, istr Instruction, currentBlock *BasicBlock) {
	// the address of the instruction immediately following the branch
	nextAddress := m.instrs.peekInstruction().Address
	fn.blocks[nextAddress] = newBlock(fn)

	if istr.Operation != OpBranch {
		// if the branch is conditional, then we need to create an entry point from istr to nextAddress
		defineFlowPath(currentBlock, fn.blocks[nextAddress])
	}

	branchAddress := istr.Operands.(*BranchOperands).AbsoluteAddr
	fn.blocks[branchAddress] = newBlock(fn)
	defineFlowPath(currentBlock, fn.blocks[branchAddress])
}

func (m *Machine) decompileFunction(fn *Function) {
	blocksStack := &link{
		Node: fn.BasicBlock,
	}

	for blocksStack != nil {
		// pop the next block
		thisBlock := blocksStack.Node.(*BasicBlock)
		blocksStack = blocksStack.next

		if _, ok := fn.blocksVisited[thisBlock.StartAddress()]; ok {
			// Dont decompile a block we've already visited
			continue
		}

		// decompile this block
		for !thisBlock.instrs.isEOF() {
			m.decompileStatement(thisBlock)
		}

		fn.blocksVisited[thisBlock.StartAddress()] = true

		// push all of the outgoing blocks of this one to the stack
		// copy the node stack at the end of decompilation into the outgoing blocks
		for _, block := range thisBlock.Outs {
			block.nodeStack = thisBlock.nodeStack
			blocksStack = &link{
				Node: block,
				next: blocksStack,
			}
		}
	}
}

func (m *Machine) decompileStatement(block *BasicBlock) {
	istr := block.peekInstruction()
	block.emitComment("\t\t\t\t\t\t\tasm(\"%v\")", istr.String())
	op := istr.Operation
	switch {
	/* standard stack ops */
	case op == OpPush:
		m.decompilePushImm(block)
	case op == OpPushStr:
		fallthrough
	case op == OpPushStrN:
		fallthrough
	case op == OpPushStrL:
		m.decompilePushStr(block)
	case op == OpDrop:
		block.popNode()
		block.nextInstruction()
	case op == OpDup:
		duped := block.peekNode()
		block.pushNode(duped)
		block.nextInstruction()

	/* variable access ops */
	case op == OpGetArrayP:
		fallthrough
	case op == OpGetArray:
		fallthrough
	case op == OpGetFieldP:
		fallthrough
	case op == OpGetField:
		fallthrough
	case op == OpGetLocalP:
		fallthrough
	case op == OpGetLocal:
		fallthrough
	case op == OpGetStaticP:
		fallthrough
	case op == OpGetStatic:
		fallthrough
	case op == OpGetGlobalP:
		fallthrough
	case op == OpGetGlobal:
		m.decompileVarAccess(block)
	case op == OpGetP:
		m.decompileGetP(block)

	/* assignment ops */
	case op == OpSetArray:
		fallthrough
	case op == OpSetField:
		fallthrough
	case op == OpSetLocal:
		fallthrough
	case op == OpSetStatic:
		fallthrough
	case op == OpSetGlobal:
		m.decompileAssignment(block)
	case op == OpSetP:
		m.decompileSetP(block, false)
	case op == OpSetPPeek:
		m.decompileSetP(block, true)

	case op == OpImplode:
		m.decompileImplode(block)
	case op == OpExplode:
		m.decompileExplode(block)

	/* control flow */
	case op == OpCall:
		m.decompileCall(block)
	case op == OpCallN:
		m.decompileCall(block)

	/* binary ops */
	case op > OpMathStart && op < OpMathEnd:
		m.decompileMathOp(block)
	/* bool ops */
	case op > OpBoolStart && op < OpBoolEnd:
		m.decompileBoolOp(block)

	case op == OpRet:
		m.decompileReturn(block)

	default:
		m.decompileUnknownOp(block)
	}
}

func (m *Machine) decompileUnknownOp(block *BasicBlock) {
	istr := block.nextInstruction()
	if istr.Operation == OpNop {
		return
	}
	block.emitStatement(AsmStmt{istr.String()})
}

func (m *Machine) decompileBranch(block *BasicBlock) {
	istr := block.nextInstruction()
	op := istr.Operands.(*BranchOperands)

	if len(block.Outs) != 2 {
		panic(fmt.Sprintf("expected two outward blocks on conditional branch, got %v", len(block.Outs)))
	}

	if op.AbsoluteAddr > istr.Address {
		cond := block.popNode()

		switch istr.Operation {
		case OpBranchZ:
			cond = NotCond{
				Node: cond,
			}
		}

		var then, els *BasicBlock
		if block.Outs[0].StartAddress() == op.AbsoluteAddr {
			then, els = block.Outs[0], block.Outs[1]
		} else {
			els, then = block.Outs[0], block.Outs[1]
		}

		block.emitStatement(IfStmt{
			Cond: cond,
			Then: then,
			Else: els,
		})
	} else {
		block.emitComment("don't know how to parse conditional branch to previous address")
		block.emitStatement(AsmStmt{istr.String()})
	}
}

func (m *Machine) decompileBoolOp(block *BasicBlock) {
	op := block.nextInstruction()
	cond := block.popNode()

	switch op.Operation {
	case OpNot:
		block.pushNode(NotCond{cond})
		return
	}

	// The remaining ops are binary
	var a, b Node
	a = block.popNode()
	if _, ok := op.Operands.(ImmediateIntOperands); ok {
		// some math ops take an immediate operand in place of a stack operand
		b = Immediate{op.Operands}
	} else {
		b = block.popNode()
	}

	var result Node
	switch op.Operation {
	case OpAnd:
		result = AndCond{
			A: a,
			B: b,
		}
	case OpOr:
		result = OrCond{
			A: a,
			B: b,
		}
	case OpXor:
		result = XorCond{
			A: a,
			B: b,
		}
	default:
		panic("unknown bool op")
	}

	block.pushNode(result)
}

func (m *Machine) decompileMathOp(block *BasicBlock) {
	var token Token
	op := block.nextInstruction()

	// Handle the neg unary operation first
	if op.Operation == OpNeg {
		a := block.popNode()

		if op.Operation == OpNeg {
			token = NegToken
		}

		unaryOp := UnaryExpr{token, a}
		block.pushNode(unaryOp)
		return
	}

	// The remaining math ops are binary
	var a, b Node
	a = block.popNode()
	if _, ok := op.Operands.(ImmediateIntOperands); ok {
		// some math ops take an immediate operand in place of a stack operand
		b = Immediate{op.Operands}
	} else {
		b = block.popNode()
	}

	switch op.Operation {
	case OpAdd:
		token = AddToken
	case OpSub:
		token = SubToken
	case OpMul:
		token = MulToken
	case OpDiv:
		token = DivToken
	case OpMod:
		token = ModToken
	case OpAnd:
		token = AndToken
	case OpOr:
		token = OrToken
	case OpXor:
		token = XorToken
	}

	binaryOp := BinaryExpr{a, token, b}
	block.pushNode(binaryOp)
}

func (m *Machine) genTempIdentifier() string {
	m.identifierIdx++
	return fmt.Sprintf("temp_%v", m.identifierIdx)
}
