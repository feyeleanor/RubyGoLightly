import(
	"tr";
	"internal";
)

func TrFixnum_add(vm *RunyVM, self, other *RubyObject) RubyObject { return TR_INT2FIX(TR_FIX2INT(self) + TR_FIX2INT(other)); }
func TrFixnum_sub(vm *RubyVM, self, other *RubyObject) RubyObject { return TR_INT2FIX(TR_FIX2INT(self) - TR_FIX2INT(other)); }
func TrFixnum_mul(vm *RubyVM, self, other *RubyObject) RubyObject { return TR_INT2FIX(TR_FIX2INT(self) * TR_FIX2INT(other)); }
func TrFixnum_div(vm *RubyVM, self, other *RubyObject) RubyObject { return TR_INT2FIX(TR_FIX2INT(self) / TR_FIX2INT(other)); }

func TrFixnum_eq(vm *RubyVM, self, other *RubyObject) RubyObject { if TR_FIX2INT(self) == TR_FIX2INT(other) { return TR_TRUE } else { return TR_FALSE } }
func TrFixnum_ne(vm *RubyVM, self, other *RubyObject) RubyObject { if TR_FIX2INT(self) != TR_FIX2INT(other) { return TR_TRUE } else { return TR_FALSE } }
func TrFixnum_lt(vm *RubyVM, self, other *RubyObject) RubyObject { if TR_FIX2INT(self) < TR_FIX2INT(other) { return TR_TRUE } else { return TR_FALSE } }
func TrFixnum_gt(vm *RubyVM, self, other *RubyObject) RubyObject { if TR_FIX2INT(self) > TR_FIX2INT(other) { return TR_TRUE } else { return TR_FALSE } }
func TrFixnum_le(vm *RubyVM, self, other *RubyObject) RubyObject { if TR_FIX2INT(self) <= TR_FIX2INT(other) { return TR_TRUE } else { return TR_FALSE } }
func TrFixnum_ge(vm *RubyVM, self, other *RubyObject) RubyObject { if TR_FIX2INT(self) >= TR_FIX2INT(other) { return TR_TRUE } else { return TR_FALSE } }

func TrFixnum_to_s(vm *RubyVM, self *RubyObject) RubyObject { return tr_sprintf(vm, "%d", TR_FIX2INT(self)); }

void TrFixnum_init(vm *RubyVM) {
  c := vm.classes[TR_T_Fixnum] = Object_const_set(vm, vm.self, tr_intern(Fixnum), newClass(vm, tr_intern(Fixnum), vm.classes[TR_T_Object]));
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