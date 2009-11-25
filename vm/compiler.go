import (
	"tr";
	"opcode";
	"internal";
)

// ast node
type ASTNode struct {
	type		TR_T;
	class		OBJ;
	ivars		*map[string] OBJ;
	ntype		int;
	args		[3]OBJ;
	line		size_t;
}

func newASTNode(vm *TrVM, type int, a, b, c OBJ, line size_t) OBJ {
	node = new(ASTNode);
	node.ntype = type;
	node.type = TR_T_Node;
	node.args[0] = a;
	node.args[1] = b;
	node.args[2] = c;
	node.line = line;
	return OBJ(node);
}

type Compiler struct {
	line		int;
	filename	OBJ;
	vm			*TrVM;
  	block		*Block;
  	reg			size_t;
  	node		OBJ;
}

// compiler

func newCompiler(vm *TrVM, filename *string) Compiler * {
	compiler := new(Compiler);
	compiler.line = 1;
	compiler.vm = vm;
	compiler.block = newBlock(compiler, 0);
	compiler.reg = 0;
	compiler.node = TR_NIL;
	compiler.filename = TrString_new2(vm, filename);
	return compiler;
}

/* code generation macros */
#define PUSH_OP(BLK,I) ({ \
  (BLK).code.Push(I); \
  BLK.code.Len()-1; \
})
#define PUSH_OP_A(BLK, OP, A)         PUSH_OP(BLK, CREATE_ABC(TR_OP_##OP, A, 0, 0))
#define PUSH_OP_AB(BLK, OP, A, B)     PUSH_OP(BLK, CREATE_ABC(TR_OP_##OP, A, B, 0))
#define PUSH_OP_ABC(BLK, OP, A, B, C) PUSH_OP(BLK, CREATE_ABC(TR_OP_##OP, A, B, C))
#define PUSH_OP_ABx(BLK, OP, A, Bx)   PUSH_OP(BLK, CREATE_ABx(TR_OP_##OP, A, Bx))
#define PUSH_OP_AsBx(BLK, OP, A, sBx) ({ \
  TrInst __i = CREATE_ABx(TR_OP_##OP, A, 0); SETARG_sBx(__i, sBx); \
  PUSH_OP(BLK, __i); \
})

func (self *ASTNode) compile_to_RK(vm *TrVM, c *Compiler, b *Block, reg int) int {
	int i;
  
 	// k value
	if self.ntype == NODE_VALUE {
    	return b.push_value(self.args[0]) | 0x100;
 
	// local
	} else if self.ntype == NODE_SEND && (i = b.find_local(self.args[1].args[0])) != -1 {
		return i;
  
 	// not a local, need to compile
	} else {
		if reg >= b.regc { b.regc = reg + 1; }
		self.compile(vm, c, b, reg);
		return reg;
	}
}

func (self *ASTNode) compile(vm *TrVM, c *Compiler, b *Block, reg int) OBJ {
	if !self { return TR_NIL; }
	start_reg := reg;
	if reg >= b.regc { b.regc = reg + 1; }
	b.line = self.line;
	// TODO this shit is very repetitive, need to refactor
	switch (self.ntype) {
		case NODE_ROOT, NODE_BLOCK:
			for node := range self.args[0].Iter() {
				nlocal := b.locals.Len();
				if reg >= b.regc { b.regc = reg + 1; }
				node.compile(vm, c, b, reg);
				reg += b.locals.Len() - nlocal;
				if reg >= b.regc { b.regc = reg + 1; }
			}

		case NODE_VALUE:
			i := b.push_value(self.args[0]);
			PUSH_OP_ABx(b, LOADK, reg, i);

		case NODE_STRING: {
			i := b.push_string(TR_STR_PTR(self.args[0]));
			PUSH_OP_ABx(b, STRING, reg, i);

		case NODE_ARRAY:
			size := 0;
			if self.args[0] {
				size = self.args[0].kv.Len();
				// compile args
				index := 0;
				for node := range self.args[0].Iter() {
					nlocal := b.locals.Len();
					new_reg := reg + index + 1;
					if new_reg >= b.regc { b.regc = new_reg + 1; }
					node.compile(vm, c, b, new_reg);
					reg += b.locals.Len() - nlocal;
					if reg >= b.regc { b.regc = reg + 1; }
					index++;
				}
				if start_reg != reg { tr_raise(SyntaxError, "Can't create local variable inside Array") }
			}
			PUSH_OP_AB(b, NEWARRAY, reg, size);

		case NODE_HASH:
			size := 0;
			if self.args[0] {
				size = self.args[0].kv.Len()
				// compile args
				index := 0;
				for node := range self.args[0].Iter() {
					nlocal := b.locals.Len();
					new_reg := reg + index + 1;
					if new_reg >= b.regc { b.regc = new_reg + 1; }
					node.compile(vm, c, b, new_reg);
					reg += b.locals.Len() - nlocal;
					if reg >= b.regc { b.regc = reg + 1; }
					index++;
				}
				if start_reg != reg { tr_raise(SyntaxError, "Can't create local variable inside Hash") }
			}
			PUSH_OP_AB(b, NEWHASH, reg, size/2);

		case NODE_RANGE:
			if reg >= b.regc { b.regc = reg + 1; }
			self.args[0].compile(vm, c, b, reg);
			next_reg := reg + 1;
			if next_reg >= b.regc { b.regc = next_reg + 1; }
			self.args[1].compile(vm, c, b, next_reg);
			if next_reg >= b.regc { b.regc = next_reg + 1; }
			if start_reg != reg { tr_raise(SyntaxError, "Can't create local variable inside Range") }
			PUSH_OP_ABC(b, NEWRANGE, reg, next_reg, self.args[2]);

		case NODE_ASSIGN:
			name := self.args[0];
			if reg >= b.regc { b.regc = reg + 1; }
			self.args[1].compile(vm, c, b, reg);
			if (b.find_upval_in_scope(name) != -1) {
				// upval
				PUSH_OP_AB(b, SETUPVAL, reg, b.push_upval(name));
			} else {
				// local
				i = b.push_local(name);
				last_inst := b.code.Last();
				switch (GET_OPCODE(last_inst)) {
					case TR_OP_ADD, TR_OP_SUB, TR_OP_LT, TR_OP_NEG, TR_OP_NOT:
						Those instructions can load direcly into a local
						SETARG_A(last_inst, i);

					default:
						if (i != reg) PUSH_OP_AB(b, MOVE, i, reg);
				}
			}

		case NODE_SETIVAR:
			if reg >= b.regc { b.regc = reg + 1; }
			self.args[1].compile(vm, c, b, reg);
			PUSH_OP_ABx(b, SETIVAR, reg, b.push_value(self.args[0]));

		case NODE_GETIVAR:
			PUSH_OP_ABx(b, GETIVAR, reg, b.push_value(self.args[0]));

		case NODE_SETCVAR:
			if reg >= b.regc { b.regc = reg + 1; }
			self.args[1].compile(vm, c, b, reg);
			PUSH_OP_ABx(b, SETCVAR, reg, b.push_value(self.args[0]));

		case NODE_GETCVAR:
			PUSH_OP_ABx(b, GETCVAR, reg, b.push_value(self.args[0]));

		case NODE_SETGLOBAL:
			if reg >= b.regc { b.regc = reg + 1; }
			self.args[1].compile(vm, c, b, reg);
			PUSH_OP_ABx(b, SETGLOBAL, reg, b.push_value(self.args[0]));

		case NODE_GETGLOBAL:
			PUSH_OP_ABx(b, GETGLOBAL, reg, b.push_value(self.args[0]));

		case NODE_SEND:
			// can also be a variable access
			msg := self.args[1];
			name = msg.args[0];
			assert(msg.ntype == NODE_MSG);
			i int;
			// local
			if (i = b.find_local(name)) != -1 {
				if reg != i { PUSH_OP_AB(b, MOVE, reg, i); }
        
			// upval
			} else if b.find_upval_in_scope(name) != -1 {
				i := b.push_upval(name);
				PUSH_OP_AB(b, GETUPVAL, reg, i);

			// method call
			} else {
				// receiver
				if self.args[0] {
					if reg >= b.regc { b.regc = reg + 1; }
					self.args[0].compile(vm, c, b, reg);
				} else {
					PUSH_OP_A(b, SELF, reg);
				}
				i = b.push_value(name);
				// args
				argc := 0;
				if msg.args[1] {
					argc = msg.args[1].kv.Len() << 1;
					index := 0;
					for argument := range msg.args[1].Iter() {
						nlocal := b.locals.Len();
						new_reg := reg + index + 2;
						if new_reg >= b.regc { b.regc = new_reg + 1; }
						argument.args[0].compile(vm, c, b, new_reg);
						reg += b.locals.Len() - nlocal;
						if argument.args[1] { argc |= 1 }		// splat
						index++;
					}
					if start_reg != reg { tr_raise(SyntaxError, "Can't create local variable inside arguments") }
				}

				// block
				blki := 0;
				blk := nil;
				if (self.args[2]) {
					blk := newBlock(c, b);
					blkn := self.args[2];
					blki = b.blocks.Len() + 1;
					blk.argc = 0;
					if blkn.args[1] {
						blk.argc = blkn.args[1].kv.Len();
						// add parameters as locals in block context
						for parameter := range blk.args[1].Iter() {
							blk.push_local(parameter.args[0])
						}
					}
					b.blocks.Push(blk);
					blk_reg := blk.locals.Len();
					if blk_reg >= b.regc { b.regc = blk_reg + 1; }
					blkn.compile(vm, c, blk, blk_reg);
					PUSH_OP_A(blk, RETURN, blk_reg);
				}
				PUSH_OP_A(b, BOING, 0);
				PUSH_OP_ABx(b, LOOKUP, reg, i);
				PUSH_OP_ABC(b, CALL, reg, argc, blki);
        
				// if passed block has upvalues generate one pseudo-instructions for each (A reg is ignored).
				if blk && blk.upvals.Len() {
				for j := 0; j < blk.upvals.Len(); ++j {
					upval_name := blk.upvals.At(j);
					vali := b.find_local(upval_name);
					if vali != -1 {
						PUSH_OP_AB(b, MOVE, 0, vali);
					} else {
						PUSH_OP_AB(b, GETUPVAL, 0, b.find_upval(upval_name));
					}
				}
			}

		case NODE_IF, NODE_UNLESS:
			// condition
			if reg >= b.regc { b.regc = reg + 1; }
			self.args[0].compile(vm, c, b, reg);

			if self.ntype == NODE_IF {
				jmp := PUSH_OP_A(b, JMPUNLESS, reg);
			} else {
				jmp := PUSH_OP_A(b, JMPIF, reg);
			}
 
			// body
			for node := range self.args[1].Iter() {
				nlocal := b.locals.Len();
				if reg >= b.regc { b.regc = reg + 1; }
				node.compile(vm, c, b, reg);
				reg += b.locals.Len() - nlocal;
				if reg >= b.regc { b.regc = reg + 1; }
			}
			SETARG_sBx(b.code.At(jmp), b.code.Len() - jmp);
			// else body
			jmp = PUSH_OP_A(b, JMP, reg);
			if self.args[2] {
				for node := range self.args[2].Iter() {
					nlocal := b.locals.Len();
					if reg >= b.regc { b.regc = reg + 1; }
					node.compile(vm, c, b, reg);
					reg += b.locals.Len() - nlocal;
					if reg >= b.regc { b.regc = reg + 1; }
				}
			} else {
				// if condition fail and not else block nil is returned
				PUSH_OP_A(b, NIL, reg);
			}
			SETARG_sBx(b.code.At(jmp), b.code.Len() - jmp - 1);

		case NODE_WHILE, NODE_UNTIL:
			jmp_beg := b.code.Len();
			// condition
			if reg >= b.regc { b.regc = reg + 1; }
			self.args[0].compile(vm, c, b, reg);

			if self.ntype == NODE_WHILE {
				PUSH_OP_ABx(b, JMPUNLESS, reg, 0);
			} else {
				PUSH_OP_ABx(b, JMPIF, reg, 0);
			}
			jmp_end := b.code.Len();
			// body
			for node := range self.args[1].Iter() {
				nlocal := b.locals.Len();
				if reg >= b.regc { b.regc = reg + 1; }
				node.compile(vm, c, b, reg);
				reg += b.locals.Len() - nlocal;
				if reg >= b.regc { b.regc = reg + 1; }
			}
			SETARG_sBx(b.code.At(jmp_end - 1), b.code.Len() - jmp_end + 1);
			PUSH_OP_AsBx(b, JMP, 0, 0-(b.code.Len() - jmp_beg) - 1);

		case NODE_AND, NODE_OR:
			// receiver
			if reg >= b.regc { b.regc = reg + 1; }
			self.args[0].compile(vm, c, b, reg);
// Appears to be erroneously compiling the same node twice
//			self.args[0].compile(vm, c, b, reg);
			if self.ntype == NODE_AND {
				jmp := PUSH_OP_A(b, JMPUNLESS, reg);
			} else {
				jmp := PUSH_OP_A(b, JMPIF, reg);
			}
			// arg
			if reg >= b.regc { b.regc = reg + 1; }
			self.args[1].compile(vm, c, b, reg);
			SETARG_sBx(b.code.At(jmp), b.code.Len() - jmp - 1);

		case NODE_BOOL:
			PUSH_OP_AB(b, BOOL, reg, self.args[0]);

		case NODE_NIL:
			PUSH_OP_A(b, NIL, reg);

		case NODE_SELF:
			PUSH_OP_A(b, SELF, reg);

		case NODE_RETURN:
			if self.args[0] {
				if reg >= b.regc { b.regc = reg + 1; }
				self.args[0].compile(vm, c, b, reg);
			}
			if b.parent {
				PUSH_OP_AB(b, THROW, TR_THROW_RETURN, reg);
			} else {
				PUSH_OP_A(b, RETURN, reg);
			}

		case NODE_BREAK:
			PUSH_OP_A(b, THROW, TR_THROW_BREAK);

		case NODE_YIELD: {
			argc := 0;
			if self.args[0] {
				argc = self.args[0].kv.Len();
				index := 0;
				for node := range self.args[0].Iter() {
					nlocal := b.locals.Len();
					new_reg := reg + index + 1;
					if new_reg >= b.regc { b.regc = new_reg + 1; }
					node.compile(vm, c, b, new_reg);
					reg += b.locals.Len() - nlocal;
					if reg >= b.regc { b.regc = reg + 1; }
					index++;
				}
				if start_reg != reg { tr_raise(SyntaxError, "Can't create local variable inside yield") }
			}
			PUSH_OP_AB(b, YIELD, reg, argc);

		case NODE_DEF: {
			method := self.args[0];
			assert(method.ntype == NODE_METHOD);
			blk := newBlock(c, 0);
			blki := b.blocks.Len();
			blk_reg := 0;
			b.blocks.Push(blk);
			if self.args[1] {
				// add parameters as locals in method context
				blk.argc = self.args[1].kv.Len();
				for parameter := range self.args[1].Iter() {
					blk.push_local(parameter.args[0]);
					if parameter.args[1] { blk.arg_splat = 1; }
					// compile default expression and store location in defaults table for later jump when executing
					if parameter.args[2] {
						if blk_reg >= b.regc { b.regc = blk_reg + 1; }
						parameter.args[2].compile(vm, c, blk, blk_reg);
						blk.defaults.push(blk.code.Len());
					}
					blk_reg++;
				}
			}
 			// compile body of method
			for node := range self.args[2].Iter() {
				nlocal := blk.locals.Len();
				if blk_reg >= b.regc { b.regc = blk_reg + 1; }
				node.compile(vm, c, blk, blk_reg);
				blk_reg += blk.locals.Len() - nlocal;
				if blk_reg >= b.regc { b.regc = blk_reg + 1; }
			}

			PUSH_OP_A(blk, RETURN, blk_reg);
			if method.args[0] {
				// metaclass def
				if reg >= b.regc { b.regc = reg + 1; }
				method.args[0].compile(vm, c, b, reg);
				PUSH_OP_ABx(b, METADEF, blki, b.push_value(method.args[1]));
				PUSH_OP_A(b, BOING, reg);
			} else {
				PUSH_OP_ABx(b, DEF, blki, b.push_value(method.args[1]));
			}

		case NODE_CLASS, NODE_MODULE:
			blk := newBlock(c, 0);
			blki := b.blocks.Len();
			b.blocks.Push(blk);
			reg = 0;

			// compile body of class
			for node := range self.args[2].Iter() {
				nlocal := blk.locals.Len();
				if reg >= b.regc { b.regc = reg + 1; }
				node.compile(vm, c, blk, reg);
				reg += blk.locals.Len() - nlocal;
				if reg >= b.regc { b.regc = reg + 1; }
			}
			PUSH_OP_A(blk, RETURN, reg);
			if (self.ntype == NODE_CLASS) {
				// superclass
				if self.args[1] {
					PUSH_OP_ABx(b, GETCONST, reg, b.push_value(self.args[1]));
				} else {
					PUSH_OP_A(b, NIL, reg);
				}
				PUSH_OP_ABx(b, CLASS, blki, b.push_value(self.args[0]));
				PUSH_OP_A(b, BOING, reg);
			} else {
				PUSH_OP_ABx(b, MODULE, blki, b.push_value(self.args[0]));
			}

		case NODE_CONST:
			PUSH_OP_ABx(b, GETCONST, reg, b.push_value(self.args[0]));

		case NODE_SETCONST:
			if reg >= b.regc { b.regc = reg + 1; }
			self.args[1].compile(vm, c, b, reg);
			PUSH_OP_ABx(b, SETCONST, reg, b.push_value(self.args[0]));

		case NODE_ADD, NODE_SUB, NODE_LT:
			rcv := self.args[0].compile_to_RK(vm, c, b, reg);
			arg := self.args[1].compile_to_RK(vm, c, b, reg + 1);
			if (reg + 1) >= b.regc { b.regc = reg + 2; }
			switch self.ntype {
				case NODE_ADD:	PUSH_OP_ABC(b, ADD, reg, rcv, arg);
				case NODE_SUB:	PUSH_OP_ABC(b, SUB, reg, rcv, arg);
				case NODE_LT:	PUSH_OP_ABC(b, LT, reg, rcv, arg);
				default:		assert(0);
			}

		case NODE_NEG, NODE_NOT:
			rcv := self.args[0].compile_to_RK(vm, c, b, reg);
			switch self.ntype {
				case NODE_NEG:	PUSH_OP_AB(b, NEG, reg, rcv);
				case NODE_NOT:	PUSH_OP_AB(b, NOT, reg, rcv);
				default:		assert(0);
			}

		default:
			printf("Compiler: unknown node type: %d in %s:%lu\n", self.ntype, TR_STR_PTR(b.filename), b.line);
			if vm.debug { assert(0); }
	}
	return TR_NIL;
}

func (self *Compiler) compile {
	b := self.block;
	b.filename = self.filename;
	self.node.compile(self.vm, c, b, 0);
	PUSH_OP_A(b, RETURN, 0);
}