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
	OpSh2Add
	OpSh2AddPk
	OpStrCpy
	OpItoF
	OpFtoI
	OpItoS
	OpAppendStr
	OpAppendInt
)

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
	50:  OpExplode,
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
	64:  OpArrayGetP,
	65:  OpArrayGet,
	66:  OpArraySet,
	67:  OpPush,
	68:  OpAdd,
	69:  OpMul,
	70:  OpSh2Add,
	71:  OpNop,
	72:  OpSh2AddPk,
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
