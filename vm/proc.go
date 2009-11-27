import (
	"tr";
	"internal";
)

type Closure struct {
	block			*Block;
	upcals			*TrUpval;
	self			*RubyObject;
	class			*RubyObject;
	parent			*Closure;
}

func newClosure(vm *RubyVM, block *Block, self, class *RubyObject, parent *Closure) Closure {
	closure = new(Closure);
	closure.block = block;
	closure.upvals = make([]TrUpval, block.upvals.Len());
	closure.self = self;
	closure.class = class;
	closure.parent = parent;
	return closure;
}