package script

import "fmt"

type Type interface {
	Node
	StackSize() int
}

type ComplexType interface {
	Type
	Explode(v Node, length int) []Node
}

type SimpleType struct {
	Type string
	Size int
}

func (t SimpleType) CString() string {
	return t.Type
}

func (t SimpleType) StackSize() int {
	return t.Size
}

type StructType struct {
	SimpleType
	Fields []*Variable
}

func (s StructType) Explode(v Node, length int) []Node {
	if length != s.StackSize() {
		panic(fmt.Sprintf("explode length %v wasn't equal to stack size %v, don't know what to do\n", length, s.StackSize()))
	}

	result := make([]Node, 0)
	for _, f := range s.Fields {
		var node Node
		node = StructField{
			Struct: v,
			Field:  f,
		}

		result = append(result, node)
	}

	return result
}

type ArrayType struct {
	BaseType Type
	NumElems int
}

func (t ArrayType) CString() string {
	return PtrType{t.BaseType}.CString()
}

func (t ArrayType) StackSize() int {
	return t.NumElems
}

func (t ArrayType) Explode(v Node, length int) []Node {
	result := make([]Node, 0)
	for i := 0; i < length; i++ {
		var node Node
		node = ArrayIndex{
			Array: v,
			Index: IntImmediate(uint32(i)),
		}

		result = append(result, node)
	}

	return result
}

type PtrType struct {
	BaseType Type
}

func (p PtrType) CString() string {
	return fmt.Sprintf("%v*", p.BaseType.CString())
}

func (p PtrType) StackSize() int {
	return 1
}

func (p PtrType) Explode(v Node, length int) []Node {
	switch typ := p.BaseType.(type) {
	case ComplexType:
		return typ.Explode(v, length)
	}

	// Don't know how to explode this.. pretend it's a struct
	fields := make([]*Variable, length)
	for i := range fields {
		fields[i] = &Variable{
			Identifier: fmt.Sprintf("field_%v", i),
			Index:      i,
			Type:       p.BaseType,
		}
	}
	struc := &StructType{
		SimpleType: SimpleType{
			Type: "unknown_struct",
			Size: length,
		},
		Fields: fields,
	}
	return struc.Explode(v, length)
}

type EngineType struct {
	SimpleType
}

func (e EngineType) Explode(v Node, length int) []Node {
	return PtrType{e}.Explode(v, length)
}

var typeMap = map[string]Type{}

var simpleTypes = []SimpleType{
	SimpleType{
		Type: "void",
		Size: 0,
	},
	SimpleType{
		Type: "unknown32",
		Size: 1,
	},
	SimpleType{
		Type: "int",
		Size: 1,
	},
	SimpleType{
		Type: "float",
		Size: 1,
	},
	SimpleType{
		Type: "char*",
		Size: 1,
	},
	SimpleType{
		Type: "BOOL",
		Size: 1,
	},
}

var engineTypes = []Type{
	EngineType{SimpleType{
		Type: "Player",
		Size: 1,
	}},
	EngineType{SimpleType{
		Type: "Ped",
		Size: 1,
	}},
	buildStructType(SimpleType{
		Type: "Entity",
		Size: 37,
	}),
	EngineType{SimpleType{
		Type: "Any",
		Size: 1,
	}},
	EngineType{SimpleType{
		Type: "Object",
		Size: 1,
	}},
	EngineType{SimpleType{
		Type: "Vehicle",
		Size: 1,
	}},
	EngineType{SimpleType{
		Type: "Pickup",
		Size: 1,
	}},
	EngineType{SimpleType{
		Type: "Hash",
		Size: 1,
	}},
	EngineType{SimpleType{
		Type: "ScrHandle",
		Size: 1,
	}},
	EngineType{SimpleType{
		Type: "Cam",
		Size: 1,
	}},
	EngineType{SimpleType{
		Type: "Blip",
		Size: 1,
	}},
}

func buildStructType(typ SimpleType) StructType {
	struc := StructType{
		SimpleType: typ,
		Fields:     make([]*Variable, typ.StackSize()),
	}

	for i := range struc.Fields {
		struc.Fields[i] = &Variable{
			Identifier: fmt.Sprintf("field_%v", i),
			Index:      i,
			Type:       UnknownType,
		}
	}

	return struc
}

func init() {
	for _, t := range simpleTypes {
		typeMap[t.Type] = t
	}

	for _, t := range engineTypes {
		typeMap[t.CString()] = t
	}

	typeMap["Vector3"] = StructType{
		SimpleType: SimpleType{
			Type: "Vector3",
			Size: 3,
		},
		Fields: []*Variable{
			&Variable{
				Identifier: "x",
				Type:       GetType("float"),
			},
			&Variable{
				Identifier: "y",
				Type:       GetType("float"),
			},
			&Variable{
				Identifier: "z",
				Type:       GetType("float"),
			},
		},
	}

	UnknownType = GetType("void*")
	IntType = GetType("int")
	VoidType = GetType("void")
	FloatType = GetType("float")
	StringType = GetType("char*")
	Vector3Type = GetType("Vector3")
}

var (
	UnknownType Type
	IntType     Type
	VoidType    Type
	FloatType   Type
	StringType  Type
	Vector3Type Type
)

func GetType(s string) Type {
	if t, ok := typeMap[s]; ok {
		return t
	}

	if Token(s[len(s)-1]) == DeRefToken {
		return PtrType{
			BaseType: GetType(s[:len(s)-1]),
		}
	}

	panic(fmt.Sprintf("no such type: %v", s))
}

func guessType(stackSize int) Type {
	switch {
	case stackSize == 0:
		return VoidType
	case stackSize == 1:
		return UnknownType
	case stackSize == 3:
		return Vector3Type
	default:
		return ArrayType{
			BaseType: guessType(1),
			NumElems: stackSize,
		}
	}
}

type DataTypeable interface {
	DataType() Type
}
