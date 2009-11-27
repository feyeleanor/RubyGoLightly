import (
	"tr";
	"internal";
	"container/vector";
)

type Array struct {
	type		TR_T;
	class		RubyObject;
	ivars		*map[string] RubyObject;
	kv			*Vector;
}

func newArray(vm *RubyVM) RubyObject {
	return Array{type: TR_T_Array, class: vm.classes[TR_T_Array], ivars: kh_init(RubyObject), kv: Vector.New(0)};
}

// Uses variadic ... parameter which replaces the mechanism used by stdarg.h
func newArray2(vm *RubyVM, argc int, ...) RubyObject {
  a := newArray(vm);
  va_list argp;
  int i;
  va_start(argp, argc);
  for (i = 0; i < argc; ++i) a.kv.Push(va_arg(argp, *RubyObject));
  va_end(argp);
  return a;
}

func newArray3(vm *RubyVM, argc int, items []RubyObject) RubyObject {
  a := newArray(vm);
  for i := 0; i < argc; ++i { a.kv.Push(items[i]) };
  return a;
}

func (self *Array) push(vm *RubyVM, x *RubyObject) *RubyObject {
	self.kv.Push(x);
	return x;
}

func (self *Array) at2index(vm *RubyVM, at *RubyObject) int {
  int i = TR_FIX2INT(at);
  if (i < 0) i = self.kv.Len() + i;
  return i;
}

func (self *Array) at(vm *RubyVM, at *RubyObject) *RubyObject {
  i := self.at2index(vm, at);
  if i < 0 || i >= self.kv.Len() { return TR_NIL; }
  return self.kv.At(i);
}

func (self *Array) set(vm *RubyVM, at, x *RubyObject) *RubyObject {
	i := self.at2index(vm, at);
	switch {
		case i < 0:				tr_raise(IndexError, "index %d out of array", i);
		case i < self.Len():	self.Set(at, x);
		case i > self.Len():	self.AppendVector(Vector.New(at - self.Len()));
								fallthrough;
		case i == self.Len():	self.Push(x);
	}
	return x;
}

func (self *Array) length(vm *RubyVM) RubyObject {
  return TR_INT2FIX(self.kv.Len());
}

void TrArray_init(vm *RubyVM) {
  c := vm.classes[TR_T_Array] = Object_const_set(vm, vm.self, tr_intern(Array), newClass(vm, tr_intern(Array), vm.classes[TR_T_Object]));
  tr_def(c, "length", TrArray_length, 0);
  tr_def(c, "size", TrArray_length, 0);
  tr_def(c, "<<", TrArray_push, 1);
  tr_def(c, "[]", TrArray_at, 1);
  tr_def(c, "[]=", TrArray_set, 2);
}