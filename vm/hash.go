import(
	"tr";
	)

func TrHash_new(vm *RubyVM) RubyObject {
	return Hash{type: TR_T_Hash, class: vm.classes[TR_T_Hash], ivars: make(map[string] RubyObject), hash: make(map[string] RubyObject)};
}

func TrHash_new2(vm *RubyVM, n size_t, items []RubyObject) RubyObject {
	hash := TrHash_new(vm);
	for i := 0; i < n; i += 2) {
		hash.hash[items[i]] = items[i + 1];
	}
	return h;
}

func TrHash_size(vm *RubyVM, self *RubyObject) RubyObject {
	if !self.(Hash) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected Hash"));
		return TR_UNDEF;
	}
	return TR_INT2FIX(len(self.hash));
}

// TODO use Object#hash as the key
func TrHash_get(vm *RubyVM, self, key *RubyObject) RubyObject {
	if !self.(Hash) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected Hash"));
		return TR_UNDEF;
	}
	return self.hash[key] || TR_NIL;
}

func TrHash_set(vm *RubyVM, self, key, value *RubyObject) RubyObject {
	if !self.(Hash) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected Hash"));
		return TR_UNDEF;
	}
	self.hash[key] = value;
	return value;
}

func TrHash_delete(vm *RubyVM, self, key *RubyObject) RubyObject {
	if !self.(Hash) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected Hash"));
		return TR_UNDEF;
	}
	if value, ok := self.hash[key]; ok {
		hash[key] = 0, false;	//	deletes the value from the map
		return value;
	} else {
		return TR_NIL;
	}
}

func TrHash_init(vm *RubyVM) {
	c := vm.classes[TR_T_Hash] = Object_const_set(vm, vm.self, TrSymbol_new(vm, Hash), newClass(vm, TrSymbol_new(vm, Hash), vm.classes[TR_T_Object]));
	c.add_method(vm, TrSymbol_new(vm, "length"), newMethod(vm, (TrFunc *)TrHash_size, TR_NIL, 0));
	c.add_method(vm, TrSymbol_new(vm, "size"), newMethod(vm, (TrFunc *)TrHash_size, TR_NIL, 0));
	c.add_method(vm, TrSymbol_new(vm, "[]"), newMethod(vm, (TrFunc *)TrHash_get, TR_NIL, 1));
	c.add_method(vm, TrSymbol_new(vm, "[]="), newMethod(vm, (TrFunc *)TrHash_set, TR_NIL, 2));
	c.add_method(vm, TrSymbol_new(vm, "delete"), newMethod(vm, (TrFunc *)TrHash_delete, TR_NIL, 1));
}