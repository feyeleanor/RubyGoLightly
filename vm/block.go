import (
	"tr";
	"opcode";
	"internal";
	"fmt";
	"container/vector";
)

type Block struct {
	// static
	k			[]OBJ;
	strings		*StringVector;
	locals		[]OBJ;
	upvals		[]OBJ;
	code		[]TrInst;
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
	block.defaults = NewIntVector(0)
	block.strings = NewStringVector(0)
	block.filename = compiler.filename;
	block.line = 1;
	block.regc = 0;
	block.argc = 0;
	block.parent = parent;
	kv_init(block.k);
	kv_init(block.code);
	kv_init(block.locals);
	kv_init(block.sites);
	return block;
}

#define INSPECT_K(K)  (K.(Symbol) ? TR_STR_PTR(K) : (sprintf(buf, "%d", TR_FIX2INT(K)), buf))

func (b *Block) dump2(vm *TrVM, level int) OBJ {
	char buf[10];
  
	size_t i;
	fmt.Println("; block definition: %p (level %d)", b, level);
	fmt.Println("; %lu registers ; %lu nested blocks", b.regc, kv_size(b.blocks));
	fmt.Printf("; %lu args ", b.argc);
	if (b.arg_splat) { fmt.Printf(", splat"); }
	fmt.Println()
	if b.defaults.Len > 0 {
		fmt.Printf("; defaults table: ");
		for (i = 0; i < b.defaults.Len; ++i) fmt.Printf("%d ", b.defaults.At(i));
		fmt.Println();
	}
	for (i = 0; i < kv_size(b.locals); ++i) { fmt.Println(".local  %-8s ; %lu", INSPECT_K(kv_A(b.locals, i)), i); }
	for (i = 0; i < kv_size(b.upvals); ++i) { fmt.Println(".upval  %-8s ; %lu", INSPECT_K(kv_A(b.upvals, i)), i); }
	for (i = 0; i < kv_size(b.k); ++i) { fmt.Println(".value  %-8s ; %lu", INSPECT_K(kv_A(b.k, i)), i); }
	for (i = 0; i < b.strings.Len; ++i) { fmt.Println(".string %-8s ; %lu", kv_A(b.strings, i), i); }
	for (i = 0; i < kv_size(b.code); ++i) {
		TrInst op = kv_A(b.code, i);
		fmt.printf("[%03lu] %-10s %3d %3d %3d", i, OPCODE_NAMES[GET_OPCODE(op)], GETARG_A(op), GETARG_B(op), GETARG_C(op));
		switch (GET_OPCODE(op)) {
			case TR_OP_LOADK:    fmt.Printf(" ; R[%d] = %s", GETARG_A(op), INSPECT_K(kv_A(b.k, GETARG_Bx(op))));
			case TR_OP_STRING:   fmt.Printf(" ; R[%d] = \"%s\"", GETARG_A(op), kv_A(b.strings, GETARG_Bx(op)));
			case TR_OP_LOOKUP:   fmt.Printf(" ; R[%d] = R[%d].method(:%s)", GETARG_A(op)+1, GETARG_A(op), INSPECT_K(kv_A(b.k, GETARG_Bx(op))));
			case TR_OP_CALL:     fmt.Printf(" ; R[%d] = R[%d].R[%d](%d)", GETARG_A(op), GETARG_A(op), GETARG_A(op)+1, GETARG_B(op)>>1);
			case TR_OP_SETUPVAL: fmt.Printf(" ; %s = R[%d]", INSPECT_K(kv_A(b.upvals, GETARG_B(op))), GETARG_A(op));
			case TR_OP_GETUPVAL: fmt.Printf(" ; R[%d] = %s", GETARG_A(op), INSPECT_K(kv_A(b.upvals, GETARG_B(op))));
			case TR_OP_JMP:      fmt.Printf(" ; %d", GETARG_sBx(op));
			case TR_OP_DEF:      fmt.Printf(" ; %s => %p", INSPECT_K(kv_A(b.k, GETARG_Bx(op))), kv_A(b.blocks, GETARG_A(op)));
		}
		fmt.Println();
	}
	fmt.Println("; block end\n");

	for (i = 0; i < kv_size(b.blocks); ++i) { kv_A(b.blocks, i).dump2(vm, level+1); }
	return TR_NIL;
}

func (b *Block) dump(vm *TrVM) {
	b.dump2(vm, 0);
}

func (block *Block) push_value(k OBJ) int {
	size_t i;
	for (i = 0; i < kv_size(block.k); ++i) { if (kv_A(block.k, i) == k) return i; }
	kv_push(OBJ, block.k, k);
	return kv_size(block.k) - 1;
}

func (block *Block) push_string(str *char) int {
	size_t i;
	for (i = 0; i < blk.strings.Len; ++i) {
		if (strcmp(kv_A(blk.strings, i), str) == 0) { return i; }
	}
	int len = strlen(str);
	char *ptr = TR_ALLOC_N(char, len+1);
	TR_MEMCPY_N(ptr, str, char, len+1);
	blk.strings.Push ptr
	return blk.strings.Len - 1;
}

func (block *Block) find_local(name OBJ) int {
	size_t i;
	for (i = 0; i < kv_size(blk.locals); ++i) {
		if (kv_A(blk.locals, i) == name) { return i; }
	}
	return -1;
}

func (block *Block) push_local(name OBJ) int {
	i = block.find_local(name);
	if i != -1 { return i; }
	kv_push(OBJ, block.locals, name);
	return kv_size(block.locals) - 1;
}

func (block *Block) find_upval(name OBJ) int {
	for (i = 0; i < kv_size(block.upvals); ++i) {
		if (kv_A(block.upvals, i) == name) { return i; }
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
		if b.find_upval(name) == -1 { kv_push(OBJ, b.upvals, name); }
		b = b.parent;
	}
	return kv_size(block.upvals)-1;
}