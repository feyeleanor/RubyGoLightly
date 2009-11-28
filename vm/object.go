import (
	"tr";
	"call";
)

type Object struct {
	type 			TR_T;
	class			*RubyObject;
	ivars			map[string] RubyObject;
}

func Object_alloc(vm *RubyVM, class *RubyObject) RubyObject {
	return Object{type: TR_T_Object, class: vm.classes[TR_T_Object], ivars: make(map[string] RubyObject), class: class};
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
	vm.throw_reason = TR_THROW_EXCEPTION;
	vm.throw_value = TrException_new(vm, vm.cNoMethodError, tr_sprintf(vm, "Method not found: `%s'", argv[0].ptr));
	return TR_UNDEF;
}

func Object_send(vm *RubyVM, self *RubyObject, argc int, argv []RubyObject) RubyObject {
	if argc == 0 {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cArgumentError, tr_sprintf(vm, "wrong number of arguments (%d for 1)", argc));
		return TR_UNDEF;
	}
	method := Object_method(vm, self, argv[0]);
	if method == TR_NIL {
		method = Object_method(vm, self, TrSymbol_new(vm, "method_missing"));
		return method.call(vm, self, argc, argv, 0, 0);
	} else {
		return method.call(vm, self, argc-1, argv+1, 0, 0);
	}
}

// TODO respect namespace
func Object_const_get(vm *RubyVM, self, name *RubyObject) RubyObject {
	return vm.consts[name] || TR_NIL;
}

func Object_const_set(vm *RubyVM, self, name, value *RubyObject) RubyObject {
	vm.consts[name] = value;
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
	class_name := Object_send(vm, Object_send(vm, self, 1, { TrSymbol_new(vm, "class") }), 1, { TrSymbol_new(vm, "name") });
	if !class_name.(String) && !class_name.(Symbol) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + class_name));
		return TR_UNDEF;
	}
	return tr_sprintf(vm, "#<%s:%p>", class_name.ptr, self);
}

func Object_preinit(vm *RubyVM) {
	return vm.classes[TR_T_Object] = Object_const_set(vm, vm.self, TrSymbol_new(vm, Object), newClass(vm, TrSymbol_new(vm, Object), vm.classes[TR_T_Object]));
}

func Object_init(vm *RubyVM) {
	c := vm.classes[TR_T_Object];
	c.add_method(vm, TrSymbol_new(vm, "class"), newMethod(vm, (TrFunc *)Object_class, TR_NIL, 0));
	c.add_method(vm, TrSymbol_new(vm, "method"), newMethod(vm, (TrFunc *)Object_method, TR_NIL, 1));
	c.add_method(vm, TrSymbol_new(vm, "method_missing"), newMethod(vm, (TrFunc *)Object_method_missing, TR_NIL, -1));
	c.add_method(vm, TrSymbol_new(vm, "send"), newMethod(vm, (TrFunc *)Object_send, TR_NIL, -1));
	c.add_method(vm, TrSymbol_new(vm, "object_id"), newMethod(vm, (TrFunc *)Object_object_id, TR_NIL, 0));
	c.add_method(vm, TrSymbol_new(vm, "instance_eval"), newMethod(vm, (TrFunc *)Object_instance_eval, TR_NIL, 1));
	c.add_method(vm, TrSymbol_new(vm, "to_s"), newMethod(vm, (TrFunc *)Object_inspect, TR_NIL, 0));
	c.add_method(vm, TrSymbol_new(vm, "inspect"), newMethod(vm, (TrFunc *)Object_inspect, TR_NIL, 0));
}