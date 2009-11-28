import (
	"tr";
	"container/vector";
)

type Array struct {
	type		TR_T;
	class		RubyObject;
	ivars		map[string] RubyObject;
	values		Vector;
}

func (vm *RubyVM) newArray() RubyObject {
	return Array{type: TR_T_Array, class: vm.classes[TR_T_Array], ivars: make(map[string] RubyObject), values: Vector.New(0)};
}

// Uses variadic ... parameter which replaces the mechanism used by stdarg.h
func (vm *RubyVM) newArray2(args []RubyObject) RubyObject {
	a := vm.newArray();
	va_list argp;
	va_start(argp, argc);
	for i := 0; i < argc; ++i { a.Push(va_arg(argp, *RubyObject)); }
	va_end(argp);
	return a;
}

func (vm *RubyVM) newArray3(argc int, items []RubyObject) RubyObject {
	a := vm.newArray();
	for i := 0; i < argc; ++i { a.Push(items[i]) };
	return a;
}

func (self *Array) Push(x *RubyObject) *RubyObject {
	self.values.Push(x);
	return x;
}

func (self *Array) at2index(at *RubyObject) int {
	i := TR_FIX2INT(at);
	if i < 0 { i = self.kv.Len() + i; }
	return i;
}

func (self *Array) At(at *RubyObject) *RubyObject {
	i := self.at2index(vm, at);
	if i < 0 || i >= self.values.Len() {
		return TR_NIL;
	} else {
		return self.At(i);
	}
}

func (self *Array) set(vm *RubyVM, at, x *RubyObject) *RubyObject {
	i := self.at2index(vm, at);
	switch {
		case i < 0:
			vm.throw_reason = TR_THROW_EXCEPTION;
			vm.throw_value = TrException_new(vm, vm.cIndexError, tr_sprintf(vm, "index %d out of array", i));
			return TR_UNDEF;

		case i < self.Len():
			self.Set(at, x);

		case i > self.Len():
			self.AppendVector(Vector.New(at - self.Len()));
			fallthrough;

		case i == self.Len():
			self.Push(x);
	}
	return x;
}

func (self *Array) length(vm *RubyVM) RubyObject {
	return TR_INT2FIX(self.kv.Len());
}

void TrArray_init(vm *RubyVM) {
	c := vm.classes[TR_T_Array] = Object_const_set(vm, vm.self, TrSymbol_new(vm, Array), newClass(vm, TrSymbol_new(vm, Array), vm.classes[TR_T_Object]));
	c.add_method(vm, TrSymbol_new(vm, "length"), newMethod(vm, (TrFunc *)TrArray_length, TR_NIL, 0));
	c.add_method(vm, TrSymbol_new(vm, "size"), newMethod(vm, (TrFunc *)TrArray_length, TR_NIL, 0));
	c.add_method(vm, TrSymbol_new(vm, "<<"), newMethod(vm, (TrFunc *)Push, TR_NIL, 1));
	c.add_method(vm, TrSymbol_new(vm, "[]"), newMethod(vm, (TrFunc *)TrArray_at, TR_NIL, 1));
	c.add_method(vm, TrSymbol_new(vm, "[]="), newMethod(vm, (TrFunc *)TrArray_set, TR_NIL, 2));
}