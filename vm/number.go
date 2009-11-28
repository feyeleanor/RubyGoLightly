import(
	"tr";
)

func TrFixnum_add(vm *RunyVM, self, other *RubyObject) RubyObject {
	return TR_INT2FIX(TR_FIX2INT(self) + TR_FIX2INT(other));
}

func TrFixnum_sub(vm *RubyVM, self, other *RubyObject) RubyObject {
	return TR_INT2FIX(TR_FIX2INT(self) - TR_FIX2INT(other));
}

func TrFixnum_mul(vm *RubyVM, self, other *RubyObject) RubyObject {
	return TR_INT2FIX(TR_FIX2INT(self) * TR_FIX2INT(other));
}

func TrFixnum_div(vm *RubyVM, self, other *RubyObject) RubyObject {
	return TR_INT2FIX(TR_FIX2INT(self) / TR_FIX2INT(other));
}

func TrFixnum_eq(vm *RubyVM, self, other *RubyObject) RubyObject {
	if TR_FIX2INT(self) == TR_FIX2INT(other) {
		return TR_TRUE;
	} else {
		return TR_FALSE;
	}
}

func TrFixnum_ne(vm *RubyVM, self, other *RubyObject) RubyObject {
	if TR_FIX2INT(self) != TR_FIX2INT(other) {
		return TR_TRUE;
	} else {
		return TR_FALSE;
	}
}

func TrFixnum_lt(vm *RubyVM, self, other *RubyObject) RubyObject {
	if TR_FIX2INT(self) < TR_FIX2INT(other) {
		return TR_TRUE;
	} else {
		return TR_FALSE;
	}
}

func TrFixnum_gt(vm *RubyVM, self, other *RubyObject) RubyObject {
	if TR_FIX2INT(self) > TR_FIX2INT(other) {
		return TR_TRUE;
	} else {
		return TR_FALSE;
	}
}

func TrFixnum_le(vm *RubyVM, self, other *RubyObject) RubyObject {
	if TR_FIX2INT(self) <= TR_FIX2INT(other) {
		return TR_TRUE;
	} else {
		return TR_FALSE;
	}
}

func TrFixnum_ge(vm *RubyVM, self, other *RubyObject) RubyObject {
	if TR_FIX2INT(self) >= TR_FIX2INT(other) {
		return TR_TRUE;
	} else {
		return TR_FALSE;
	}
}

func TrFixnum_to_s(vm *RubyVM, self *RubyObject) RubyObject {
	return tr_sprintf(vm, "%d", TR_FIX2INT(self));
}

void TrFixnum_init(vm *RubyVM) {
	c := vm.classes[TR_T_Fixnum] = Object_const_set(vm, vm.self, TrSymbol_new(vm, Fixnum), newClass(vm, TrSymbol_new(vm, Fixnum), vm.classes[TR_T_Object]));
	c.add_method(vm, TrSymbol_new(vm, "+"), newMethod(vm, (TrFunc *)TrFixnum_add, TR_NIL, 1));
	c.add_method(vm, TrSymbol_new(vm, "-"), newMethod(vm, (TrFunc *)TrFixnum_sub, TR_NIL, 1));
	c.add_method(vm, TrSymbol_new(vm, "*"), newMethod(vm, (TrFunc *)TrFixnum_mul, TR_NIL, 1));
	c.add_method(vm, TrSymbol_new(vm, "/"), newMethod(vm, (TrFunc *)TrFixnum_div, TR_NIL, 1));
	c.add_method(vm, TrSymbol_new(vm, "=="), newMethod(vm, (TrFunc *)TrFixnum_eq, TR_NIL, 1));
	c.add_method(vm, TrSymbol_new(vm, "!="), newMethod(vm, (TrFunc *)TrFixnum_eq, TR_NIL, 1));
	c.add_method(vm, TrSymbol_new(vm, "<"), newMethod(vm, (TrFunc *)TrFixnum_lt, TR_NIL, 1));
	c.add_method(vm, TrSymbol_new(vm, "<="), newMethod(vm, (TrFunc *)TrFixnum_le, TR_NIL, 1));
	c.add_method(vm, TrSymbol_new(vm, ">"), newMethod(vm, (TrFunc *)TrFixnum_gt, TR_NIL, 1));
	c.add_method(vm, TrSymbol_new(vm, ">="), newMethod(vm, (TrFunc *)TrFixnum_ge, TR_NIL, 1));
	c.add_method(vm, TrSymbol_new(vm, "to_s"), newMethod(vm, (TrFunc *)TrFixnum_to_s, TR_NIL, 0));
}