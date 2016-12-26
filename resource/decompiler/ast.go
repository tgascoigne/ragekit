package decompiler

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/tgascoigne/ragekit/resource/script"
)

// Tokens
type Token string

const (
	AddToken    Token = "+"
	SubToken    Token = "-"
	MulToken    Token = "*"
	DivToken    Token = "/"
	ModToken    Token = "%"
	AndToken    Token = "&"
	OrToken     Token = "|"
	XorToken    Token = "^"
	NotToken    Token = "!"
	NegToken    Token = "-"
	DeRefToken  Token = "*"
	ReturnToken Token = "return"
)

func (t Token) CString() string {
	return string(t)
}

// A CStringer is something which can be printed to C Code
type CStringer interface {
	CString() string
}

// A File is the top level AST Object
type File struct {
	Decls Declarations
	Nodes []Node
}

func (f File) CString() string {
	buf := new(bytes.Buffer)

	fmt.Fprintf(buf, "%v\n", f.Decls.CString())

	for _, n := range f.Nodes {
		fmt.Fprintf(buf, "%v\n", n.CString())
	}

	return buf.String()
}

type TypeInferrable interface {
	InferType(typ script.Type)
}

// A Node is an element in the C AST
type Node interface {
	CStringer
}

type Comment string

func (c Comment) CString() string {
	return "/* " + string(c) + " */"
}

// An Immediate is a plain old immediate value
type Immediate struct {
	Value script.Operands /* We expect this to be one of the immediate operands.. */
}

func (i Immediate) DataType() script.Type {
	return i.Value.(script.DataTypeable).DataType()
}

func (i Immediate) CString() string {
	return i.Value.String()
}

type Declarations struct {
	Vars []*VariableDeclaration
}

func (d *Declarations) AddVariable(decl *VariableDeclaration) {
	d.Vars = append(d.Vars, decl)
}

func (d *Declarations) VariableByName(identifier string) *Variable {
	for _, v := range d.Vars {
		if v.Identifier == identifier {
			return v.Variable
		}
	}

	panic(fmt.Sprintf("no variable declaration with name %v\n", identifier))
}

func (d *Declarations) CString() string {
	decls := make([]string, len(d.Vars))
	for i, v := range d.Vars {
		decls[i] = fmt.Sprintf("%v;\n", v.CString())
	}

	return fmt.Sprintf("%v", strings.Join(decls, ""))
}

// A Variable is an assignable memory location
type Variable struct {
	Identifier string
	Type       script.Type

	typeInferred bool
}

func (v *Variable) CString() string {
	return v.Identifier
}

func (v *Variable) Declaration() *VariableDeclaration {
	return &VariableDeclaration{
		Variable: v,
		Value:    nil,
	}
}

func (v *Variable) Reference() *VariableReference {
	return &VariableReference{
		Variable: v,
	}
}

func (v *Variable) DeclarationWithValue(value interface{}) *VariableDeclaration {
	return &VariableDeclaration{
		Variable: v,
		Value:    value,
	}
}

func (v *Variable) DataType() script.Type {
	return v.Type
}

func (v *Variable) InferType(typ script.Type) {
	if v.typeInferred && typ != v.Type {
		fmt.Printf("type of variable %v ambiguous", v.Identifier)
		return
	}

	fmt.Printf("inferring type %v = %v\n", v.Identifier, typ)

	v.typeInferred = true
	v.Type = typ
}

type VariableReference struct {
	*Variable
}

func (r *VariableReference) CString() string {
	return r.Identifier
}

// A VariableDeclaration declares a Variable
type VariableDeclaration struct {
	*Variable
	Value interface{}
}

func (d VariableDeclaration) CString() string {
	if d.Value != nil {
		return fmt.Sprintf("%v %v = %v", d.Type, d.Identifier, d.Value)
	} else {
		return fmt.Sprintf("%v %v", d.Type, d.Identifier)
	}
}

// A Function accepts inputs, executes a set of Nodes, and returns a single output
type Function struct {
	Identifier string
	In         Declarations
	Out        *Variable
	Decls      Declarations
	Statements []Node

	retInstrs []script.RetOperands

	nodeStack    []Node
	nodeStackIdx int
}

func (fn Function) CString() string {
	stmts := make([]string, len(fn.Statements))
	for i, s := range fn.Statements {
		stmts[i] = fmt.Sprintf("\t%v;\n", s.CString())
	}

	args := make([]string, len(fn.In.Vars))
	for i, arg := range fn.In.Vars {
		args[i] = arg.CString()
	}

	return fmt.Sprintf("%v %v(%v) {\n%v\n%v}", fn.Out.Type.CString(), fn.Identifier, strings.Join(args, ", "), fn.Decls.CString(), strings.Join(stmts, ""))
}

// An AssignStmt assigns a value to a Variable
type AssignStmt struct {
	Dest  Node
	Value Node
}

func (st AssignStmt) CString() string {
	return fmt.Sprintf("%v = %v", st.Dest.CString(), st.Value.CString())
}

type ReturnStmt struct {
	Value Node
}

func (st ReturnStmt) CString() string {
	if st.Value != nil {
		return fmt.Sprintf("%v %v", ReturnToken, st.Value.CString())
	}
	return fmt.Sprintf("%v", ReturnToken)
}

// A BinaryExpr performs an operation on two nodes
type BinaryExpr struct {
	Left  Node
	Op    Token
	Right Node
}

func (expr BinaryExpr) DataType() script.Type {
	return expr.Left.(script.DataTypeable).DataType()
}

func (expr BinaryExpr) CString() string {
	return fmt.Sprintf("%v %v %v", expr.Left.CString(), expr.Op.CString(), expr.Right.CString())
}

// A UnaryExpr performs an operation on one node
type UnaryExpr struct {
	Op   Token
	Node Node
}

func (expr UnaryExpr) DataType() script.Type {
	return expr.Node.(script.DataTypeable).DataType()
}

func (expr UnaryExpr) CString() string {
	return fmt.Sprintf("%v%v", expr.Op.CString(), expr.Node.CString())
}

type DeRefExpr struct {
	Node Node
}

func (expr DeRefExpr) CString() string {
	return UnaryExpr{Op: DeRefToken, Node: expr.Node}.CString()
}

func (expr DeRefExpr) InferType(typ script.Type) {
	expr.Node.(*Variable).InferType(typ)
}

// An AsmExpr performs inline assembly
type AsmStmt struct {
	Asm string
}

func (expr AsmStmt) CString() string {
	return fmt.Sprintf("asm(\"%v\")", expr.Asm)
}

type ArrayLiteral []Node

func (arr ArrayLiteral) CString() string {
	elems := make([]string, len(arr))
	for i := range elems {
		elems[i] = arr[i].CString()
	}
	return fmt.Sprintf("{%v}", strings.Join(elems, ", "))
}
