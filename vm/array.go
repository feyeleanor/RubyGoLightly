import (
	"tr";
	"internal";
	"container/vector";
)

// rephrase arrays in terms of the stdlib vector which is dynamically expandable

/* array macros */
#define TR_ARRAY_PUSH(X,I)   kv_push(OBJ, ((TrArray*)(X)).kv, (I))
#define TR_ARRAY_AT(X,I)     kv_A((TR_CARRAY(X)).kv, (I))
#define TR_ARRAY_SIZE(X)     kv_size(TR_CARRAY(X).kv)
#define TR_ARRAY_EACH(T,I,V,B) ({ \
    Array *__a##V = TR_CARRAY(T); \
    if (kv_size(__a##V.kv) != 0) { \
      size_t I; \
      for (I = 0; I < kv_size(__a##V.kv); I++) { \
        OBJ V = kv_A(__a##V.kv, I); \
        B \
      } \
    } \
  })

type Array struct {
	type		TR_T;
	class		OBJ;
	ivars		*khash_t(OBJ);
	kv			ObjectVector;
}

func newArray(vm struct TrVM *) OBJ {
  Array *a = TR_INIT_CORE_OBJECT(Array);
  kv_init(a.kv);
  return (OBJ)a;
}

// Uses variadic ... parameter which replaces the mechanism used by stdarg.h
func newArray2(vm struct TrVM *, argc int, ...) OBJ {
  OBJ a = newArray(vm);
  va_list argp;
  int i;
  va_start(argp, argc);
  for (i = 0; i < argc; ++i) TR_ARRAY_PUSH(a, va_arg(argp, OBJ));
  va_end(argp);
  return a;
}

func newArray3(vm struct TrVM *, argc int, items []OBJ) OBJ {
  OBJ a = newArray(vm);
  int i;
  for (i = 0; i < argc; ++i) TR_ARRAY_PUSH(a, items[i]);
  return a;
}

func (self *Array) push(vm struct TrVM *, x OBJ) OBJ {
	TR_ARRAY_PUSH(self, x);
	return x;
}

func (self *Array) at2index(vm struct TrVM *, at OBJ) int {
  int i = TR_FIX2INT(at);
  if (i < 0) i = TR_ARRAY_SIZE(self) + i;
  return i;
}

func (self *Array) at(vm struct TrVM *, at OBJ) OBJ {
  int i = self.at2index(vm, at);
  if (i < 0 || i >= (int)TR_ARRAY_SIZE(self)) return TR_NIL;
  return TR_ARRAY_AT(self, i);
}

func (self *Array) set(vm struct TrVM *, at, x OBJ) OBJ {
  int i = self.at2index(vm, at);
  if (i < 0) tr_raise(IndexError, "index %d out of array", i);
  kv_a(OBJ, (TR_CARRAY(self)).kv, i) = x;
  return x;
}

func (self *Array) length(vm struct TrVM *) OBJ {
  return TR_INT2FIX(TR_ARRAY_SIZE(self));
}

void TrArray_init(vm struct TrVM *) {
  OBJ c = TR_INIT_CORE_CLASS(Array, Object);
  tr_def(c, "length", TrArray_length, 0);
  tr_def(c, "size", TrArray_length, 0);
  tr_def(c, "<<", TrArray_push, 1);
  tr_def(c, "[]", TrArray_at, 1);
  tr_def(c, "[]=", TrArray_set, 2);
}