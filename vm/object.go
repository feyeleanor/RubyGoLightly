import (
	"tr";
	"internal";
	"call";
)

type Object struct {
	type 			TR_T;
	class			*RubyObject;
	ivars			*map[string] *RubyObject;
}

func Object_alloc(vm *RubyVM, class *RubyObject) RubyObject {
	return Object{type: TR_T_Object, class: vm.classes[TR_T_Object], ivars: kh_init(RubyObject), class: class};
}

func Object_type(vm *RubyVM, obj *RubyObject) int {
	switch (obj) {
		case TR_NIL: return TR_T_NilClass;
		case TR_TRUE: return TR_T_TrueClass;
		case TR_FALSE: return TR_T_FalseClass;
	}
	if TR_IS_FIX(obj) { return TR_T_Fixnum }
	return Object *(obj).type;
}

func Object_method(vm *RubyVM, self, name *RubyObject) RubyObject {
	if TR_IMMEDIATE(self) {
		class := vm.classes[Object_type(vm, self)];
	} else {
		Object *(self).class;
	}
	return class.instance_method(vm, name);
}

func Object_method_missing(vm *RubyVM, self *RubyObject, argc int, argv []RubyObject) RubyObject {
	assert(argc > 0);
	if !argv[0].(String) && !argv[0].(Symbol) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + argv[0]));
		return TR_UNDEF;
	}
	tr_raise(NoMethodError, "Method not found: `%s'", argv[0].ptr);
}

func Object_send(vm *RubyVM, self *RubyObject, argc int, argv []RubyObject) RubyObject {
	if argc == 0 { tr_raise(ArgumentError, "wrong number of arguments (%d for 1)", argc); }
	method := Object_method(vm, self, argv[0]);
	if method == TR_NIL {
		method = Object_method(vm, self, tr_intern("method_missing"));
		return method.call(vm, self, argc, argv, 0, 0);
	} else {
		return method.call(vm, self, argc-1, argv+1, 0, 0);
	}
}

// TODO respect namespace
func Object_const_get(vm *RubyVM, self, name *RubyObject) RubyObject {
	k := kh_get(RubyObject, vm.consts, name);
	if (k != kh_end(vm.consts)) return kh_value(vm.consts, k);
	return TR_NIL;
}

func Object_const_set(vm *RubyVM, self, name, value *RubyObject) RubyObject {
	int ret;
	k := kh_put(RubyObject, vm.consts, name, &ret);
	if (!ret) kh_del(RubyObject, vm.consts, k);
	kh_value(vm.consts, k) = value;
	return value;
}

func Object_add_singleton_method(vm *RubyVM, self, name, method *RubyObject) RubyObject {
	object := Object *(self);
	if !object.(Class) && !object.(Module) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + object));
		return TR_UNDEF;
	}
	if !object.class.meta { object.class := newMetaClass(vm, object.class); }
	assert(object.class.meta && "first class must be the metaclass");
	object.class.add_method(vm, name, method);
	return method;
}

func Object_class(vm *RubyVM, self *RubyObject) RubyObject {
	if TR_IMMEDIATE(self) {
		class := vm.classes[Object_type(vm, (self))];
	} else {
		class := Object *(self).class;
	}
	// find the first non-metaclass
	if !class.(Class) && !class.(Module) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + class));
		return TR_UNDEF;
	}
	while (class && (!class.(Class) || class.meta)) { class = class.super; }
	assert(class && "classless object");
	return class;
}

func Object_object_id(vm *RubyVM, self *RubyObject) RubyObject {
	return TR_INT2FIX((int)&self);
}

func Object_instance_eval(vm *RubyVM, self, code *RubyObject) RubyObject {
	if !code.(String) && !code.(Symbol) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + code));
		return TR_UNDEF;
	}
	if block := Block_compile(vm, code.ptr, "<eval>", 0) {
		return vm.run(block, self, Object *(self).class, nil);
	} else {
		return TR_UNDEF;
	}
}

func Object_inspect(vm *RubyVM, self *RubyObject) RubyObject {
	class_name := tr_send2(tr_send2(self, "class"), "name")
	if !class_name.(String) && !class_name.(Symbol) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + class_name));
		return TR_UNDEF;
	}
	return tr_sprintf(vm, "#<%s:%p>", class_name.ptr, self);
}

func Object_preinit(vm *RubyVM) {
	return vm.classes[TR_T_Object] = Object_const_set(vm, vm.self, tr_intern(Object), newClass(vm, tr_intern(Object), vm.classes[TR_T_Object]));
}

func Object_init(vm *RubyVM) {
	c := vm.classes[TR_T_Object];
	tr_def(c, "class", Object_class, 0);
	tr_def(c, "method", Object_method, 1);
	tr_def(c, "method_missing", Object_method_missing, -1);
	tr_def(c, "send", Object_send, -1);
	tr_def(c, "object_id", Object_object_id, 0);
	tr_def(c, "instance_eval", Object_instance_eval, 1);
	tr_def(c, "to_s", Object_inspect, 0);
	tr_def(c, "inspect", Object_inspect, 0);
}