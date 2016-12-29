package script

import (
	"bytes"
	"fmt"
	"strings"
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
	RefToken    Token = "&"
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
	Decls     Declarations
	Functions []*Function
	Nodes     []Node
}

func (f File) FunctionByName(identifier string) *Function {
	for _, fn := range f.Functions {
		if fn.Identifier == identifier {
			return fn
		}
	}

	panic(fmt.Sprintf("unknown function: ", identifier))
}

func (f File) FunctionByAddress(addr uint32) *Function {
	for _, fn := range f.Functions {
		if fn.Address == addr {
			return fn
		}
	}

	panic(fmt.Sprintf("unknown function with address: %x", addr))
}

func (f File) FunctionForNative(db *NativeDB, native Native64) *Function {
	spec := db.LookupNative(native)
	if spec == nil {
		panic(fmt.Sprintf("unknown function with hash: %x", native))
	}

	fn := &Function{
		Identifier: spec.Name,
		Out: &Variable{
			Identifier: "ERROR",
			Type:       spec.Results,
		},
	}

	for _, param := range spec.Params {
		fn.In.AddVariable(&VariableDeclaration{
			Variable: &Variable{
				Identifier: param.Name,
				Type:       param.Type,
			},
		})
	}

	return fn
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
	InferType(typ Type)
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
	Value Operands /* We expect this to be one of the immediate operands.. */
}

func (i Immediate) DataType() Type {
	return i.Value.(DataTypeable).DataType()
}

func (i Immediate) CString() string {
	return i.Value.String()
}

func IntImmediate(val uint32) Node {
	return Immediate{
		&Immediate32Operands{
			Val: val,
		},
	}
}

type Declarations struct {
	Vars []*VariableDeclaration
}

func (d *Declarations) Size() int {
	return len(d.Vars)
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
	Type       Type

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

func (v *Variable) DataType() Type {
	return v.Type
}

func (v *Variable) InferType(typ Type) {
	if v.typeInferred && typ.CString() != v.Type.CString() {
		fmt.Printf("WARNING: type of variable %v ambiguous\n", v.Identifier)
		//return
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
		valueStr := fmt.Sprintf("%v", d.Value)
		if val, ok := d.Value.(CStringer); ok {
			valueStr = val.CString()
		}
		return fmt.Sprintf("%v %v = %v", d.Type.CString(), d.Identifier, valueStr)
	} else {
		return fmt.Sprintf("%v %v", d.Type.CString(), d.Identifier)
	}
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

func (expr BinaryExpr) DataType() Type {
	return expr.Left.(DataTypeable).DataType()
}

func (expr BinaryExpr) CString() string {
	return fmt.Sprintf("%v %v %v", expr.Left.CString(), expr.Op.CString(), expr.Right.CString())
}

// A UnaryExpr performs an operation on one node
type UnaryExpr struct {
	Op   Token
	Node Node
}

func (expr UnaryExpr) DataType() Type {
	return expr.Node.(DataTypeable).DataType()
}

func (expr UnaryExpr) CString() string {
	return fmt.Sprintf("%v%v", expr.Op.CString(), expr.Node.CString())
}

type PtrNode struct {
	Node Node
}

func (expr PtrNode) DeRef() Node {
	return expr.Node
}

func (expr PtrNode) CString() string {
	return fmt.Sprintf("%v%v", RefToken, expr.Node.CString())
}

func (expr PtrNode) InferType(typ Type) {
	expr.Node.(*Variable).InferType(typ)
}

func (expr PtrNode) DataType() Type {
	return expr.Node.(*Variable).DataType()
}

type DeRefExpr struct {
	Node Node
}

func (expr DeRefExpr) CString() string {
	return UnaryExpr{Op: DeRefToken, Node: expr.Node}.CString()
}

func (expr DeRefExpr) InferType(typ Type) {
	expr.Node.(*Variable).InferType(typ)
}

func (expr DeRefExpr) DataType() Type {
	return expr.Node.(*Variable).DataType()
}

// An AsmExpr performs inline assembly
type AsmStmt struct {
	Asm string
}

func (expr AsmStmt) CString() string {
	return fmt.Sprintf("asm(\"%v\")", expr.Asm)
}

// A FuntionCall is an invocation of a local function
type FunctionCall struct {
	Fn   *Function
	Args []Node
}

func (fc FunctionCall) CString() string {
	elems := make([]string, len(fc.Args))
	for i := range elems {
		elems[i] = fc.Args[i].CString()
	}

	return fmt.Sprintf("%v(%v)", fc.Fn.Identifier, strings.Join(elems, ", "))
}

// An ArrayLiteral is a literal representation of an array
type ArrayLiteral []Node

func (arr ArrayLiteral) CString() string {
	elems := make([]string, len(arr))
	for i := range elems {
		elems[i] = arr[i].CString()
	}
	return fmt.Sprintf("{%v}", strings.Join(elems, ", "))
}

type ArrayIndex struct {
	Array Node
	Index Node
}

func (idx ArrayIndex) DataType() Type {
	return UnknownType
}

func (idx ArrayIndex) CString() string {
	if ptr, ok := idx.Array.(PtrNode); ok {
		return fmt.Sprintf("%v->[%v]", ptr.DeRef().CString(), idx.Index.CString())
	}
	return fmt.Sprintf("%v.[%v]", idx.Array.CString(), idx.Index.CString())
}

type StructField struct {
	Struct Node
	Field  *Variable
}

func (s StructField) DataType() Type {
	return s.Field.DataType()
}

func (s StructField) CString() string {
	if ptr, ok := s.Struct.(PtrNode); ok {
		return fmt.Sprintf("%v->%v", ptr.DeRef().CString(), s.Field.CString())
	}
	return fmt.Sprintf("%v.%v", s.Struct.CString(), s.Field.CString())
}

func (s StructField) InferType(typ Type) {
	s.Field.InferType(typ)
}
