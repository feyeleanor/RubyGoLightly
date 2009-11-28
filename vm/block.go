import (
	"tr";
	"opcode";
	"fmt";
	"container/vector";
)

type Block struct {
	// static
	k			Vector;
	strings		StringVector;
	locals		Vector;
	upvals		Vector;
	code		Vector;
	defaults	Vector;
	blocks		[]Block;
	regc		int;
	argc		int;
	arg_splat	int;
	filename	RubyObject;
	line		int;
	parent 		*Block;
	// dynamic
	sites		Vector;
}

func (compiler *Compiler) newBlock(parent *Block) *Block {
	return Block{	parent:		parent,
					k:			Vector.New(0),
					strings:	StringVector.new(0),
					locals: 	Vector.new(0),
					upvals:		Vector.new(0),
					code:		Vector.new(0),
					defaults:	Vector.new(0),
					sites:		Vector.new(0),
					blocks:		Vector.new(0),
					regc:		0,
					argc:		0,
					line:		1,
					filename:	compiler.filename,
				 }
}

func (b *Block) dump(vm *RubyVM, level int) RubyObject {
	char buf[10];
  
	size_t i;
	fmt.Println("; block definition: %p (level %d)", b, level);
	fmt.Println("; %lu registers ; %lu nested blocks", b.regc, b.blocks.Len());
	fmt.Printf("; %lu args ", b.argc);
	if (b.arg_splat) { fmt.Printf(", splat"); }
	fmt.Println()
	if b.defaults.Len > 0 {
		fmt.Printf("; defaults table: ");
		for (i = 0; i < b.defaults.Len; ++i) fmt.Printf("%d ", b.defaults.At(i));
		fmt.Println();
	}
	for (i = 0; i < b.locals.Len(); ++i) {
		local := b.locals.At(i);
		if local.(Symbol) {
			local := local.ptr;
		} else {
			sprintf(buf, "%d", TR_FIX2INT(local));
			local := buf;
		}
		fmt.Println(".local  %-8s ; %lu", local, i);
	}
	for (i = 0; i < b.upvals.Len(); ++i) {
		upval := b.upvals.At(i);
		if upval.(Symbol) {
			upval := upval.ptr;
		} else {
			sprintf(buf, "%d", TR_FIX2INT(upval));
			upval := buf;
		}
		fmt.Println(".upval  %-8s ; %lu", upval, i);
	}
	for (i = 0; i < b.k.Len(); ++i) {
		k := b.k.At(i);
		if k.(Symbol) {
			k := k.ptr;
		} else {
			sprintf(buf, "%d", TR_FIX2INT(k));
			k := buf;
		}
		fmt.Println(".value  %-8s ; %lu",k, i);
	}
	for (i = 0; i < b.strings.Len; ++i) {
		fmt.Println(".string %-8s ; %lu", b.strings.At(i), i);
	}
	for (i = 0; i < b.code.Len(); ++i) {
		op := b.code.At(i);
		fmt.printf("[%03lu] %-10s %3d %3d %3d", i, OPCODE_NAMES[op.OpCode], op.A, op.B, op.C);
		switch (op.OpCode) {
			case TR_OP_LOADK:
				k := b.k.At(op.Get_Bx());
				if k.(Symbol) {
					k := k.ptr;
				} else {
					sprintf(buf, "%d", TR_FIX2INT(k));
					k := buf;
				}
				fmt.Printf(" ; R[%d] = %s", op.A, k);

			case TR_OP_STRING:
				fmt.Printf(" ; R[%d] = \"%s\"", op.A, b.strings.At(op.Get_Bx()));

			case TR_OP_LOOKUP:
				k := b.l.At(op.Get_Bx());
				if k.(Symbol) {
					k := k.ptr;
				} else {
					sprintf(buf, "%d", TR_FIX2INT(k));
					k := buf;
				}
				fmt.Printf(" ; R[%d] = R[%d].method(:%s)", op.A + 1, op.A, k);

			case TR_OP_CALL:
				fmt.Printf(" ; R[%d] = R[%d].R[%d](%d)", op.A, op.A, op.A + 1, op.B >> 1);

			case TR_OP_SETUPVAL:
				upval := b.upvals.At(op.B);
				if upval.(Symbol) {
					upval := upval.ptr;
				} else {
					sprintf(buf, "%d", TR_FIX2INT(upval));
					upval := buf;
				}
				fmt.Printf(" ; %s = R[%d]", upval, op.A);

			case TR_OP_GETUPVAL:
				upval := b.upvals.At(op.B);
				if upval.(Symbol) {
					upval := upval.ptr;
				} else {
					sprintf(buf, "%d", TR_FIX2INT(upval));
					upval := buf;
				}
				fmt.Printf(" ; R[%d] = %s", op.A, upval);

			case TR_OP_JMP:
				fmt.Printf(" ; %d", op.Get_sBx());

			case TR_OP_DEF:
				k := b.k.At(op.Get_Bx());
				if k.(Symbol) {
					k := k.ptr;
				} else {
					sprintf(buf, "%d", TR_FIX2INT(k));
					k := buf;
				}
				fmt.Printf(" ; %s => %p", k, b.blocks.At(op.A));
		}
		fmt.Println();
	}
	fmt.Println("; block end\n");

	for (i = 0; i < b.blocks.Len(); ++i) { b.blocks.At(i).dump(vm, level+1); }
	return TR_NIL;
}

func (block *Block) push_value(k *RubyObject) int {
	size_t i;
	for i = 0; i < block.k.Len(); ++i {
		if block.k.At(i) == k { return i; }
	}
	block.k.Push(k);
	return block.k.Len() - 1;
}

func (block *Block) push_string(str *string) int {
	size_t i;
	for (i = 0; i < blk.strings.Len; ++i) {
		if strcmp(blk.strings.At(i) == str { return i; }
	}
	len := strlen(str);
	ptr := make([]byte, len + 1);
	memcpy(ptr, str, sizeof(char) * (len + 1));
	blk.strings.Push ptr
	return blk.strings.Len - 1;
}

func (block *Block) find_local(name *RubyObject) int {
	size_t i;
	for (i = 0; i < blk.locals.Len(); ++i) {
		if blk.locals.At(i) == name { return i; }
	}
	return -1;
}

func (block *Block) push_local(name *RubyObject) int {
	i = block.find_local(name);
	if i != -1 { return i; }
	block.locals.Push(name);
	return block.locals.Len() - 1;
}

func (block *Block) find_upval(name *RubyObject) int {
	for (i = 0; i < block.upvals.Len(); ++i) {
		if block.upvals.At(i) == name { return i; }
	}
	return -1;
}

func (block *Block) find_upval_in_scope(name *RubyObject) int {
	if (!block.parent) { return -1; }
	i = -1;
	while (block && (i = block.find_local(name)) == -1) { block = block.parent; }
	return i;
}

func (block *Block) push_upval(name *RubyObject) int {
	i = block.find_upval(name);
	if (i != -1) return i;

	Block *b = block;
	while (b.parent) {
		if b.find_upval(name) == -1 { b.upvals.Push(name); }
		b = b.parent;
	}
	return block.upvals.Len()-1;
}