package script

import "fmt"

type Machine struct {
	file   *File
	script *Script
	instrs *Instructions

	identifierIdx int
}

func NewMachine(script *Script, code []Instruction) *Machine {
	return &Machine{
		instrs: &Instructions{
			code: code,
			idx:  0,
		},
		script: script,
		file:   &File{},
	}
}

func (m *Machine) createStaticDecls() {
	statics := m.script.StaticValues
	for i, v := range statics {
		staticVar := Variable{
			Identifier: fmt.Sprintf("static_%v", i),
			Type:       IntType,
		}

		var staticDecl *VariableDeclaration
		if v != 0 {
			staticDecl = staticVar.DeclarationWithValue(v)
		} else {
			staticDecl = staticVar.Declaration()
		}

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
			Identifier: fmt.Sprintf("local_%v", i),
			Type:       IntType, /* FIXME types */
		}
		function.In.AddVariable(arg.Declaration())
		function.Decls.AddVariable(arg.Declaration())
	}

	// Create local decls
	for i := int(operands.NumArgs); i < int(operands.NumLocals); i++ {
		local := &Variable{
			Identifier: fmt.Sprintf("local_%v", i),
			Type:       IntType, /* FIXME types */
		}
		function.Decls.AddVariable(local.Declaration())
	}

	// Scan for RETs and copy the instructions in this function into a new buffer
	for !m.instrs.isEOF() {
		nextIstr := m.instrs.peekInstruction()
		if nextIstr.Operation == OpEnter {
			break
		}

		function.instrs.append(nextIstr)
		m.instrs.nextInstruction()

		if nextIstr.Operation == OpRet {
			function.inferReturnType(nextIstr)
		}
	}

	return function
}

func (m *Machine) decompileFunction(fn *Function) {
	for !fn.instrs.isEOF() {
		m.decompileStatement(fn.BasicBlock)
	}
}

func (m *Machine) decompileStatement(block *BasicBlock) {
	istr := block.instrs.peekInstruction()
	block.emitComment("asm(\"%v\")", istr.String())
	op := istr.Operation
	switch {
	/* standard stack ops */
	case op == OpPush:
		fallthrough
	case op == OpPushStr:
		fallthrough
	case op == OpPushStrN:
		fallthrough
	case op == OpPushStrL:
		block.pushNode(Immediate{istr.Operands})
		block.instrs.nextInstruction()
	case op == OpDrop:
		block.popNode()
		block.instrs.nextInstruction()
	case op == OpDup:
		duped := block.peekNode()
		block.pushNode(duped)
		block.instrs.nextInstruction()

	/* variable access ops */
	case op == OpGetStaticP:
		fallthrough
	case op == OpGetStatic:
		m.decompileVarAccess(block)
	case op == OpGetLocalP:
		fallthrough
	case op == OpGetLocal:
		m.decompileVarAccess(block)

	/* assignment ops */
	case op == OpSetLocal:
		fallthrough
	case op == OpSetStatic:
		m.decompileAssignment(block)

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

	case op == OpRet:
		m.decompileReturn(block)

	default:
		m.decompileUnknownOp(block)
	}
}

func (m *Machine) decompileUnknownOp(block *BasicBlock) {
	istr := block.instrs.nextInstruction()
	if istr.Operation == OpNop {
		return
	}
	block.emitStatement(AsmStmt{istr.String()})
}

func (m *Machine) decompileCall(block *BasicBlock) {
	istr := block.instrs.nextInstruction()

	var node Node
	var targetFn *Function

	switch op := istr.Operands.(type) {
	case *CallOperands:
		targetFn = m.file.FunctionByAddress(op.Val)
	case *CallNOperands:
		targetFn = m.file.FunctionForNative(m.script.HashTable, op.Native)
	}

	args := make([]Node, targetFn.In.Size())
	for i := range args {
		thisArg := block.popNode()
		argIdx := targetFn.In.Size() - i - 1

		if inferrable, ok := thisArg.(TypeInferrable); ok {
			expectedType := targetFn.In.Vars[argIdx].DataType()
			inferrable.InferType(expectedType)
		}

		args[argIdx] = thisArg
	}

	node = FunctionCall{
		Fn:   targetFn,
		Args: args,
	}

	outSize := targetFn.Out.DataType().StackSize()
	if outSize > 0 {
		tempDecl := VariableDeclaration{
			Variable: &Variable{
				Identifier: m.genTempIdentifier(),
				Type:       targetFn.Out.DataType(),
			},
			Value: node,
		}

		resultRef := tempDecl.Variable.Reference()
		if outSize > 1 {
			outType := targetFn.Out.DataType().(ComplexType)
			exploded := outType.Explode(resultRef, outType.StackSize())
			for _, n := range exploded {
				block.pushNode(n)
			}
		} else {
			block.pushNode(resultRef)
		}

		node = tempDecl
	}

	block.emitStatement(node)
}

func (m *Machine) decompileReturn(block *BasicBlock) {
	istr := block.instrs.nextInstruction()
	op := istr.Operands.(*RetOperands)

	retVar := block.ParentFunc.Out

	if op.NumReturnVals == 0 {
		retVar.InferType(VoidType)
		block.emitStatement(ReturnStmt{nil})
	} else if op.NumReturnVals == 1 {
		retVal := block.peekNode()
		retVar.InferType(retVal.(DataTypeable).DataType())

		if v, ok := retVal.(*Variable); ok {
			retVal = v.Reference()
		}

		block.emitStatement(ReturnStmt{retVal})
	} else {
		panic("unable to infer return value of function")
	}
}

func (m *Machine) decompileVarAccess(block *BasicBlock) {
	istr := block.instrs.nextInstruction()
	op := istr.Operands.(ImmediateIntOperands)

	deRef := false
	var src Node
	switch istr.Operation {
	case OpGetStaticP:
		deRef = true
		fallthrough
	case OpGetStatic:
		src = m.file.Decls.VariableByName(fmt.Sprintf("static_%v", op.Int()))

	case OpGetLocalP:
		deRef = true
		fallthrough
	case OpGetLocal:
		src = block.VariableByName(fmt.Sprintf("local_%v", op.Int()))

	default:
		fmt.Printf("dont know how to find var\n")
	}

	if deRef {
		src = DeRefExpr{
			Node: src,
		}
	}

	block.pushNode(src)
}

func (m *Machine) decompileImplode(block *BasicBlock) {
	_ = block.instrs.nextInstruction()
	dest := block.popNode()

	length := block.popNode().(Immediate).Value.(ImmediateIntOperands).Int()
	elems := make(ArrayLiteral, length)
	for i := range elems {
		elems[length-i-1] = block.popNode()
	}

	expectedtype := elems[0].(DataTypeable).DataType()
	dest.(TypeInferrable).InferType(expectedtype)

	block.emitStatement(AssignStmt{dest, elems})
}

func (m *Machine) decompileExplode(block *BasicBlock) {
	_ = block.instrs.nextInstruction()
	src := block.popNode()
	length := block.popNode().(Immediate).Value.(ImmediateIntOperands).Int()

	// Is it a vec3 that we dont know about?
	if inferrable, ok := src.(TypeInferrable); ok && length == 3 {
		inferrable.InferType(GetType("Vector3"))
	}

	// eww
	srcType := src.(DataTypeable).DataType().(ComplexType)

	exploded := srcType.Explode(src, length)
	for _, n := range exploded {
		block.pushNode(n)
	}

}

func (m *Machine) doExplode(block *BasicBlock, src Node, length int) {
	for i := 0; i < length; i++ {
		block.pushNode(ArrayIndex{
			Array: src,
			Index: IntImmediate(uint32(i)),
		})
	}
}

func (m *Machine) decompileAssignment(block *BasicBlock) {
	istr := block.instrs.nextInstruction()
	op := istr.Operands.(ImmediateIntOperands)

	var dest *Variable
	switch istr.Operation {
	case OpSetStatic:
		dest = m.file.Decls.VariableByName(fmt.Sprintf("static_%v", op.Int()))
	case OpSetLocal:
		dest = block.VariableByName(fmt.Sprintf("local_%v", op.Int()))
	default:
		fmt.Printf("dont know how to find var\n")
	}

	value := block.popNode()

	expectedtype := value.(DataTypeable).DataType()
	dest.InferType(expectedtype)

	block.emitStatement(AssignStmt{dest, value})
}

func (m *Machine) decompileMathOp(block *BasicBlock) {
	var token Token
	op := block.instrs.nextInstruction()

	// Handle the two unary operations first
	if op.Operation == OpNot || op.Operation == OpNeg {
		a := block.popNode()

		if op.Operation == OpNot {
			token = NotToken
		}

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
