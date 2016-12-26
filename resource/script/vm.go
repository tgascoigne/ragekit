/* Primitive stack emulation */
package script

import (
	//	"log"

	"github.com/tgascoigne/ragekit/util/stack"
)

type VM struct {
	script *Script
	Stack  stack.Stack
}

func (vm *VM) Init(script *Script) {
	vm.Stack.Allocate(4096)
	vm.script = script
}

func (vm *VM) PopInt() int {
	item := vm.Stack.Pop()
	//	log.Printf("pop %v\n", item.Value.(int))
	return item.Value.(int)
}

func (vm *VM) PushInt(val int) {
	//	log.Printf("push %v\n", val)
	vm.Stack.Push(&stack.Item{Value: val})
}

func (vm *VM) PopFloat() float32 {
	item := vm.Stack.Pop()
	return item.Value.(float32)
}

func (vm *VM) PushFloat(val float32) {
	vm.Stack.Push(&stack.Item{Value: val})
}

func (vm *VM) Execute(istr *Instruction) {
	defer func() {
		if r := recover(); r != nil {
			/* this occurs often, mostly because we dont emulate control flow */
		}
	}()

	/* do some basic stack emulation */
	if istr.Operation >= OpMathStart && istr.Operation <= OpMathEnd {
		vm.execMath(istr)
	}

	if istr.Operation >= OpCmpStart && istr.Operation <= OpCmpEnd {
		vm.execComparison(istr)
	}

	if istr.Operation >= OpStackStart && istr.Operation <= OpStackEnd {
		vm.execStackOp(istr)
	}

	if istr.Operation >= OpVarStart && istr.Operation <= OpVarEnd {
		vm.execVarOp(istr)
	}

	switch istr.Operation {
	case OpEnter:
		//		enterOperands := istr.Operands.(*EnterOperands)
		///		vm.Stack.Reserve(int(enterOperands.FrameSize))
	case OpPushStr:
		strOperands := istr.Operands.(*StringOperands)
		strIndex := vm.PopInt()
		strOperands.Val = vm.script.StringTableEntry(strIndex)
		vm.Stack.Push(&stack.Item{Value: strOperands.Val})
	}
}

func (vm *VM) execMath(istr *Instruction) {
	/* check if we're dealing with floating operands.. */
	var floatingOperands bool
	switch vm.Stack.Peek().Value.(type) {
	case float32:
		floatingOperands = true
	}

	if floatingOperands {
		switch istr.Operation {
		case OpAdd:
			vm.PushFloat(vm.PopFloat() + vm.PopFloat())
		case OpSub:
			vm.PushFloat(vm.PopFloat() - vm.PopFloat())
		case OpMul:
			vm.PushFloat(vm.PopFloat() * vm.PopFloat())
		case OpDiv:
			vm.PushFloat(vm.PopFloat() / vm.PopFloat())
		case OpNeg:
			vm.PushFloat(-vm.PopFloat())
		}
	} else {
		switch istr.Operation {
		case OpAdd:
			vm.PushInt(vm.PopInt() + vm.PopInt())
		case OpSub:
			vm.PushInt(vm.PopInt() - vm.PopInt())
		case OpMul:
			vm.PushInt(vm.PopInt() * vm.PopInt())
		case OpDiv:
			vm.PushInt(vm.PopInt() / vm.PopInt())
		case OpMod:
			vm.PushInt(vm.PopInt() % vm.PopInt())
		case OpNot:
			vm.PushInt(^vm.PopInt())
		case OpNeg:
			vm.PushInt(-vm.PopInt())
		case OpAnd:
			vm.PushInt(vm.PopInt() & vm.PopInt())
		case OpOr:
			vm.PushInt(vm.PopInt() | vm.PopInt())
		case OpXor:
			vm.PushInt(vm.PopInt() ^ vm.PopInt())
		}
	}
	/* todo: vector math */
}

func (vm *VM) execComparison(istr *Instruction) {
	if vm.Stack.Count() == 0 {
		return
	}

	/* check if we're dealing with floating operands.. */
	var floatingOperands bool
	switch vm.Stack.Peek().Value.(type) {
	case float32:
		floatingOperands = true
	}

	if floatingOperands {
		switch istr.Operation {
		case OpCmpEq:
			if vm.PopFloat() == vm.PopFloat() {
				vm.PushInt(1)
			} else {
				vm.PushInt(0)
			}
		case OpCmpNe:
			if vm.PopFloat() != vm.PopFloat() {
				vm.PushInt(1)
			} else {
				vm.PushInt(0)
			}
		case OpCmpGt:
			if vm.PopFloat() > vm.PopFloat() {
				vm.PushInt(1)
			} else {
				vm.PushInt(0)
			}
		case OpCmpGe:
			if vm.PopFloat() >= vm.PopFloat() {
				vm.PushInt(1)
			} else {
				vm.PushInt(0)
			}
		case OpCmpLt:
			if vm.PopFloat() < vm.PopFloat() {
				vm.PushInt(1)
			} else {
				vm.PushInt(0)
			}
		case OpCmpLe:
			if vm.PopFloat() <= vm.PopFloat() {
				vm.PushInt(1)
			} else {
				vm.PushInt(0)
			}
		}
	} else {
		switch istr.Operation {
		case OpCmpEq:
			if vm.PopInt() == vm.PopInt() {
				vm.PushInt(1)
			} else {
				vm.PushInt(0)
			}
		case OpCmpNe:
			if vm.PopInt() != vm.PopInt() {
				vm.PushInt(1)
			} else {
				vm.PushInt(0)
			}
		case OpCmpGt:
			if vm.PopInt() > vm.PopInt() {
				vm.PushInt(1)
			} else {
				vm.PushInt(0)
			}
		case OpCmpGe:
			if vm.PopInt() >= vm.PopInt() {
				vm.PushInt(1)
			} else {
				vm.PushInt(0)
			}
		case OpCmpLt:
			if vm.PopInt() < vm.PopInt() {
				vm.PushInt(1)
			} else {
				vm.PushInt(0)
			}
		case OpCmpLe:
			if vm.PopInt() <= vm.PopInt() {
				vm.PushInt(1)
			} else {
				vm.PushInt(0)
			}
		}
	}
}

func (vm *VM) execStackOp(istr *Instruction) {
	switch istr.Operation {
	case OpDup:
		val := vm.Stack.Peek()
		vm.Stack.Push(val)
	case OpDrop:
		vm.Stack.Pop()
	}

	var floatingOperands bool
	switch istr.Operands.(type) {
	case *ImmediateF32Operands, *ImplicitFOperands:
		floatingOperands = true
	}

	if floatingOperands {
		var operand float32
		switch op := istr.Operands.(type) {
		case *ImmediateF32Operands:
			operand = float32(op.Val)
		case *ImplicitFOperands:
			operand = float32(op.Val)
		}

		switch istr.Operation {
		case OpPush:
			vm.PushFloat(operand)
		}
	} else {
		var operand int
		switch op := istr.Operands.(type) {
		case *Immediate8Operands:
			operand = int(op.Val)
		case *Immediate16Operands:
			operand = int(op.Val)
		case *Immediate24Operands:
			operand = int(op.Val)
		case *Immediate32Operands:
			operand = int(op.Val)
		case *ImplicitOperands:
			operand = int(op.Val)
		}

		switch istr.Operation {
		case OpPush:
			vm.PushInt(operand)
		}
	}
}

func (vm *VM) execVarOp(istr *Instruction) {
	vm.PushInt(0)
	/* todo: all var ops
	OpExplode
	OpImplode
	OpArrayGetP
	OpArrayGet
	OpArraySet
	OpGetP
	OpSetP
	OpSetPPeek
	OpGetLocalP
	OpGetLocal
	OpSetLocal
	OpGetStaticP
	OpGetStatic
	OpSetStatic
	OpGetGlobalP
	OpGetGlobal
	OpSetGlobal
	*/
}
