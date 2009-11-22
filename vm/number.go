#include "tr.h"
#include "internal.h"

#define MATH(A,OP,B)  TR_INT2FIX(TR_FIX2INT(A) OP TR_FIX2INT(B))
#define CMP(A,OP,B)   TR_BOOL(TR_FIX2INT(A) OP TR_FIX2INT(B))

OBJ TrFixnum_add(vm *struct TrVM, OBJ self, OBJ other) { return MATH(self, +, other); }
OBJ TrFixnum_sub(vm *struct TrVM, OBJ self, OBJ other) { return MATH(self, -, other); }
OBJ TrFixnum_mul(vm *struct TrVM, OBJ self, OBJ other) { return MATH(self, *, other); }
OBJ TrFixnum_div(vm *struct TrVM, OBJ self, OBJ other) { return MATH(self, /, other); }

OBJ TrFixnum_eq(vm *struct TrVM, OBJ self, OBJ other) { return CMP(self, ==, other); }
OBJ TrFixnum_ne(vm *struct TrVM, OBJ self, OBJ other) { return CMP(self, !=, other); }
OBJ TrFixnum_lt(vm *struct TrVM, OBJ self, OBJ other) { return CMP(self, <, other); }
OBJ TrFixnum_gt(vm *struct TrVM, OBJ self, OBJ other) { return CMP(self, >, other); }
OBJ TrFixnum_le(vm *struct TrVM, OBJ self, OBJ other) { return CMP(self, <=, other); }
OBJ TrFixnum_ge(vm *struct TrVM, OBJ self, OBJ other) { return CMP(self, >=, other); }

OBJ TrFixnum_to_s(vm *struct TrVM, OBJ self) {
  return tr_sprintf(vm, "%d", TR_FIX2INT(self));
}

void TrFixnum_init(vm *struct TrVM) {
  OBJ c = TR_INIT_CORE_CLASS(Fixnum, Object);
  tr_def(c, "+", TrFixnum_add, 1);
  tr_def(c, "-", TrFixnum_sub, 1);
  tr_def(c, "*", TrFixnum_mul, 1);
  tr_def(c, "/", TrFixnum_div, 1);
  tr_def(c, "==", TrFixnum_eq, 1);
  tr_def(c, "!=", TrFixnum_eq, 1);
  tr_def(c, "<", TrFixnum_lt, 1);
  tr_def(c, "<=", TrFixnum_le, 1);
  tr_def(c, ">", TrFixnum_gt, 1);
  tr_def(c, ">=", TrFixnum_ge, 1);
  tr_def(c, "to_s", TrFixnum_to_s, 0);
}