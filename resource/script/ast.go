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
	AndToken    Token = "&&"
	OrToken     Token = "||"
	XorToken    Token = "^"
	NotToken    Token = "!"
	NegToken    Token = "-"
	RefToken    Token = "&"
	DeRefToken  Token = "*"
	IfToken     Token = "if"
	NotEqToken  Token = "!="
	ReturnToken Token = "return"
	ExternToken Token = "extern"
	StaticToken Token = "static"
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

func (f *File) GlobalByIndex(index int) Node {
	identifier := fmt.Sprintf("global_%v", index)
	if !f.Decls.HasVariable(identifier) {
		decl := &VariableDeclaration{
			Variable: &Variable{
				Identifier: identifier,
				Type:       UnknownType,
			},
			Scope: ExternToken,
		}
		fmt.Printf("adding global decl %v\n", decl.CString())
		f.Decls.AddVariable(decl)
	}

	return f.Decls.VariableByName(identifier)
}

func (f *File) FunctionByName(identifier string) *Function {
	for _, fn := range f.Functions {
		if fn.Identifier == identifier {
			return fn
		}
	}

	panic(fmt.Sprintf("unknown function: ", identifier))
}

func (f *File) FunctionByAddress(addr uint32) *Function {
	for _, fn := range f.Functions {
		if fn.Address == addr {
			return fn
		}
	}

	panic(fmt.Sprintf("unknown function with address: %x", addr))
}

func (f *File) FunctionForNative(db *NativeDB, operands *CallNOperands) (fn *Function, generated bool) {
	native := operands.Native
	spec := db.LookupNative(native)
	if spec == nil {
		// No spec for this native, create one using the calln operands
		fmt.Printf("WARNING: unknown function with hash: %x\n", native)
		numIn, numOut := operands.InSize, operands.OutSize

		params := make([]NativeParam, numIn)
		for i := 0; i < int(numIn); i++ {
			params[i] = NativeParam{
				Name: fmt.Sprintf("arg_%v", i),
				Type: GetType("Any"),
			}
		}

		spec = &NativeSpec{
			Name:   fmt.Sprintf("unk_%x", native),
			Params: params,
			Results: ArrayType{
				BaseType: UnknownType,
				NumElems: int(numOut),
			},
		}

		// Let the caller know that we it's been auto-generated
		generated = true
	}

	fn = &Function{
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

	return fn, false
}

func (f *File) CString() string {
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

func (d *Declarations) HasVariable(identifier string) bool {
	for _, v := range d.Vars {
		if v.Identifier == identifier {
			return true
		}
	}

	return false
}

func (d *Declarations) VariableByIndex(index int) *Variable {
	for _, v := range d.Vars {
		if v.Index == index {
			v.IsReferenced = true
			return v.Variable
		}
	}

	panic(fmt.Sprintf("no variable declaration with index %v\n", index))
}

func (d *Declarations) VariableByName(identifier string) *Variable {
	for _, v := range d.Vars {
		if v.Identifier == identifier {
			v.IsReferenced = true
			return v.Variable
		}
	}

	panic(fmt.Sprintf("no variable declaration with name %v\n", identifier))
}

func (d *Declarations) CString() string {
	decls := make([]string, 0)
	for _, v := range d.Vars {
		if v.IsArgument || !v.IsReferenced {
			continue
		}
		decls = append(decls, fmt.Sprintf("%v;\n", v.CString()))
	}

	return fmt.Sprintf("%v", strings.Join(decls, ""))
}

// A Variable is an assignable memory location
type Variable struct {
	Index        int
	Identifier   string
	Type         Type
	IsArgument   bool // used to omit func arguments from the main set of local decls
	IsReferenced bool // used to omit locals which are unused

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

	fmt.Printf("inferring type %v = %#v\n", v.CString(), typ)

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
	Scope Token
}

func (d *VariableDeclaration) CString() string {
	var valueStr, specifierStr string

	if d.Value != nil {
		valueStr = fmt.Sprintf(" = %v", d.Value)
		if val, ok := d.Value.(CStringer); ok {
			valueStr = fmt.Sprintf(" = %v", val.CString())
		}
	}

	if d.Scope != Token("") {
		specifierStr = fmt.Sprintf("%v ", d.Scope)
	}

	return fmt.Sprintf("%v%v %v%v", specifierStr, d.Type.CString(), d.Identifier, valueStr)
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
	if ptrTyp, ok := typ.(PtrType); ok {
		typ = ptrTyp.BaseType
	} else {
		fmt.Printf("WARNING: expected inferred type to be a pointer\n")
	}

	expr.Node.(TypeInferrable).InferType(typ)
}

func (expr PtrNode) DataType() Type {
	return PtrType{expr.Node.(DataTypeable).DataType()}
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
	switch typ := idx.Array.(type) {
	case ArrayType:
		return typ.BaseType
	case PtrType:
		return typ.BaseType
	case ArrayIndex:
		return typ.DataType()
	}
	return UnknownType
}

func (idx ArrayIndex) CString() string {
	if ptr, ok := idx.Array.(PtrNode); ok {
		return fmt.Sprintf("%v[%v]", ptr.DeRef().CString(), idx.Index.CString())
	}
	return fmt.Sprintf("%v[%v]", idx.Array.CString(), idx.Index.CString())
}

func (idx ArrayIndex) InferType(typ Type) {
	idx.Array.(TypeInferrable).InferType(ArrayType{
		BaseType: typ,
		NumElems: 0, // can we guess the number of elems from anywhere?
	})
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

type IfStmt struct {
	Cond Node
	Then *BasicBlock
	Else *BasicBlock
}

func (s IfStmt) CString() string {
	buf := new(bytes.Buffer)

	cond, then, els := s.Cond, s.Then, s.Else

	// simplify empty 'then' blocks
	// swap then with else, and invert cond
	if then.Empty() {
		cond = NotCond{cond}
		then = els
		els = nil
	}

	fmt.Fprintf(buf, "%v (%v) {\n%v\n}", IfToken, cond.CString(), then.CString())
	if els != nil {
		fmt.Fprintf(buf, " else {\n%v\n}", els.CString())
	}
	return buf.String()
}

type NotCond struct {
	Node
}

func (c NotCond) CString() string {
	if c, ok := c.Node.(NotCond); ok {
		// Simplify double negations
		return c.Node.CString()
	}

	return UnaryExpr{
		Node: c.Node,
		Op:   NotToken,
	}.CString()
}

type AndCond struct {
	A, B Node
}

func (c AndCond) CString() string {
	return BinaryExpr{
		Left:  c.A,
		Op:    AndToken,
		Right: c.B,
	}.CString()
}

type OrCond struct {
	A, B Node
}

func (c OrCond) CString() string {
	return BinaryExpr{
		Left:  c.A,
		Op:    OrToken,
		Right: c.B,
	}.CString()
}

type XorCond struct {
	A, B Node
}

func (c XorCond) CString() string {
	return BinaryExpr{
		Left:  c.A,
		Op:    XorToken,
		Right: c.B,
	}.CString()
}
