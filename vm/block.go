import (
	"tr";
	"opcode";
	"internal";
	"fmt";
	"container/vector";
)

type Block struct {
	// static
	k			*Vector;
	strings		*StringVector;
	locals		*Vector;
	upvals		*Vector;
	code		*Vector;
	defaults	*IntVector;
	blocks		[]Block;
	regc		int;
	argc		int;
	arg_splat	int;
	filename	OBJ;
	line		int;
	parent 		*Block;
	// dynamic
	sites		CallSiteVector;
}


func newBlock(compiler *Compiler, parent *Block) *Block {
	block = new(Block);
	block.defaults = IntVector.New(0)
	block.strings = StringVector.New(0)
	block.filename = compiler.filename;
	block.line = 1;
	block.regc = 0;
	block.argc = 0;
	block.parent = parent;
	block.k = Vector.New(0);
	block.code = Vector.New(0);
	block.locals = Vector.New(0);
	block.sites = Vector.New(0);
	return block;
}

#define INSPECT_K(K)  (K.(Symbol) ? TR_STR_PTR(K) : (sprintf(buf, "%d", TR_FIX2INT(K)), buf))

func (b *Block) dump2(vm *RubyVM, level int) OBJ {
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
	for (i = 0; i < b.locals.Len(); ++i) { fmt.Println(".local  %-8s ; %lu", INSPECT_K(b.locals.At(i)), i); }
	for (i = 0; i < b.upvals.Len(); ++i) { fmt.Println(".upval  %-8s ; %lu", INSPECT_K(b.upvals.At(i)), i); }
	for (i = 0; i < b.k.Len(); ++i) { fmt.Println(".value  %-8s ; %lu", INSPECT_K(b.k.At(i)), i); }
	for (i = 0; i < b.strings.Len; ++i) { fmt.Println(".string %-8s ; %lu", b.strings.At(i), i); }
	for (i = 0; i < b.code.Len(); ++i) {
		op := b.code.At(i);
		fmt.printf("[%03lu] %-10s %3d %3d %3d", i, OPCODE_NAMES[op.OpCode], op.A, op.B, op.C);
		switch (op.OpCode) {
			case TR_OP_LOADK:    fmt.Printf(" ; R[%d] = %s", op.A, INSPECT_K(b.k.At(op.Get_Bx())));
			case TR_OP_STRING:   fmt.Printf(" ; R[%d] = \"%s\"", op.A, b.strings.At(op.Get_Bx()));
			case TR_OP_LOOKUP:   fmt.Printf(" ; R[%d] = R[%d].method(:%s)", op.A + 1, op.A, INSPECT_K(b.k.At(op.Get_Bx())));
			case TR_OP_CALL:     fmt.Printf(" ; R[%d] = R[%d].R[%d](%d)", op.A, op.A, op.A + 1, op.B >> 1);
			case TR_OP_SETUPVAL: fmt.Printf(" ; %s = R[%d]", INSPECT_K(b.upvals.At(op.B)), op.A);
			case TR_OP_GETUPVAL: fmt.Printf(" ; R[%d] = %s", op.A, INSPECT_K(b.upvals.At(op.B)));
			case TR_OP_JMP:      fmt.Printf(" ; %d", op.Get_sBx());
			case TR_OP_DEF:      fmt.Printf(" ; %s => %p", INSPECT_K(b.k.At(op.Get_Bx())), b.blocks.At(op.A));
		}
		fmt.Println();
	}
	fmt.Println("; block end\n");

	for (i = 0; i < b.blocks.Len(); ++i) { b.blocks.At(i).dump2(vm, level+1); }
	return TR_NIL;
}

func (b *Block) dump(vm *RubyVM) {
	b.dump2(vm, 0);
}

func (block *Block) push_value(k OBJ) int {
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

func (block *Block) find_local(name OBJ) int {
	size_t i;
	for (i = 0; i < blk.locals.Len(); ++i) {
		if blk.locals.At(i) == name { return i; }
	}
	return -1;
}

func (block *Block) push_local(name OBJ) int {
	i = block.find_local(name);
	if i != -1 { return i; }
	block.locals.Push(name);
	return block.locals.Len() - 1;
}

func (block *Block) find_upval(name OBJ) int {
	for (i = 0; i < block.upvals.Len(); ++i) {
		if block.upvals.At(i) == name { return i; }
	}
	return -1;
}

func (block *Block) find_upval_in_scope(name OBJ) int {
	if (!block.parent) { return -1; }
	i = -1;
	while (block && (i = block.find_local(name)) == -1) { block = block.parent; }
	return i;
}

func (block *Block) push_upval(name OBJ) int {
	i = block.find_upval(name);
	if (i != -1) return i;

	Block *b = block;
	while (b.parent) {
		if b.find_upval(name) == -1 { b.upvals.Push(name); }
		b = b.parent;
	}
	return block.upvals.Len()-1;
}