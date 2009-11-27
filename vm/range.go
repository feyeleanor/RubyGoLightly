import (
	"tr";
	"internal";
	)

func TrRange_new(vm *RubyVM, first, last *RubyObject, exclusive int) RubyObject {
	return Range{type: TR_T_Range, class: vm.classes[TR_T_Range], ivars: kh_init(RubyObject), first: first, last: last, exclusive: exclusive};
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
  c := vm.classes[TR_T_Range] = Object_const_set(vm, vm.self, tr_intern(Range), newClass(vm, tr_intern(Range), vm.classes[TR_T_Object]));
  tr_def(c, "first", TrRange_first, 0);
  tr_def(c, "last", TrRange_last, 0);
  tr_def(c, "exclude_end?", TrRange_exclude_end, 0);
}