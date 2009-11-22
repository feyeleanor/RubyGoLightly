import (
	"tr";
	"internal";
)

type Closure struct {
	block			*Block;
	upcals			*TrUpval;
	self, class		OBJ;
	parent			*Closure;
}

func newClosure(vm *struct TrVM, b *Block, self, class OBJ, parent *Closure) Closure {
	closure = new(Closure);
	closure.block = b;
	closure.upvals = make([]TrUpval, kv_size(b.upvals));
	closure.self = self;
	closure.class = class;
	closure.parent = parent;
	return closure;
}