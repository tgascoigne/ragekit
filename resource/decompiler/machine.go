package decompiler

import (
	"fmt"

	"github.com/tgascoigne/ragekit/resource/script"
)

type Machine struct {
	code   []script.Instruction
	idx    int
	file   *File
	script *script.Script

	identifierIdx int
}

func NewMachine(script *script.Script, code []script.Instruction) *Machine {
	return &Machine{
		code:   code,
		script: script,
		file:   &File{},
	}
}

func (m *Machine) popInstruction() script.Instruction {
	istr := m.peekInstruction()
	m.idx++
	return istr
}

func (m *Machine) peekInstruction() script.Instruction {
	if m.idx > len(m.code) {
		panic("eof when peeking instruction")
	}

	return m.code[m.idx]
}

func (m *Machine) isEOF() bool {
	return m.idx >= len(m.code)
}

func (m *Machine) createStaticDecls() {
	statics := m.script.StaticValues
	for i, v := range statics {
		staticVar := Variable{
			Identifier: fmt.Sprintf("static_%v", i),
			Type:       script.IntType,
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

	for !m.isEOF() {
		istr := m.peekInstruction()
		var node Node

		switch istr.Operation {
		case script.OpEnter:
			node = m.decompileFunction()
		case script.OpNop:

		default:
			panic(fmt.Sprintf("unexpected instruction %v\n", istr))
		}

		m.file.Nodes = append(m.file.Nodes, node)
	}

	return *m.file
}

func (m *Machine) decompileFunction() Function {
	function := &Function{}

	enterIstr := m.popInstruction()
	operands := enterIstr.Operands.(*script.EnterOperands)

	function.Identifier = operands.Name
	function.Out = &Variable{
		Identifier: "ERROR",
		Type:       script.VoidType,
	}

	for i := 0; i < int(operands.NumArgs); i++ {
		arg := &Variable{
			Identifier: fmt.Sprintf("local_%v", i),
			Type:       script.IntType, /* FIXME types */
		}
		function.In.AddVariable(arg.Declaration())
		function.Decls.AddVariable(arg.Declaration())
	}

	for i := int(operands.NumArgs); i < int(operands.NumLocals); i++ {
		local := &Variable{
			Identifier: fmt.Sprintf("local_%v", i),
			Type:       script.IntType, /* FIXME types */
		}
		function.Decls.AddVariable(local.Declaration())
	}

	/* parse body */
	for !m.isEOF() {
		/* FIXME for now, we can only tell a function has ended because another has begun
		   need control flow handling */
		istr := m.peekInstruction()
		if istr.Operation == script.OpEnter {
			break
		}

		m.decompileStatement(function)
	}

	return *function
}

func (fn *Function) emitStatement(stmt Node) {
	fn.Statements = append(fn.Statements, stmt)
}

func (fn *Function) emitComment(format string, args ...interface{}) {
	commentStr := fmt.Sprintf(format, args...)
	comment := Comment(commentStr)
	fn.emitStatement(comment)
	fmt.Println(commentStr)
}

func (fn *Function) pushNode(node Node) {
	fn.nodeStack = append(fn.nodeStack, node)
	fn.nodeStackIdx++
	fn.emitComment("pushing %v at stack idx %v", node.CString(), fn.nodeStackIdx)
}

func (fn *Function) popNode() Node {
	if fn.nodeStackIdx <= 0 {
		fmt.Println("node stack underflow")
		return Immediate{&script.Immediate32Operands{Val: 0xBABE}}
	}

	node := fn.peekNode()
	fn.nodeStackIdx--
	fn.nodeStack = fn.nodeStack[:fn.nodeStackIdx]
	//	fmt.Printf("popping %v %v\n", node.CString(), fn.nodeStackIdx)
	return node
}

func (fn *Function) peekNode() Node {
	if fn.nodeStackIdx <= 0 {
		//		fmt.Printf("peek %v\n", fn.nodeStackIdx)
		fmt.Println("node stack underflow")
		return Immediate{&script.Immediate32Operands{Val: 0xBABE}}
	}

	return fn.nodeStack[fn.nodeStackIdx-1]
}

func (m *Machine) decompileStatement(fn *Function) {
	istr := m.peekInstruction()
	//fn.emitComment("asm(\"%v\")", istr.String())
	op := istr.Operation
	switch {
	/* standard stack ops */
	case op == script.OpPush:
		fallthrough
	case op == script.OpPushStr:
		fallthrough
	case op == script.OpPushStrN:
		fallthrough
	case op == script.OpPushStrL:
		fn.pushNode(Immediate{istr.Operands})
		m.popInstruction()
	case op == script.OpDrop:
		fn.popNode()
		m.popInstruction()
	case op == script.OpDup:
		duped := fn.peekNode()
		fn.pushNode(duped)
		m.popInstruction()

	/* variable access ops */
	case op == script.OpGetStaticP:
		fallthrough
	case op == script.OpGetStatic:
		m.decompileVarAccess(fn)
	case op == script.OpGetLocalP:
		fallthrough
	case op == script.OpGetLocal:
		m.decompileVarAccess(fn)

	/* assignment ops */
	case op == script.OpSetLocal:
		fallthrough
	case op == script.OpSetStatic:
		m.decompileAssignment(fn)

	case op == script.OpImplode:
		m.decompileImplode(fn)

	/* binary ops */
	case op > script.OpMathStart && op < script.OpMathEnd:
		m.decompileMathOp(fn)

	case op == script.OpRet:
		m.decompileReturn(fn)

	default:
		m.decompileUnknownOp(fn)
	}
}

func (m *Machine) decompileUnknownOp(fn *Function) {
	istr := m.popInstruction()
	if istr.Operation == script.OpNop {
		return
	}
	fn.emitStatement(AsmStmt{istr.String()})
}

func (m *Machine) decompileReturn(fn *Function) {
	istr := m.popInstruction()
	op := istr.Operands.(*script.RetOperands)

	retVar := fn.Out

	if op.NumReturnVals == 0 {
		retVar.InferType(script.VoidType)
		fn.emitStatement(ReturnStmt{nil})
	} else if op.NumReturnVals == 1 {
		retVal := fn.peekNode()
		retVar.InferType(retVal.(script.DataTypeable).DataType())

		if v, ok := retVal.(*Variable); ok {
			retVal = v.Reference()
		}

		fn.emitStatement(ReturnStmt{retVal})
	} else {
		panic("unable to infer return value of function")
	}
}

func (m *Machine) decompileVarAccess(fn *Function) {
	istr := m.popInstruction()
	op := istr.Operands.(script.ImmediateIntOperands)

	deRef := false
	var src Node
	switch istr.Operation {
	case script.OpGetStaticP:
		deRef = true
		fallthrough
	case script.OpGetStatic:
		src = m.file.Decls.VariableByName(fmt.Sprintf("static_%v", op.Int()))

	case script.OpGetLocalP:
		deRef = true
		fallthrough
	case script.OpGetLocal:
		src = fn.Decls.VariableByName(fmt.Sprintf("local_%v", op.Int()))

	default:
		fmt.Printf("dont know how to find var\n")
	}

	if deRef {
		src = DeRefExpr{
			Node: src,
		}
	}

	fn.pushNode(src)
}

func (m *Machine) decompileImplode(fn *Function) {
	_ = m.popInstruction()
	dest := fn.popNode()

	length := fn.popNode().(Immediate).Value.(script.ImmediateIntOperands).Int() // ewww
	elems := make(ArrayLiteral, length)
	for i := range elems {
		elems[length-i-1] = fn.popNode()
	}

	// FIXME need array types
	expectedtype := elems[0].(script.DataTypeable).DataType()
	dest.(TypeInferrable).InferType(expectedtype)

	fn.emitStatement(AssignStmt{dest, elems})
}

func (m *Machine) decompileAssignment(fn *Function) {
	istr := m.popInstruction()
	op := istr.Operands.(script.ImmediateIntOperands)

	var dest *Variable
	switch istr.Operation {
	case script.OpSetStatic:
		dest = m.file.Decls.VariableByName(fmt.Sprintf("static_%v", op.Int()))
	case script.OpSetLocal:
		dest = fn.Decls.VariableByName(fmt.Sprintf("local_%v", op.Int()))
	default:
		fmt.Printf("dont know how to find var\n")
	}

	value := fn.popNode()

	expectedtype := value.(script.DataTypeable).DataType()
	dest.InferType(expectedtype)

	fn.emitStatement(AssignStmt{dest, value})
}

func (m *Machine) decompileMathOp(fn *Function) {
	var token Token
	op := m.popInstruction()

	fmt.Printf("math op is %v\n", op.String())

	// Handle the two unary operations first
	if op.Operation == script.OpNot || op.Operation == script.OpNeg {
		a := fn.popNode()

		if op.Operation == script.OpNot {
			token = NotToken
		}

		if op.Operation == script.OpNeg {
			token = NegToken
		}

		unaryOp := UnaryExpr{token, a}
		fn.pushNode(unaryOp)
		return
	}

	// The remaining math ops are binary
	var a, b Node
	a = fn.popNode()
	if _, ok := op.Operands.(script.ImmediateIntOperands); ok {
		// some math ops take an immediate operand in place of a stack operand
		b = Immediate{op.Operands}
		fmt.Println("using immediate instead")
	} else {
		b = fn.popNode()
	}

	switch op.Operation {
	case script.OpAdd:
		token = AddToken
	case script.OpSub:
		token = SubToken
	case script.OpMul:
		token = MulToken
	case script.OpDiv:
		token = DivToken
	case script.OpMod:
		token = ModToken
	case script.OpAnd:
		token = AndToken
	case script.OpOr:
		token = OrToken
	case script.OpXor:
		token = XorToken
	}

	binaryOp := BinaryExpr{a, token, b}
	fn.pushNode(binaryOp)
}

func (m *Machine) genIdentifier(prefix string) string {
	m.identifierIdx++
	return fmt.Sprintf("%v_%v", prefix, m.identifierIdx)
}

func (m *Machine) genLocalIdentifier() string {
	return m.genIdentifier("local")
}
