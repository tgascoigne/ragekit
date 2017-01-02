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

type ImmediateIntOperands interface {
	Int() int
}

type NoOperands struct{}

func (op *NoOperands) String() string {
	return ""
}

func (op *NoOperands) Unpack(istr *Instruction, script *Script, res *resource.Container) {}

type Immediate8Operands struct {
	Val uint8
}

func (op *Immediate8Operands) DataType() Type {
	return IntType
}

func (op *Immediate8Operands) Int() int {
	return int(op.Val)
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

func (op *Immediate8x2Operands) DataType() Type {
	return IntType
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

func (op *Immediate8x3Operands) DataType() Type {
	return IntType
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

func (op *Immediate24Operands) DataType() Type {
	return IntType
}

func (op *Immediate24Operands) Int() int {
	return int(op.Val)
}

func (op *Immediate24Operands) String() string {
	return fmt.Sprintf("%v", op.Val)
}

func (op *Immediate24Operands) Unpack(istr *Instruction, script *Script, res *resource.Container) {
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

type Immediate16Operands struct {
	Val uint16
}

func (op *Immediate16Operands) DataType() Type {
	return IntType
}

func (op *Immediate16Operands) Int() int {
	return int(op.Val)
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

func (op *Immediate32Operands) DataType() Type {
	return IntType
}

func (op *Immediate32Operands) Int() int {
	return int(op.Val)
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

func (op *ImmediateF32Operands) DataType() Type {
	return FloatType
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
	InSize     uint8
	OutSize    uint8
	Native     Native64
	NativeStrs []string
}

func (op *CallNOperands) String() string {
	return fmt.Sprintf("%x %v %v <%v>", op.Native, op.InSize, op.OutSize, strings.Join(op.NativeStrs, ","))
}

func (op *CallNOperands) Unpack(istr *Instruction, script *Script, res *resource.Container) {
	var nativeIndex uint16
	var nativeOperand uint8
	res.Parse(&nativeOperand)
	res.ParseBigEndian(&nativeIndex)

	op.Native = script.NativeTable[nativeIndex]
	if nativeStrs, ok := script.NativeLookup(op.Native); ok {
		op.NativeStrs = nativeStrs
	} else {
		op.NativeStrs = []string{"unknown"}
	}

	op.InSize = (nativeOperand >> 2)
	op.OutSize = nativeOperand & 0x3
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
	NumArgs    uint8
	NumLocals  uint8
	Unknown2   uint8
	NameLength uint8
	Name       string
}

func (op *EnterOperands) String() string {
	return fmt.Sprintf("%v %v %v %v <%v>", op.NumArgs, op.NumLocals, op.Unknown2, op.NameLength, op.Name)
}

func (op *EnterOperands) Unpack(istr *Instruction, script *Script, res *resource.Container) {
	res.Parse(&op.NumArgs)
	res.Parse(&op.NumLocals)
	res.Parse(&op.Unknown2)
	res.Parse(&op.NameLength)
	if op.NameLength > 0 {
		res.Parse(&op.Name)
	} else {
		op.Name = fmt.Sprintf("anonymous_%x", istr.Address)
	}
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

func (op *ImplicitOperands) DataType() Type {
	return IntType
}

func (op *ImplicitOperands) String() string {
	return fmt.Sprintf("%v", op.Val)
}

func (op *ImplicitOperands) Unpack(istr *Instruction, script *Script, res *resource.Container) {
	op.Val = int(istr.Opcode) - op.offset
}

func (op *ImplicitOperands) Int() int {
	return op.Val
}

type ImplicitFOperands struct {
	offset int
	Val    float32
}

func (op *ImplicitFOperands) DataType() Type {
	return FloatType
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

func (op *StringOperands) DataType() Type {
	return StringType
}

func (op *StringOperands) String() string {
	return fmt.Sprintf("\"%v\"", op.Val)
}

func (op *StringOperands) Unpack(istr *Instruction, script *Script, res *resource.Container) {

}
