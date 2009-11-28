import (
	"tr";
)

func TrNil_to_s(vm *RubyVM, self *RubyObject) RubyObject {
	return TrString_new2(vm, "");
}

func TrTrue_to_s(vm *RubyVM, self *RubyObject) RubyObject {
	return TrString_new2(vm, "true");
}

func TrFalse_to_s(vm *RubyVM, self *RubyObject) RubyObject {
	return TrString_new2(vm, "false");
}

func TrPrimitive_init(vm *RubyVM) {
	nilc := vm.classes[TR_T_NilClass] = Object_const_set(vm, vm.self, TrSymbol_new(vm, NilClass), newClass(vm, TrSymbol_new(vm, NilClass), vm.classes[TR_T_Object]));
	truec := vm.classes[TR_T_TrueClass] = Object_const_set(vm, vm.self, TrSymbol_new(vm, TrueClass), newClass(vm, TrSymbol_new(vm, TrueClass), vm.classes[TR_T_Object]));
	falsec := vm.classes[TR_T_FalseClass] = Object_const_set(vm, vm.self, TrSymbol_new(vm, FalseClass), newClass(vm, TrSymbol_new(vm, FalseClass), vm.classes[TR_T_Object]));
	nilc.add_method(vm, TrSymbol_new(vm, "to_s"), newMethod(vm, (TrFunc *)TrNil_to_s, TR_NIL, 0));
	truec.add_method(vm, TrSymbol_new(vm, "to_s"), newMethod(vm, (TrFunc *)TrTrue_to_s, TR_NIL, 0));
	falsec.add_method(vm, TrSymbol_new(vm, "to_s"), newMethod(vm, (TrFunc *)TrFalse_to_s, TR_NIL, 0));
}