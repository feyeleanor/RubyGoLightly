#include "tr.h"
#include "internal.h"

#define MATH(A,OP,B)  TR_INT2FIX(TR_FIX2INT(A) OP TR_FIX2INT(B))
#define CMP(A,OP,B)   TR_BOOL(TR_FIX2INT(A) OP TR_FIX2INT(B))

func TrFixnum_add(vm *RunyVM, self, other OBJ) OBJ { return MATH(self, +, other); }
func TrFixnum_sub(vm *RubyVM, self, other OBJ) OBJ { return MATH(self, -, other); }
func TrFixnum_mul(vm *RubyVM, self, other OBJ) OBJ { return MATH(self, *, other); }
func TrFixnum_div(vm *RubyVM, self, other OBJ) OBJ { return MATH(self, /, other); }

func TrFixnum_eq(vm *RubyVM, self, other OBJ) OBJ { return CMP(self, ==, other); }
func TrFixnum_ne(vm *RubyVM, self, other OBJ) OBJ { return CMP(self, !=, other); }
func TrFixnum_lt(vm *RubyVM, self, other OBJ) OBJ { return CMP(self, <, other); }
func TrFixnum_gt(vm *RubyVM, self, other OBJ) OBJ { return CMP(self, >, other); }
func TrFixnum_le(vm *RubyVM, self, other OBJ) OBJ { return CMP(self, <=, other); }
func TrFixnum_ge(vm *RubyVM, self, other OBJ) OBJ { return CMP(self, >=, other); }
func TrFixnum_to_s(vm *RubyVM, OBJ self) OBJ { return tr_sprintf(vm, "%d", TR_FIX2INT(self)); }

void TrFixnum_init(vm *RubyVM) {
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