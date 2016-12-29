package script

const (
	OpNop       uint8 = iota
	OpMathStart       /* used to test for math ops */
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpMod
	OpNot
	OpNeg
	OpAnd
	OpOr
	OpXor
	OpMathEnd  /* used to test for math ops */
	OpCmpStart /* used to test for cmp ops */
	OpCmpEq
	OpCmpNe
	OpCmpGt
	OpCmpGe
	OpCmpLt
	OpCmpLe
	OpCmpEnd     /* used to test for cmp ops */
	OpStackStart /* used to test for stack ops */
	OpPush
	OpDup
	OpDrop
	OpPushStr
	OpPushStrL
	OpPushStrN
	OpStackEnd  /* used to test for stack ops */
	OpFlowStart /* used to test for flow control ops */
	OpCallN
	OpEnter
	OpCall
	OpRet
	OpBranch
	OpBranchZ
	OpBranchNe
	OpBranchEq
	OpBranchGt
	OpBranchGe
	OpBranchLt
	OpBranchLe
	OpSwitch
	OpCatch
	OpThrow
	OpCallP
	OpFlowEnd  /* used to test for flow control ops */
	OpVarStart /* used to test for var ops */
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
	OpVarEnd /* used to test for var ops */
	OpFieldGetP
	OpFieldSet
	OpFieldGet
	OpStrCpy
	OpItoF
	OpFtoI
	OpItoS
	OpAppendStr
	OpAppendInt
)

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

var OpType = map[uint8]uint8{
	0:   OpNop,
	1:   OpAdd,
	2:   OpSub,
	3:   OpMul,
	4:   OpDiv,
	5:   OpMod,
	6:   OpNot,
	7:   OpNeg,
	8:   OpCmpEq,
	9:   OpCmpNe,
	10:  OpCmpGt,
	11:  OpCmpGe,
	12:  OpCmpLt,
	13:  OpCmpLe,
	14:  OpAdd,
	15:  OpSub,
	16:  OpMul,
	17:  OpDiv,
	18:  OpMod,
	19:  OpNeg,
	20:  OpCmpEq,
	21:  OpCmpNe,
	22:  OpCmpGt,
	23:  OpCmpGe,
	24:  OpCmpLt,
	25:  OpCmpLe,
	26:  OpAdd,
	27:  OpSub,
	28:  OpMul,
	29:  OpDiv,
	30:  OpNeg,
	31:  OpAnd,
	32:  OpOr,
	33:  OpXor,
	34:  OpItoF,
	35:  OpFtoI,
	36:  OpDup,
	37:  OpPush,
	38:  OpPush,
	39:  OpPush,
	40:  OpPush,
	41:  OpPush,
	42:  OpDup,
	43:  OpDrop,
	44:  OpCallN,
	45:  OpEnter,
	46:  OpRet,
	47:  OpGetP,
	48:  OpSetP,
	49:  OpSetPPeek,
	50:  OpExplode, //tostack
	51:  OpImplode,
	52:  OpArrayGetP,
	53:  OpArrayGet,
	54:  OpArraySet,
	55:  OpGetLocalP,
	56:  OpGetLocal,
	57:  OpSetLocal,
	58:  OpGetStaticP,
	59:  OpGetStatic,
	60:  OpSetStatic,
	61:  OpAdd,
	62:  OpMul,
	63:  OpArrayGetP,
	64:  OpFieldGetP,
	65:  OpFieldGet,
	66:  OpFieldSet,
	67:  OpPush,
	68:  OpAdd,
	69:  OpMul,
	70:  OpFieldGetP,
	71:  OpFieldGet,
	72:  OpFieldSet,
	73:  OpArrayGetP,
	74:  OpArrayGet,
	75:  OpArraySet,
	76:  OpGetLocalP,
	77:  OpGetLocal,
	78:  OpSetLocal,
	79:  OpGetStaticP,
	80:  OpGetStatic,
	81:  OpSetStatic,
	82:  OpGetGlobalP,
	83:  OpGetGlobal,
	84:  OpSetGlobal,
	85:  OpBranch,
	86:  OpBranchZ,
	87:  OpBranchNe,
	88:  OpBranchEq,
	89:  OpBranchGt,
	90:  OpBranchGe,
	91:  OpBranchLt,
	92:  OpBranchLe,
	93:  OpCall,
	94:  OpGetGlobalP,
	95:  OpGetGlobal,
	96:  OpSetGlobal,
	97:  OpPush,
	98:  OpSwitch,
	99:  OpPushStr,
	100: OpPushStrN,
	101: OpStrCpy,
	102: OpItoS,
	103: OpAppendStr,
	104: OpAppendInt,
	105: OpStrCpy,
	106: OpCatch,
	107: OpThrow,
	108: OpCallP,
	109: OpPush,
	110: OpPush,
	111: OpPush,
	112: OpPush,
	113: OpPush,
	114: OpPush,
	115: OpPush,
	116: OpPush,
	117: OpPush,
	118: OpPush,
	119: OpPush,
	120: OpPush,
	121: OpPush,
	122: OpPush,
	123: OpPush,
	124: OpPush,
	125: OpPush,
	126: OpPush,
}

var OpMnemonic = map[uint8]string{
	OpNop:        "nop",
	OpAdd:        "add",
	OpSub:        "sub",
	OpMul:        "mul",
	OpDiv:        "div",
	OpMod:        "mod",
	OpNot:        "not",
	OpNeg:        "neg",
	OpCmpEq:      "cmpeq",
	OpCmpNe:      "cmpne",
	OpCmpGt:      "cmpgt",
	OpCmpGe:      "cmpge",
	OpCmpLt:      "cmplt",
	OpCmpLe:      "cmple",
	OpAnd:        "and",
	OpOr:         "or",
	OpXor:        "xor",
	OpItoF:       "itof",
	OpFtoI:       "ftoi",
	OpPush:       "push",
	OpDup:        "dup",
	OpDrop:       "drop",
	OpCallN:      "calln",
	OpEnter:      "enter",
	OpRet:        "ret",
	OpGetP:       "getp",
	OpSetP:       "setp",
	OpSetPPeek:   "setpp",
	OpExplode:    "explode",
	OpImplode:    "implode",
	OpArrayGetP:  "getarrayp",
	OpArrayGet:   "getarray",
	OpArraySet:   "setarray",
	OpGetLocalP:  "getlocalp",
	OpGetLocal:   "getlocal",
	OpSetLocal:   "setlocal",
	OpGetStaticP: "getstaticp",
	OpGetStatic:  "getstatic",
	OpSetStatic:  "setstatic",
	OpGetGlobalP: "getglobalp",
	OpGetGlobal:  "getglobal",
	OpSetGlobal:  "setglobal",
	OpCall:       "call",
	OpBranch:     "b",
	OpBranchZ:    "bz",
	OpBranchNe:   "bne",
	OpBranchEq:   "be",
	OpBranchGt:   "bgt",
	OpBranchGe:   "bge",
	OpBranchLt:   "blt",
	OpBranchLe:   "ble",
	OpPushStr:    "pushstr",
	OpPushStrL:   "pushstrl",
	OpPushStrN:   "pushstrn",
	OpStrCpy:     "strcpy",
	OpItoS:       "itos",
	OpAppendStr:  "apps",
	OpAppendInt:  "appi",
	OpCatch:      "catch",
	OpThrow:      "throw",
	OpCallP:      "callp",
	OpSwitch:     "switch",
	OpFieldGetP:  "getfieldp",
	OpFieldGet:   "getfield",
	OpFieldSet:   "setfield",
}

var OpSuffix = map[uint8]string{
	1:   "i",
	2:   "i",
	3:   "i",
	4:   "i",
	5:   "i",
	6:   "i",
	7:   "i",
	8:   "i",
	9:   "i",
	10:  "i",
	11:  "i",
	12:  "i",
	13:  "i",
	14:  "f",
	15:  "f",
	16:  "f",
	17:  "f",
	18:  "f",
	19:  "f",
	20:  "f",
	21:  "f",
	22:  "f",
	23:  "f",
	24:  "f",
	25:  "f",
	26:  "v",
	27:  "v",
	28:  "v",
	29:  "v",
	30:  "v",
	37:  "b",
	38:  "b",
	39:  "b",
	40:  "i",
	41:  "f",
	61:  "imb",
	62:  "imb",
	63:  "b",
	64:  "b",
	65:  "b",
	66:  "b",
	67:  "s",
	68:  "ims",
	69:  "ims",
	70:  "s",
	71:  "s",
	72:  "s",
	73:  "s",
	74:  "s",
	75:  "s",
	76:  "s",
	77:  "s",
	78:  "s",
	79:  "s",
	80:  "s",
	81:  "s",
	82:  "s",
	83:  "s",
	84:  "s",
	94:  "t",
	95:  "t",
	96:  "t",
	97:  "t",
	109: "im",
	110: "im",
	111: "im",
	112: "im",
	113: "im",
	114: "im",
	115: "im",
	116: "im",
	117: "im",
	118: "imf",
	119: "imf",
	120: "imf",
	121: "imf",
	122: "imf",
	123: "imf",
	124: "imf",
	125: "imf",
	126: "imf",
}
