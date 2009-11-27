import(
	"tr";
	"internal";
	)

func TrHash_new(vm *RubyVM) RubyObject {
	return Hash{type: TR_T_Hash, class: vm.classes[TR_T_Hash], ivars: kh_init(RubyObject), kh: kh_init(RubyObject)};
}

func TrHash_new2(vm *RubyVM, n size_t, items []RubyObject) RubyObject {
  h := TrHash_new(vm);
  i size_t;
  ret int;
  for (i = 0; i < n; i+=2) {
    k := kh_put(RubyObject, h.kh, items[i], &ret);
    kh_value(h.kh, k) = items[i+1];
  }
  return h;
}

func TrHash_size(vm *RubyVM, self *RubyObject) RubyObject {
	if !self.(Hash) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected Hash"));
		return TR_UNDEF;
	}
	h := TrHash *(self);
	return TR_INT2FIX(kh_size(h.kh));
}

// TODO use Object#hash as the key
func TrHash_get(vm *RubyVM, self, key *RubyObject) RubyObject {
	if !self.(Hash) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected Hash"));
		return TR_UNDEF;
	}
	h := TrHash *(self);
	k := kh_get(RubyObject, h.kh, key);
	if (k != kh_end(h.kh)) return kh_value(h.kh, k);
	return TR_NIL;
}

func TrHash_set(vm *RubyVM, self, key, value *RubyObject) RubyObject {
	if !self.(Hash) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected Hash"));
		return TR_UNDEF;
	}
	h := TrHash *(self);
	ret int;
	k := kh_put(RubyObject, h.kh, key, &ret);
	if (!ret) kh_del(RubyObject, h.kh, k);
	kh_value(h.kh, k) = value;
	return value;
}

func TrHash_delete(vm *RubyVM, self, key *RubyObject) RubyObject {
	if !self.(Hash) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected Hash"));
		return TR_UNDEF;
	}
	h := TrHash *(self);
	k := kh_get(RubyObject, h.kh, key);
	if (k != kh_end(h.kh)) {
		value := kh_value(h.kh, k);
		kh_del(RubyObject, h.kh, k);
		return value;
	}
	return TR_NIL;
}

func TrHash_init(vm *RubyVM) {
	c := vm.classes[TR_T_Hash] = Object_const_set(vm, vm.self, tr_intern(Hash), newClass(vm, tr_intern(Hash), vm.classes[TR_T_Object]));
	tr_def(c, "length", TrHash_size, 0);
	tr_def(c, "size", TrHash_size, 0);
	tr_def(c, "[]", TrHash_get, 1);
	tr_def(c, "[]=", TrHash_set, 2);
	tr_def(c, "delete", TrHash_delete, 1);
}