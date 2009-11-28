import (
	"tr";
	)

func TrRange_new(vm *RubyVM, first, last *RubyObject, exclusive int) RubyObject {
	return Range{type: TR_T_Range, class: vm.classes[TR_T_Range], ivars: make(map[string] RubyObject), first: first, last: last, exclusive: exclusive};
}

func TrRange_first(vm *RubyVM, self *RubyObject) RubyObject {
	if !self.(Range) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected Range"));
		return TR_UNDEF;
	}
	TrRange *(self).first;
	}

func TrRange_last(vm *RubyVM, self *RubyObject) RubyObject {
	if !self.(Range) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected Range"));
		return TR_UNDEF;
	}
	TrRange *(self).last;
}


func TrRange_exclude_end(vm *RubyVM, self *RubyObject) RubyObject {
	if !self.(Range) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected Range"));
		return TR_UNDEF;
	}
	if TrRange *(self).exclusive {
		return TR_TRUE;
	} else {
		return TR_FALSE;
	}
}

func TrRange_init(vm *RubyVM) {
	c := vm.classes[TR_T_Range] = Object_const_set(vm, vm.self, TrSymbol_new(vm, Range), newClass(vm, TrSymbol_new(vm, Range), vm.classes[TR_T_Object]));
	c.add_method(vm, TrSymbol_new(vm, "first"), newMethod(vm, (TrFunc *)TrRange_first, TR_NIL, 0));
	c.add_method(vm, TrSymbol_new(vm, "last"), newMethod(vm, (TrFunc *)TrRange_last, TR_NIL, 0));
	c.add_method(vm, TrSymbol_new(vm, "exclude_end?"), newMethod(vm, (TrFunc *)TrRange_exclude_end, TR_NIL, 0));
}