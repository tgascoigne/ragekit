package decompiler

import (
	"fmt"

	"github.com/tgascoigne/ragekit/resource/script"
)

type Machine struct {
	file   *File
	script *script.Script
	instrs *Instructions

	identifierIdx int
}

func NewMachine(script *script.Script, code []script.Instruction) *Machine {
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

	m.file.Functions = make([]*Function, 0)

	for !m.instrs.isEOF() {
		istr := m.instrs.peekInstruction()

		switch istr.Operation {
		case script.OpEnter:
			newFunc := m.scanFunction()
			m.file.Functions = append(m.file.Functions, &newFunc)
		case script.OpNop:

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

func (m *Machine) scanFunction() Function {
	function := Function{
		instrs: &Instructions{},
	}

	enterIstr := m.instrs.nextInstruction()
	operands := enterIstr.Operands.(*script.EnterOperands)

	function.Address = enterIstr.Address
	function.Identifier = operands.Name
	function.Out = &Variable{
		Identifier: "ERROR",
		Type:       script.VoidType,
	}

	// Create arg decls
	for i := 0; i < int(operands.NumArgs); i++ {
		arg := &Variable{
			Identifier: fmt.Sprintf("local_%v", i),
			Type:       script.IntType, /* FIXME types */
		}
		function.In.AddVariable(arg.Declaration())
		function.Decls.AddVariable(arg.Declaration())
	}

	// Create local decls
	for i := int(operands.NumArgs); i < int(operands.NumLocals); i++ {
		local := &Variable{
			Identifier: fmt.Sprintf("local_%v", i),
			Type:       script.IntType, /* FIXME types */
		}
		function.Decls.AddVariable(local.Declaration())
	}

	// Scan for RETs and copy the instructions in this function into a new buffer
	for !m.instrs.isEOF() {
		nextIstr := m.instrs.peekInstruction()
		if nextIstr.Operation == script.OpEnter {
			break
		}

		function.instrs.append(nextIstr)
		m.instrs.nextInstruction()

		if nextIstr.Operation == script.OpRet {
			function.inferReturnType(nextIstr)
		}
	}

	return function
}

func (fn *Function) inferReturnType(ret script.Instruction) {
	retVar := fn.Out
	op := ret.Operands.(*script.RetOperands)
	if op.NumReturnVals == 0 {
		retVar.InferType(script.VoidType)
	} else if op.NumReturnVals == 1 {
		retVal := fn.peekNode()
		retVar.InferType(retVal.(script.DataTypeable).DataType())
	} else {
		panic("unable to infer return value of function")
	}
}

func (m *Machine) decompileFunction(fn *Function) {
	for !fn.instrs.isEOF() {
		m.decompileStatement(fn)
	}
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
	istr := fn.instrs.peekInstruction()
	fn.emitComment("asm(\"%v\")", istr.String())
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
		fn.instrs.nextInstruction()
	case op == script.OpDrop:
		fn.popNode()
		fn.instrs.nextInstruction()
	case op == script.OpDup:
		duped := fn.peekNode()
		fn.pushNode(duped)
		fn.instrs.nextInstruction()

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
	case op == script.OpExplode:
		m.decompileExplode(fn)

	/* control flow */
	case op == script.OpCall:
		m.decompileCall(fn)
	case op == script.OpCallN:
		m.decompileCall(fn)

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
	istr := fn.instrs.nextInstruction()
	if istr.Operation == script.OpNop {
		return
	}
	fn.emitStatement(AsmStmt{istr.String()})
}

func (m *Machine) decompileCall(fn *Function) {
	istr := fn.instrs.nextInstruction()

	var node Node
	var targetFn *Function

	switch op := istr.Operands.(type) {
	case *script.CallOperands:
		targetFn = m.file.FunctionByAddress(op.Val)
	case *script.CallNOperands:
		targetFn = m.file.FunctionForNative(m.script.HashTable, op.Native)
	}

	args := make([]Node, targetFn.In.Size())
	for i := range args {
		thisArg := fn.popNode()

		if inferrable, ok := thisArg.(TypeInferrable); ok {
			expectedType := targetFn.In.Vars[i].DataType()
			inferrable.InferType(expectedType)
		}

		args[targetFn.In.Size()-i-1] = thisArg
	}

	node = FunctionCall{
		Fn:   targetFn,
		Args: args,
	}

	if targetFn.Out.DataType() != script.VoidType {
		tempDecl := VariableDeclaration{
			Variable: &Variable{
				Identifier: m.genTempIdentifier(),
				Type:       targetFn.Out.DataType(),
			},
			Value: node,
		}
		fn.pushNode(tempDecl.Variable.Reference())
		node = tempDecl
	}

	fn.emitStatement(node)
}

func (m *Machine) decompileReturn(fn *Function) {
	istr := fn.instrs.nextInstruction()
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
	istr := fn.instrs.nextInstruction()
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
	_ = fn.instrs.nextInstruction()
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

func (m *Machine) decompileExplode(fn *Function) {
	_ = fn.instrs.nextInstruction()
	src := fn.popNode()

	length := fn.popNode().(Immediate).Value.(script.ImmediateIntOperands).Int() // ewww
	m.doExplode(fn, src, length)
}

func (m *Machine) doExplode(fn *Function, src Node, length int) {
	for i := 0; i < length; i++ {
		/*
			tempDecl := VariableDeclaration{
				Variable: &Variable{
					Identifier: m.genTempIdentifier(),
					Type:       script.UnknownType,
				},
				Value: ArrayIndex{
					Array: src,
					Index: IntImmediate(uint32(i)),
				},
			}
			fn.emitStatement(tempDecl)
			fn.pushNode(tempDecl.Variable.Reference())
		*/
		fn.pushNode(ArrayIndex{
			Array: src,
			Index: IntImmediate(uint32(i)),
		})
	}
}

func (m *Machine) decompileAssignment(fn *Function) {
	istr := fn.instrs.nextInstruction()
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
	op := fn.instrs.nextInstruction()

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

func (m *Machine) genTempIdentifier() string {
	m.identifierIdx++
	return fmt.Sprintf("temp_%v", m.identifierIdx)
}
