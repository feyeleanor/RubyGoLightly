import (
	"tr";
	"internal";
	"container/vector";
)

// rephrase arrays in terms of the stdlib vector which is dynamically expandable

// array macros
#define TR_ARRAY_EACH(T,I,V,B) ({			\
	index := 0; for V := range T.Iter() {	\
		B;									\
		index++;							\
  	}										\
})

type Array struct {
	type		TR_T;
	class		OBJ;
	ivars		*map[string] OBJ;
	kv			*Vector;
}

func newArray(vm struct TrVM *) OBJ {
	Array *a = TR_INIT_CORE_OBJECT(Array);
	a.kv = Vector.New(0);
	return OBJ(a);
}

// Uses variadic ... parameter which replaces the mechanism used by stdarg.h
func newArray2(vm struct TrVM *, argc int, ...) OBJ {
  OBJ a = newArray(vm);
  va_list argp;
  int i;
  va_start(argp, argc);
  for (i = 0; i < argc; ++i) a.kv.Push(va_arg(argp, OBJ));
  va_end(argp);
  return a;
}

func newArray3(vm struct TrVM *, argc int, items []OBJ) OBJ {
  a := newArray(vm);
  for i := 0; i < argc; ++i { a.kv.Push(items[i]) };
  return a;
}

func (self *Array) push(vm struct TrVM *, x OBJ) OBJ {
	self.kv.Push(x);
	return x;
}

func (self *Array) at2index(vm struct TrVM *, at OBJ) int {
  int i = TR_FIX2INT(at);
  if (i < 0) i = self.kv.Len() + i;
  return i;
}

func (self *Array) at(vm struct TrVM *, at OBJ) OBJ {
  i := self.at2index(vm, at);
  if i < 0 || i >= self.kv.Len() { return TR_NIL; }
  return self.kv.At(i);
}

func (self *Array) set(vm struct TrVM *, at, x OBJ) OBJ {
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

func (self *Array) length(vm struct TrVM *) OBJ {
  return TR_INT2FIX(self.kv.Len());
}

void TrArray_init(vm struct TrVM *) {
  OBJ c = TR_INIT_CORE_CLASS(Array, Object);
  tr_def(c, "length", TrArray_length, 0);
  tr_def(c, "size", TrArray_length, 0);
  tr_def(c, "<<", TrArray_push, 1);
  tr_def(c, "[]", TrArray_at, 1);
  tr_def(c, "[]=", TrArray_set, 2);
}