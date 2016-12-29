package script

import (
	"fmt"
	"strings"
)

// A Function accepts inputs, executes a set of Nodes, and returns a single output
type Function struct {
	Identifier string
	In         Declarations
	Out        *Variable
	Address    uint32
	Decls      Declarations

	*BasicBlock
}

func NewFunction() *Function {
	fn := &Function{}
	block := &BasicBlock{
		ParentFunc: fn,
		instrs:     &Instructions{},
	}
	fn.BasicBlock = block
	return fn
}

func (fn *Function) CString() string {
	args := make([]string, len(fn.In.Vars))
	for i, arg := range fn.In.Vars {
		args[i] = arg.CString()
	}

	return fmt.Sprintf("%v %v(%v) {\n%v\n%v}", fn.Out.Type.CString(), fn.Identifier, strings.Join(args, ", "), fn.Decls.CString(), fn.BasicBlock.CString())
}

func (fn *Function) inferReturnType(ret Instruction) {
	retVar := fn.Out
	op := ret.Operands.(*RetOperands)
	if op.NumReturnVals == 0 {
		retVar.InferType(VoidType)
	} else if op.NumReturnVals == 1 {
		retVal := fn.peekNode()
		retVar.InferType(retVal.(DataTypeable).DataType())
	} else {
		panic("unable to infer return value of function")
	}
}

type BasicBlock struct {
	Statements []Node
	ParentFunc *Function

	instrs *Instructions

	nodeStack    []Node
	nodeStackIdx int
}

func (b *BasicBlock) VariableByName(identifier string) *Variable {
	return b.ParentFunc.Decls.VariableByName(identifier)
}

func (block *BasicBlock) emitStatement(stmt Node) {
	block.Statements = append(block.Statements, stmt)
}

func (block *BasicBlock) emitComment(format string, args ...interface{}) {
	commentStr := fmt.Sprintf(format, args...)
	comment := Comment(commentStr)
	block.emitStatement(comment)
	fmt.Println(commentStr)
}

func (block *BasicBlock) pushNode(node Node) {
	block.nodeStack = append(block.nodeStack, node)
	block.nodeStackIdx++
	//block.emitComment("pushing %v at stack idx %v", node.CString(), block.nodeStackIdx)
}

func (block *BasicBlock) popNode() Node {
	if block.nodeStackIdx <= 0 {
		fmt.Println("node stack underflow")
		return Immediate{&Immediate32Operands{Val: 0xBABE}}
	}

	node := block.peekNode()
	block.nodeStackIdx--
	block.nodeStack = block.nodeStack[:block.nodeStackIdx]
	//	fmt.Printf("popping %v %v\n", node.CString(), block.nodeStackIdx)
	return node
}

func (block *BasicBlock) peekNode() Node {
	if block.nodeStackIdx <= 0 {
		//		fmt.Printf("peek %v\n", block.nodeStackIdx)
		fmt.Println("node stack underflow")
		return Immediate{&Immediate32Operands{Val: 0xBABE}}
	}

	return block.nodeStack[block.nodeStackIdx-1]
}

func (block *BasicBlock) CString() string {
	stmts := make([]string, len(block.Statements))
	for i, s := range block.Statements {
		stmts[i] = fmt.Sprintf("\t%v;\n", s.CString())
	}

	return strings.Join(stmts, "")
}
