package script

import (
	"fmt"
	"strings"

	"github.com/tgascoigne/ragekit/resource"
)

type Operands interface {
	String() string /* first is the operand string, second is a mnemonic suffix */
	Unpack(*Instruction, *Script, *resource.Container)
}

type InitOperandFunc func() Operands

var OperandFunc = map[uint8]InitOperandFunc{
	37:  func() Operands { return &Immediate8Operands{} },
	38:  func() Operands { return &Immediate8x2Operands{} },
	39:  func() Operands { return &Immediate8x3Operands{} },
	40:  func() Operands { return &Immediate32Operands{} },
	41:  func() Operands { return &ImmediateF32Operands{} },
	44:  func() Operands { return &CallNOperands{} },
	45:  func() Operands { return &EnterOperands{} },
	46:  func() Operands { return &RetOperands{} },
	52:  func() Operands { return &Immediate8Operands{} },
	53:  func() Operands { return &Immediate8Operands{} },
	54:  func() Operands { return &Immediate8Operands{} },
	55:  func() Operands { return &Immediate8Operands{} },
	56:  func() Operands { return &Immediate8Operands{} },
	57:  func() Operands { return &Immediate8Operands{} },
	58:  func() Operands { return &Immediate8Operands{} },
	59:  func() Operands { return &Immediate8Operands{} },
	60:  func() Operands { return &Immediate8Operands{} },
	61:  func() Operands { return &Immediate8Operands{} },
	62:  func() Operands { return &Immediate8Operands{} },
	64:  func() Operands { return &Immediate8Operands{} },
	65:  func() Operands { return &Immediate8Operands{} },
	66:  func() Operands { return &Immediate8Operands{} },
	67:  func() Operands { return &Immediate16Operands{} },
	68:  func() Operands { return &Immediate16Operands{} },
	69:  func() Operands { return &Immediate16Operands{} },
	70:  func() Operands { return &Immediate16Operands{} },
	71:  func() Operands { return &Immediate16Operands{} },
	72:  func() Operands { return &Immediate16Operands{} },
	73:  func() Operands { return &Immediate16Operands{} },
	74:  func() Operands { return &Immediate16Operands{} },
	75:  func() Operands { return &Immediate16Operands{} },
	76:  func() Operands { return &Immediate16Operands{} },
	77:  func() Operands { return &Immediate16Operands{} },
	78:  func() Operands { return &Immediate16Operands{} },
	79:  func() Operands { return &Immediate16Operands{} },
	80:  func() Operands { return &Immediate16Operands{} },
	81:  func() Operands { return &Immediate16Operands{} },
	82:  func() Operands { return &Immediate16Operands{} },
	83:  func() Operands { return &Immediate16Operands{} },
	84:  func() Operands { return &Immediate16Operands{} },
	85:  func() Operands { return &BranchOperands{} },
	86:  func() Operands { return &BranchOperands{} },
	87:  func() Operands { return &BranchOperands{} },
	88:  func() Operands { return &BranchOperands{} },
	89:  func() Operands { return &BranchOperands{} },
	90:  func() Operands { return &BranchOperands{} },
	91:  func() Operands { return &BranchOperands{} },
	92:  func() Operands { return &BranchOperands{} },
	93:  func() Operands { return &CallOperands{} },
	94:  func() Operands { return &Immediate24Operands{} },
	95:  func() Operands { return &Immediate24Operands{} },
	96:  func() Operands { return &Immediate24Operands{} },
	97:  func() Operands { return &Immediate24Operands{} },
	98:  func() Operands { return &SwitchOperands{} },
	99:  func() Operands { return &StringOperands{} },
	101: func() Operands { return &Immediate8Operands{} },
	102: func() Operands { return &Immediate8Operands{} },
	103: func() Operands { return &Immediate8Operands{} },
	104: func() Operands { return &Immediate8Operands{} },
	109: func() Operands { return &ImplicitOperands{offset: 110} },
	110: func() Operands { return &ImplicitOperands{offset: 110} },
	111: func() Operands { return &ImplicitOperands{offset: 110} },
	112: func() Operands { return &ImplicitOperands{offset: 110} },
	113: func() Operands { return &ImplicitOperands{offset: 110} },
	114: func() Operands { return &ImplicitOperands{offset: 110} },
	115: func() Operands { return &ImplicitOperands{offset: 110} },
	116: func() Operands { return &ImplicitOperands{offset: 110} },
	117: func() Operands { return &ImplicitOperands{offset: 110} },
	118: func() Operands { return &ImplicitFOperands{offset: 119} },
	119: func() Operands { return &ImplicitFOperands{offset: 119} },
	120: func() Operands { return &ImplicitFOperands{offset: 119} },
	121: func() Operands { return &ImplicitFOperands{offset: 119} },
	122: func() Operands { return &ImplicitFOperands{offset: 119} },
	123: func() Operands { return &ImplicitFOperands{offset: 119} },
	124: func() Operands { return &ImplicitFOperands{offset: 119} },
	125: func() Operands { return &ImplicitFOperands{offset: 119} },
	126: func() Operands { return &ImplicitFOperands{offset: 119} },
}

type NoOperands struct{}

func (op *NoOperands) String() string {
	return ""
}

func (op *NoOperands) Unpack(istr *Instruction, script *Script, res *resource.Container) {}

type Immediate8Operands struct {
	Val uint8
}

func (op *Immediate8Operands) String() string {
	return fmt.Sprintf("%v", op.Val)
}

func (op *Immediate8Operands) Unpack(istr *Instruction, script *Script, res *resource.Container) {
	res.Parse(&op.Val)
}

type Immediate8x2Operands struct {
	Val0, Val1 uint8
}

func (op *Immediate8x2Operands) String() string {
	return fmt.Sprintf("%v %v", op.Val0, op.Val1)
}

func (op *Immediate8x2Operands) Unpack(istr *Instruction, script *Script, res *resource.Container) {
	res.Parse(&op.Val0)
	res.Parse(&op.Val1)
}

type Immediate8x3Operands struct {
	Val0, Val1, Val2 uint8
}

func (op *Immediate8x3Operands) String() string {
	return fmt.Sprintf("%v %v %v", op.Val0, op.Val1, op.Val2)
}

func (op *Immediate8x3Operands) Unpack(istr *Instruction, script *Script, res *resource.Container) {
	res.Parse(&op.Val0)
	res.Parse(&op.Val1)
	res.Parse(&op.Val2)
}

type Immediate24Operands struct {
	Val uint32
}

func (op *Immediate24Operands) String() string {
	return fmt.Sprintf("%v", op.Val)
}

func (op *Immediate24Operands) Unpack(istr *Instruction, script *Script, res *resource.Container) {
	var val0, val1, val2 uint8
	res.Parse(&val0)
	res.Parse(&val1)
	res.Parse(&val2)
	op.Val = uint32(val0)
	op.Val <<= 8
	op.Val += uint32(val1)
	op.Val <<= 8
	op.Val += uint32(val2)
}

type Immediate16Operands struct {
	Val uint16
}

func (op *Immediate16Operands) String() string {
	return fmt.Sprintf("%v", op.Val)
}

func (op *Immediate16Operands) Unpack(istr *Instruction, script *Script, res *resource.Container) {
	res.Parse(&op.Val)
}

type Immediate32Operands struct {
	Val      uint32
	HashStrs []string
}

func (op *Immediate32Operands) String() string {
	var hashMatches string
	if len(op.HashStrs) > 0 {
		hashMatches = fmt.Sprintf(" <%v>", strings.Join(op.HashStrs, ","))
	}
	return fmt.Sprintf("%v%v", op.Val, hashMatches)
}

func (op *Immediate32Operands) Unpack(istr *Instruction, script *Script, res *resource.Container) {
	res.Parse(&op.Val)

	fmt.Printf("FIXME: Immediate32 operand!\n")

	/*if hashStrs, ok := script.HashLookup(op.Val); ok {
		op.HashStrs = hashStrs
	} else {
		op.HashStrs = []string{}
	}*/
}

type ImmediateF32Operands struct {
	Val float32
}

func (op *ImmediateF32Operands) String() string {
	return fmt.Sprintf("%v", op.Val)
}

func (op *ImmediateF32Operands) Unpack(istr *Instruction, script *Script, res *resource.Container) {
	res.Parse(&op.Val)
}

type BranchOperands struct {
	RelativeAddr int16
	AbsoluteAddr uint32
}

func (op *BranchOperands) String() string {
	return fmt.Sprintf("%.8x", op.AbsoluteAddr)
}

func (op *BranchOperands) Unpack(istr *Instruction, script *Script, res *resource.Container) {
	res.Parse(&op.RelativeAddr)
	/* relative to the end of this instruction, hence +3 */
	op.AbsoluteAddr = uint32(int32(istr.Address) + int32(op.RelativeAddr) + 3)
}

type CallNOperands struct {
	ParamSize  uint8
	Native     Native64
	NativeStrs []string
}

func (op *CallNOperands) String() string {
	return fmt.Sprintf("%x %v <%v>", op.Native, op.ParamSize, strings.Join(op.NativeStrs, ","))
}

func (op *CallNOperands) Unpack(istr *Instruction, script *Script, res *resource.Container) {
	var nativeIndex uint16
	res.Parse(&op.ParamSize)
	res.ParseBigEndian(&nativeIndex)

	op.Native = script.NativeTable[nativeIndex]
	if nativeStrs, ok := script.HashLookup(op.Native); ok {
		op.NativeStrs = nativeStrs
	} else {
		op.NativeStrs = []string{"unknown"}
	}
}

type CallOperands struct {
	Val uint32
}

func (op *CallOperands) String() string {
	return fmt.Sprintf("%.8x", op.Val)
}

func (op *CallOperands) Unpack(istr *Instruction, script *Script, res *resource.Container) {
	var val0, val1, val2 uint8
	res.Parse(&val0)
	res.Parse(&val1)
	res.Parse(&val2)
	op.Val = uint32(val2)
	op.Val <<= 8
	op.Val += uint32(val1)
	op.Val <<= 8
	op.Val += uint32(val0)
}

type EnterOperands struct {
	NumArgs  uint8
	Unknown1 uint8
	Unknown2 uint8
	Unknown3 uint8
}

func (op *EnterOperands) String() string {
	return fmt.Sprintf("%v %v %v %v", op.NumArgs, op.Unknown1, op.Unknown2, op.Unknown3)
}

func (op *EnterOperands) Unpack(istr *Instruction, script *Script, res *resource.Container) {
	res.Parse(&op.NumArgs)
	res.Parse(&op.Unknown1)
	res.Parse(&op.Unknown2)
	res.Parse(&op.Unknown3)
}

type RetOperands struct {
	NumParams     uint8
	NumReturnVals uint8
}

func (op *RetOperands) String() string {
	return fmt.Sprintf("%v %v", op.NumParams, op.NumReturnVals)
}

func (op *RetOperands) Unpack(istr *Instruction, script *Script, res *resource.Container) {
	res.Parse(&op.NumParams)
	res.Parse(&op.NumReturnVals)
}

type ImplicitOperands struct {
	offset int
	Val    int
}

func (op *ImplicitOperands) String() string {
	return fmt.Sprintf("%v", op.Val)
}

func (op *ImplicitOperands) Unpack(istr *Instruction, script *Script, res *resource.Container) {
	op.Val = int(istr.Opcode) - op.offset
}

type ImplicitFOperands struct {
	offset int
	Val    float32
}

func (op *ImplicitFOperands) String() string {
	return fmt.Sprintf("%v", op.Val)
}

func (op *ImplicitFOperands) Unpack(istr *Instruction, script *Script, res *resource.Container) {
	op.Val = float32(int(istr.Opcode) - op.offset)
}

type SwitchOperands struct {
	JumpTableRel map[uint32]uint16
	JumpTableAbs map[uint32]uint32
	HashStrs     map[uint32]string
}

func (op *SwitchOperands) String() string {
	targets := make([]string, 0)
	for cond, addr := range op.JumpTableAbs {
		targets = append(targets, fmt.Sprintf("%v%v: %.8x", cond, op.HashStrs[cond], addr))
	}

	return strings.Join(targets, ", ")
}

func (op *SwitchOperands) Unpack(istr *Instruction, script *Script, res *resource.Container) {
	var length uint8
	op.JumpTableRel = make(map[uint32]uint16)
	op.JumpTableAbs = make(map[uint32]uint32)
	op.HashStrs = make(map[uint32]string)
	res.Parse(&length)
	for i := 0; i < int(length); i++ {
		var value uint32
		var relAddr uint16
		res.Parse(&value)
		res.Parse(&relAddr)
		curAddrVirt := istr.Address + uint32(2+((i+1)*6))
		op.JumpTableRel[value] = relAddr
		op.JumpTableAbs[value] = curAddrVirt + uint32(relAddr)

		//var hashMatches string
		//hashStrs, _ := script.HashLookup(value)
		//if len(hashStrs) > 0 {
		//	hashMatches = fmt.Sprintf(" <%v>", strings.Join(hashStrs, ","))
		//}

		op.HashStrs[value] = " <unknown>"
	}
}

type StringOperands struct {
	Val string
}

func (op *StringOperands) String() string {
	return fmt.Sprintf("\"%v\"", op.Val)
}

func (op *StringOperands) Unpack(istr *Instruction, script *Script, res *resource.Container) {

}
