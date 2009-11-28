import (
	"tr";
	"opcode";
)

// ast node
type ASTNode struct {
	type		TR_T;
	class		*RubyObject;
	ivars		map[string] RubyObject;
	ntype		int;
	args		[3]RubyObject;
	line		size_t;
}

// types of nodes in the AST built by the parser
const (
	NODE_ROOT = iota;
	NODE_BLOCK;
	NODE_VALUE;
	NODE_STRING;
	NODE_ASSIGN;
	NODE_ARG;
	NODE_SEND;
	NODE_MSG;
	NODE_IF;
	NODE_UNLESS;
	NODE_AND;
	NODE_OR;
	NODE_WHILE;
	NODE_UNTIL;
	NODE_BOOL;
	NODE_NIL;
	NODE_SELF;
	NODE_LEAVE;
	NODE_RETURN;
	NODE_BREAK;
	NODE_YIELD;
	NODE_DEF;
	NODE_METHOD;
	NODE_PARAM;
	NODE_CLASS;
	NODE_MODULE;
	NODE_CONST;
	NODE_SETCONST;
	NODE_ARRAY;
	NODE_HASH;
	NODE_RANGE;
	NODE_GETIVAR;
	NODE_SETIVAR;
	NODE_GETCVAR;
	NODE_SETCVAR;
	NODE_GETGLOBAL;
	NODE_SETGLOBAL;
	NODE_ADD;
	NODE_SUB;
	NODE_LT;
	NODE_NEG;
	NODE_NOT;
)

func newASTNode(vm *RubyVM, type int, a, b, c *RubyObject, line size_t) RubyObject {
	return ASTNode{ntype: type, type: TR_T_NODE, args: {a, b, c}, line: line}
}

type Compiler struct {
	line		int;
	filename	*RubyObject;
	vm			*RubyVM;
  	block		*Block;
  	reg			size_t;
  	node		*RubyObject;
}

// compiler

func newCompiler(vm *RubyVM, filename *string) Compiler * {
	compiler := new(Compiler);
	compiler.line = 1;
	compiler.vm = vm;
	compiler.block = compiler.newBlock(nil);
	compiler.reg = 0;
	compiler.node = TR_NIL;
	compiler.filename = TrString_new2(vm, filename);
	return compiler;
}

func (self *ASTNode) compile_to_RK(vm *RubyVM, c *Compiler, b *Block, reg int) int {
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

func (self *ASTNode) compile(vm *RubyVM, c *Compiler, b *Block, reg int) RubyObject {
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
			b.code.Push(newExtendedOP(TR_OP_LOADK, reg, b.push_value(self.args[0])));

		case NODE_STRING: {
			if !self.args[0].(String) && !self.args[0].(Symbol) {
				vm.throw_reason = TR_THROW_EXCEPTION;
				vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + self.args[0]));
				return TR_UNDEF;
			}
			b.code.Push(newExtendedOP(TR_OP_STRING, reg, b.push_string(self.args[0].ptr));

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
				if start_reg != reg {
					vm.throw_reason = TR_THROW_EXCEPTION;
					vm.throw_value = TrException_new(vm, vm.cSyntaxError, tr_sprintf(vm, "Can't create local variable inside Array"));
					return TR_UNDEF;
				}
			}
			b.code.Push(MachineOp{OpCode: TR_OP_NEWARRAY, A: reg, B: size});

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
				if start_reg != reg {
					vm.throw_reason = TR_THROW_EXCEPTION;
					vm.throw_value = TrException_new(vm, vm.cSyntaxError, tr_sprintf(vm, "Can't create local variable inside Hash"));
					return TR_UNDEF;
				}
			}
			b.code.Push(MachineOp{OpCode: TR_OP_NEWHASH, A: reg, B: size / 2});

		case NODE_RANGE:
			if reg >= b.regc { b.regc = reg + 1; }
			self.args[0].compile(vm, c, b, reg);
			next_reg := reg + 1;
			if next_reg >= b.regc { b.regc = next_reg + 1; }
			self.args[1].compile(vm, c, b, next_reg);
			if next_reg >= b.regc { b.regc = next_reg + 1; }
			if start_reg != reg {
				vm.throw_reason = TR_THROW_EXCEPTION;
				vm.throw_value = TrException_new(vm, vm.cSyntaxError, tr_sprintf(vm, "Can't create local variable inside Range"));
				return TR_UNDEF;
			}
			b.code.Push(MachineOp{OpCode: TR_OP_NEWRANGE, A: reg, B: next_reg, C: self.args[2]});

		case NODE_ASSIGN:
			name := self.args[0];
			if reg >= b.regc { b.regc = reg + 1; }
			self.args[1].compile(vm, c, b, reg);
			if (b.find_upval_in_scope(name) != -1) {
				// upval
				b.code.Push(MachineOp{OpCode: TR_OP_SETUPVAL, A: reg, B: b.push_upval(name)});

			} else {
				// local
				i = b.push_local(name);
				last_inst := b.code.Last();
				switch (last_inst.OpCode) {
					case TR_OP_ADD, TR_OP_SUB, TR_OP_LT, TR_OP_NEG, TR_OP_NOT:
						// Those instructions can load direcly into a local
						SETARG_A(last_inst, i);

					default:
						if i != reg { b.code.Push(MachineOp{OpCode: TR_OP_MOVE, A: i, B: reg}); }
				}
			}

		case NODE_SETIVAR:
			if reg >= b.regc { b.regc = reg + 1; }
			self.args[1].compile(vm, c, b, reg);
			b.code.Push(newExtendedOP(TR_OP_SETIVAR, reg, b.push_value(self.args[0])));

		case NODE_GETIVAR:
			b.code.Push(newExtendedOP(TR_OP_GETIVAR, reg, b.push_value(self.args[0])));

		case NODE_SETCVAR:
			if reg >= b.regc { b.regc = reg + 1; }
			self.args[1].compile(vm, c, b, reg);
			b.code.Push(newExtendedOP(TR_OP_SETCVAR, reg, b.push_value(self.args[0])));

		case NODE_GETCVAR:
			b.code.Push(newExtendedOP(TR_OP_GETCVAR, reg, b.push_value(self.args[0])));

		case NODE_SETGLOBAL:
			if reg >= b.regc { b.regc = reg + 1; }
			self.args[1].compile(vm, c, b, reg);
			b.code.Push(newExtendedOP(TR_OP_SETGLOBAL, reg, b.push_value(self.args[0])));

		case NODE_GETGLOBAL:
			b.code.Push(newExtendedOP(TR_OP_GETGLOBAL, reg, b.push_value(self.args[0])));

		case NODE_SEND:
			// can also be a variable access
			msg := self.args[1];
			name = msg.args[0];
			assert(msg.ntype == NODE_MSG);
			// local
			if (i := b.find_local(name)) != -1 {
				if reg != i { b.code.Push(MachineOp{OpCode: TR_OP_MOVE, A: reg, B: i}); }

			// upval
			} else if b.find_upval_in_scope(name) != -1 {
				b.code.Push(MachineOp{OpCode: TR_OP_GETUPVAL, A: reg, B: b.push_upval(name)});

			// method call
			} else {
				// receiver
				if self.args[0] {
					if reg >= b.regc { b.regc = reg + 1; }
					self.args[0].compile(vm, c, b, reg);
				} else {
					b.code.Push(MachineOp{OpCode: TR_OP_SELF, A: reg});
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
					if start_reg != reg {
						vm.throw_reason = TR_THROW_EXCEPTION;
						vm.throw_value = TrException_new(vm, vm.cSyntaxError, tr_sprintf(vm, "Can't create local variable inside arguments"));
						return TR_UNDEF;
					}
				}

				// block
				blki := 0;
				blk := nil;
				if (self.args[2]) {
					blk := c.newBlock(b);
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
					blk.code.Push(MachineOp{OpCode: TR_OP_RETURN, A: blk_reg});
				}
				b.code.Push(MachineOp{OpCode: TR_OP_BOING});
				b.code.Push(newExtendedOP(TR_OP_LOOKUP, reg, i));
				b.code.Push(MachineOp{OpCode: TR_OP_CALL, A: reg, B: argc, C: blki});

				// if passed block has upvalues generate one pseudo-instructions for each (A reg is ignored).
				if blk && blk.upvals.Len() {
				for j := 0; j < blk.upvals.Len(); ++j {
					upval_name := blk.upvals.At(j);
					vali := b.find_local(upval_name);
					if vali != -1 {
						b.code.Push(MachineOp{OpCode: TR_OP_MOVE, B: vali});
					} else {
						b.code.Push(MachineOp{OpCode: TR_OP_GETUPVAL, B: b.find_upval(upval_name)});
					}
				}
			}

		case NODE_IF, NODE_UNLESS:
			// condition
			if reg >= b.regc { b.regc = reg + 1; }
			self.args[0].compile(vm, c, b, reg);

			if self.ntype == NODE_IF {
				b.code.Push(MachineOp{OpCode: TR_OP_JMPUNLESS, A: reg});
			} else {
				b.code.Push(MachineOp{OpCode: TR_OP_JMPIF, A: reg});
			}
			jmp := b.code.Len() - 1;
 
			// body
			for node := range self.args[1].Iter() {
				nlocal := b.locals.Len();
				if reg >= b.regc { b.regc = reg + 1; }
				node.compile(vm, c, b, reg);
				reg += b.locals.Len() - nlocal;
				if reg >= b.regc { b.regc = reg + 1; }
			}
			b.code.At(jmp).SetxBx(b.code.Len() - jmp);
			// else body
			b.code.Push(MachineOp{OpCode: TR_OP_JMP, A: reg});
			jmp := b.code.Len() - 1;

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
				b.code.Push(MachineOp{OpCode: TR_OP_NIL, A: reg});
			}
			b.code.At(jmp).Set_sBx(b.code.Len() - jmp - 1);

		case NODE_WHILE, NODE_UNTIL:
			jmp_beg := b.code.Len();
			// condition
			if reg >= b.regc { b.regc = reg + 1; }
			self.args[0].compile(vm, c, b, reg);

			if self.ntype == NODE_WHILE {
				b.code.Push(newExtendedOP(TR_OP_JMPUNLESS, reg, 0));
			} else {
				b.code.Push(newExtendedOP(TR_OP_JMPIF, reg, 0));
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
			b.code.At(jmp_end - 1).Set_sBx(b.code.Len() - jmp_end + 1);
		  	i := newExtendedOP(TR_OP_JMP, 0, 0);
		  	i.SetxBx(jmp_beg - (b.code.Len() + 1));
		  	b.code.Push(i);

		case NODE_AND, NODE_OR:
			// receiver
			if reg >= b.regc { b.regc = reg + 1; }
			self.args[0].compile(vm, c, b, reg);
// Appears to be erroneously compiling the same node twice
//			self.args[0].compile(vm, c, b, reg);
			if self.ntype == NODE_AND {
				b.code.Push(MachineOp{OpCode: TR_OP_JMPUNLESS, A: reg});
			} else {
				b.code.Push(MachineOp{OpCode: TR_OP_JMPIF, A: reg});
			}
			jmp := b.code.Len() - 1;

			// arg
			if reg >= b.regc { b.regc = reg + 1; }
			self.args[1].compile(vm, c, b, reg);
			b.code.At(jmp).Set_sBx(b.code.Len() - jmp - 1);

		case NODE_BOOL:
			b.code.Push(MachineOp{OpCode: TR_OP_BOOL, A: reg, B: self.args[0]});

		case NODE_NIL:
			b.code.Push(MachineOp{OpCode: TR_OP_NIL, A: reg});

		case NODE_SELF:
			b.code.Push(MachineOp{OpCode: TR_OP_SELF, A: reg});

		case NODE_RETURN:
			if self.args[0] {
				if reg >= b.regc { b.regc = reg + 1; }
				self.args[0].compile(vm, c, b, reg);
			}
			if b.parent {
				b.code.Push(MachineOp{OpCode: TR_OP_THROW, A: TR_THROW_RETURN, B: reg});
			} else {
				b.code.Push(MachineOp{OpCode: TR_OP_RETURN, A: reg});
			}

		case NODE_BREAK:
			b.code.Push(MachineOp{OpCode: TR_OP_THROW, A: TR_THROW_BREAK});

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
				if start_reg != reg {
					vm.throw_reason = TR_THROW_EXCEPTION;
					vm.throw_value = TrException_new(vm, vm.cSyntaxError, tr_sprintf(vm, "Can't create local variable inside yield"));
					return TR_UNDEF;
				}
			}
			b.code.Push(MachineOp{OpCode: TR_OP_YIELD, A: reg, B:argc});

		case NODE_DEF: {
			method := self.args[0];
			assert(method.ntype == NODE_METHOD);
			blk := c.newBlock(nil);
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
			blk.code.Push(MachineOp{OpCode: TR_OP_RETURN, A: blk_reg});

			if method.args[0] {
				// metaclass def
				if reg >= b.regc { b.regc = reg + 1; }
				method.args[0].compile(vm, c, b, reg);
				b.code.Push(newExtendedOP(TR_OP_METADEF, blki, b.push_value(method.args[1])));
				b.code.Push(MachineOp{OpCode: TR_OP_BOING, A: reg});
			}


			} else {
				b.code.Push(newExtendedOP(TR_OP_DEF, blki, b.push_value(method.args[1])));
			}

		case NODE_CLASS, NODE_MODULE:
			blk := c.newBlock(nil);
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
			blk.code.Push(MachineOp{OpCode: TR_OP_RETURN, A: reg});

			if (self.ntype == NODE_CLASS) {
				// superclass
				if self.args[1] {
					b.code.Push(newExtendedOP(TR_OP_GETCONST, reg, b.push_value(self.args[1])));
				} else {
					b.code.Push(MachineOp{OpCode: TR_OP_NIL, A: reg});
				}
				b.code.Push(newExtendedOP(TR_OP_CLASS, blki, b.push_value(self.args[0])));
				b.code.Push(MachineOp{OpCode: TR_OP_BOING, A: reg});

			} else {
				b.code.Push(newExtendedOP(TR_OP_MODULE, blki, b.push_value(self.args[0])));
			}

		case NODE_CONST:
			b.code.Push(newExtendedOP(TR_OP_GETCONST, reg, b.push_value(self.args[0])));

		case NODE_SETCONST:
			if reg >= b.regc { b.regc = reg + 1; }
			self.args[1].compile(vm, c, b, reg);
			b.code.Push(newExtendedOP(TR_OP_SETCONST, reg, b.push_value(self.args[0])));

		case NODE_ADD, NODE_SUB, NODE_LT:
			rcv := self.args[0].compile_to_RK(vm, c, b, reg);
			arg := self.args[1].compile_to_RK(vm, c, b, reg + 1);
			if (reg + 1) >= b.regc { b.regc = reg + 2; }
			switch self.ntype {
				case NODE_ADD:	b.code.Push(MachineOp{OpCode: TR_OP_ADD, A: reg, B: rcv, C: arg});
				case NODE_SUB:	b.code.Push(MachineOp{OpCode: TR_OP_SUB, A: reg, B: rcv, C: arg});
				case NODE_LT:	b.code.Push(MachineOp{OpCode: TR_OP_LT, A: reg, B: rcv, C: arg});
				default:		assert(0);
			}

		case NODE_NEG, NODE_NOT:
			rcv := self.args[0].compile_to_RK(vm, c, b, reg);
			switch self.ntype {
				case NODE_NEG:	b.code.Push(MachineOp{OpCode: TR_OP_NEG, A: reg, B: rcv});
				case NODE_NOT:	b.code.Push(MachineOp{OpCode: TR_OP_NOT, A: reg, B: rcv});
				default:		assert(0);
			}

		default:
			if !b.filename.(String) && !b.filename.(Symbol) {
				vm.throw_reason = TR_THROW_EXCEPTION;
				vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + b.filename));
				return TR_UNDEF;
			}
			printf("Compiler: unknown node type: %d in %s:%lu\n", self.ntype, b.filename.ptr, b.line);
			if vm.debug { assert(0); }
	}
	return TR_NIL;
}

func (self *Compiler) compile {
	b := self.block;
	b.filename = self.filename;
	self.node.compile(self.vm, c, b, 0);
	b.code.Push(MachineOp{OpCode: TR_OP_RETURN});
}