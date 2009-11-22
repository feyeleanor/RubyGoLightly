const SIZE_C = 9;
const SIZE_B = 9;
const SIZE_Bx = SIZE_C + SIZE_B;
const SIZE_A = 8;
const SIZE_OP = 6;

const POS_OP = 0;
const POS_A = POS_OP + SIZE_OP;
const POS_C = POS_A + SIZE_A;
const POS_B = POS_C + SIZE_C;
const POS_Bx = POS_C;

const MAXARG_Bx = (1 << SIZE_Bx) - 1;
const MAXARG_sBx = MAXARG_Bx >> 1;			// `sBx' is signed

// creates a mask with 'n' 1 bits at position `p'
#define MASK1(n,p)  ((~((~(TrInst)0)<<n))<<p)

// creates a mask with `n' 0 bits at position `p'
#define MASK0(n,p)  (~MASK1(n,p))

// the following macros help to manipulate instructions (TrInst)

#define GET_OPCODE(i) ((int) ((i)>>POS_OP) & MASK1(SIZE_OP,0)))
#define SET_OPCODE(i,o) ((i) = (((i)&MASK0(SIZE_OP,POS_OP)) | (((TrInst) o)<<POS_OP)&MASK1(SIZE_OP,POS_OP))))

#define GETARG_A(i) ((int) ((i)>>POS_A) & MASK1(SIZE_A,0)))
#define SETARG_A(i,u) ((i) = (((i)&MASK0(SIZE_A,POS_A)) | (((TrInst) u)<<POS_A)&MASK1(SIZE_A,POS_A))))

#define GETARG_B(i) ((int) ((i)>>POS_B) & MASK1(SIZE_B,0)))
#define SETARG_B(i,b) ((i) = (((i)&MASK0(SIZE_B,POS_B)) | (((TrInst) b)<<POS_B)&MASK1(SIZE_B,POS_B))))

#define GETARG_C(i) ((int) ((i)>>POS_C) & MASK1(SIZE_C,0)))
#define SETARG_C(i,b) ((i) = (((i)&MASK0(SIZE_C,POS_C)) | (((TrInst) b)<<POS_C)&MASK1(SIZE_C,POS_C))))

#define GETARG_Bx(i)  ((int) ((i)>>POS_Bx) & MASK1(SIZE_Bx,0)))
#define SETARG_Bx(i,b)  ((i) = (((i)&MASK0(SIZE_Bx,POS_Bx)) | (((TrInst) b)<<POS_Bx)&MASK1(SIZE_Bx,POS_Bx))))

#define GETARG_sBx(i) ((int) GETARG_Bx(i)-MAXARG_sBx)
#define SETARG_sBx(i,b) SETARG_Bx((i), ((unsigned int) (b) + MAXARG_sBx))

#define CREATE_ABC(o,a,b,c) (((TrInst) o)<<POS_OP) | ((TrInst) a)<<POS_A) | ((TrInst) b)<<POS_B) | ((TrInst) c)<<POS_C))
#define CREATE_ABx(o,a,bc)  (((TrInst) o)<<POS_OP) | ((TrInst) a)<<POS_A) | ((TrInst) bc)<<POS_Bx))

/*
== TinyRb opcodes.
Format of one instruction: OPCODE A B C
Bx    -- unsigned value of BC
sBx   -- signed value of BC
R[A]  -- Value of register which index is stored in A of the current instruction.
R[nA] -- Value of the register A in the next instruction (instruction will be ignored).
K[A]  -- Value which index is stored in A of the current instruction.
RK[A] -- Register A or a constant index
*/
const (
  // opname					operands	description
  TR_OP_BOING = iota;		//          do nothing with elegance and frivolity
  TR_OP_MOVE;       		// A B      R[A] = R[B]
  TR_OP_LOADK;      		// A Bx     R[A] = K[Bx]
  TR_OP_STRING;     		// A Bx     R[A] = strings[Bx]
  TR_OP_BOOL;       		// A B      R[A] = B + 1
  TR_OP_NIL;        		// A        R[A] = nil
  TR_OP_SELF;       		// A        put self in R[A]
  TR_OP_LOOKUP;     		// A Bx     R[A+1] = lookup method K[Bx] on R[A] and store
  TR_OP_CACHE;      		// A B C    if sites[C] matches R[A].type, jmp +B and next call will be on sites[C]
  TR_OP_CALL;       		/* A B C    call last looked up method on R[A] with B>>1 args starting at R[A+2],
                                		if B & 1, splat last arg,
                                		if C > 0 pass block[C-1] */
  TR_OP_JMP;        		//   sBx    jump sBx instructions
  TR_OP_JMPIF;      		// A sBx    jump sBx instructions if R[A]
  TR_OP_JMPUNLESS;  		// A sBx    jump sBx instructions unless R[A]
  TR_OP_RETURN;     		// A        return R[A] (can be non local)
  TR_OP_THROW;      		// A B      throw type=A value=R[B]
  TR_OP_SETUPVAL;   		// A B      upvals[B] = R[A]
  TR_OP_GETUPVAL;   		// A B      R[A] = upvals[B]
  TR_OP_DEF;        		// A Bx     define method k[Bx] on self w/ blocks[A]
  TR_OP_METADEF;    		// A Bx     define method k[Bx] on R[nA] w/ blocks[A]
  TR_OP_GETCONST;   		// A Bx     R[A] = Consts[k[Bx]]
  TR_OP_SETCONST;   		// A Bx     Consts[k[Bx]] = R[A]
  TR_OP_CLASS;      		// A Bx     define class k[Bx] on self w/ blocks[A] and superclass R[nA]
  TR_OP_MODULE;     		// A Bx     define module k[Bx] on self w/ blocks[A]
  TR_OP_NEWARRAY;   		// A B      R[A] = Array.new(R[A+1]..R[A+1+B])
  TR_OP_NEWHASH;    		// A B      R[A] = Hash.new(R[A+1] => R[A+2] .. R[A+1+B*2] => R[A+2+B*2])
  TR_OP_YIELD;      		// A B      R[A] = passed_block.call(R[A+1]..R[A+1+B])
  TR_OP_GETIVAR;    		// A Bx     R[A] = self.ivars[k[Bx]]
  TR_OP_SETIVAR;    		// A Bx     self.ivars[k[Bx]] = R[A]
  TR_OP_GETCVAR;    		// A Bx     R[A] = class.ivars[k[Bx]]
  TR_OP_SETCVAR;    		// A Bx     class.ivars[k[Bx]] = R[A]
  TR_OP_GETGLOBAL;  		// A Bx     R[A] = globals[k[Bx]]
  TR_OP_SETGLOBAL;  		// A Bx     globals[k[Bx]] = R[A]
  TR_OP_NEWRANGE;   		// A B C    R[A] = Range.new(start:R[A], end:R[B], exclusive:C)
  TR_OP_ADD;        		// A B C    R[A] = RK[B] + RK[C]
  TR_OP_SUB;        		// A B C    R[A] = RK[B] - RK[C]
  TR_OP_LT;         		// A B C    R[A] = RK[B] < RK[C]
  TR_OP_NEG;        		// A B      R[A] = -RK[B]
  TR_OP_NOT;        		// A B      R[A] = !RK[B]
  TR_OP_SUPER;    			// TODO
)

const OPCODE_NAMES = []string {
	"boing",		"move",		"loadk",	"string",	"bool",			"nil",		"self",			"lookup",
	"cache",		"call",		"jmp",		"jmpif",	"jmpunless",	"return",	"throw",		"setupval",
	"getupval",		"def",		"metadef",	"getconst",	"setconst",		"class",	"module",		"newarray",
	"newhash",		"yield",	"getivar",	"setivar",	"getcvar",		"setcvar",	"getglobal",	"setglobal",
	"newrange",		"add",		"sub",		"lt",		"neg",			"not",		"super"
}