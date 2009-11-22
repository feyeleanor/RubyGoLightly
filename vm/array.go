#include "tr.h"
#include "internal.h"

TrArray_new(vm struct TrVM *) OBJ {
  TrArray *a = TR_INIT_CORE_OBJECT(Array);
  kv_init(a.kv);
  return (OBJ)a;
}

// Uses variadic ... parameter which replaces the mechanism used by stdarg.h
func TrArray_new2(vm struct TrVM *, argc int, ...) OBJ {
  OBJ a = TrArray_new(vm);
  va_list argp;
  int i;
  va_start(argp, argc);
  for (i = 0; i < argc; ++i) TR_ARRAY_PUSH(a, va_arg(argp, OBJ));
  va_end(argp);
  return a;
}

func TrArray_new3(vm struct TrVM *, argc int, items []OBJ) OBJ {
  OBJ a = TrArray_new(vm);
  int i;
  for (i = 0; i < argc; ++i) TR_ARRAY_PUSH(a, items[i]);
  return a;
}

func TrArray_push(vm struct TrVM *, self, x OBJ) OBJ {
	TR_ARRAY_PUSH(self, x);
	return x;
}

func TrArray_at2index(vm struct TrVM *, self, at OBJ) int {
  int i = TR_FIX2INT(at);
  if (i < 0) i = TR_ARRAY_SIZE(self) + i;
  return i;
}

func TrArray_at(vm struct TrVM *, self, at OBJ) OBJ {
  int i = TrArray_at2index(vm, self, at);
  if (i < 0 || i >= (int)TR_ARRAY_SIZE(self)) return TR_NIL;
  return TR_ARRAY_AT(self, i);
}

func TrArray_set(vm struct TrVM *, self, at, x OBJ) OBJ {
  int i = TrArray_at2index(vm, self, at);
  if (i < 0) tr_raise(IndexError, "index %d out of array", i);
  kv_a(OBJ, (TR_CARRAY(self)).kv, i) = x;
  return x;
}

func TrArray_length(vm struct TrVM *, self OBJ) OBJ {
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