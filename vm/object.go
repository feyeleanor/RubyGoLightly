import (
	"tr";
	"internal";
	"call";
)

func TrObject_alloc(vm *RubyVM, class OBJ) OBJ {
	o := TR_INIT_CORE_OBJECT(Object);
	if class { o.class = class; }
	return OBJ(o);
}

int TrObject_type(vm *RubyVM, OBJ obj) {
	switch (obj) {
		case TR_NIL: return TR_T_NilClass;
		case TR_TRUE: return TR_T_TrueClass;
		case TR_FALSE: return TR_T_FalseClass;
	}
	if TR_IS_FIX(obj) { return TR_T_Fixnum }
	return TR_COBJECT(obj).type;
}

func TrObject_method(vm *RubyVM, self OBJ, name OBJ) OBJ {
	return TR_CLASS(self).instance_method(vm, name);
}

func TrObject_method_missing(vm *RubyVM, self OBJ, argc int, argv []OBJ) OBJ {
	assert(argc > 0);
	tr_raise(NoMethodError, "Method not found: `%s'", TR_STR_PTR(argv[0]));
}

func TrObject_send(vm *RubyVM, self OBJ, argc int, argv []OBJ) OBJ {
	if argc == 0 { tr_raise(ArgumentError, "wrong number of arguments (%d for 1)", argc); }
	method := TrObject_method(vm, self, argv[0]);
	if method == TR_NIL {
		method = TrObject_method(vm, self, tr_intern("method_missing"));
		return method.call(vm, self, argc, argv, 0, 0);
	} else {
		return method.call(vm, self, argc-1, argv+1, 0, 0);
	}
}

// TODO respect namespace
func TrObject_const_get(vm *RubyVM, OBJ self, OBJ name) OBJ {
	k := kh_get(OBJ, vm.consts, name);
	if (k != kh_end(vm.consts)) return kh_value(vm.consts, k);
	return TR_NIL;
}

func TrObject_const_set(vm *RubyVM, OBJ self, OBJ name, OBJ value) OBJ {
	int ret;
	k := kh_put(OBJ, vm.consts, name, &ret);
	if (!ret) kh_del(OBJ, vm.consts, k);
	kh_value(vm.consts, k) = value;
	return value;
}

func TrObject_add_singleton_method(vm *RubyVM, OBJ self, OBJ name, OBJ method) OBJ {
	object := *TR_COBJECT(self);
	if (!TR_CCLASS(object.class).meta) { object.class := newMetaClass(vm, object.class); }
	assert(TR_CCLASS(object.class).meta && "first class must be the metaclass");
	object.class.add_method(vm, name, method);
	return method;
}

func TrObject_class(vm *RubyVM, OBJ self) OBJ {
	class := TR_CLASS(self);
	// find the first non-metaclass
	while (class && (!class.(Class) || TR_CCLASS(class).meta)) { class = TR_CCLASS(class).super; }
	assert(class && "classless object");
	return class;
}

func TrObject_object_id(vm *RubyVM, OBJ self) OBJ {
	return TR_INT2FIX((int)&self);
}

func TrObject_instance_eval(vm *RubyVM, OBJ self, OBJ code) OBJ {
	if block := Block_compile(vm, TR_STR_PTR(code), "<eval>", 0) {
		return vm.run(block, self, TR_COBJECT(self).class, nil);
	} else {
		return TR_UNDEF;
	}
}

func TrObject_inspect(vm *RubyVM, OBJ self) OBJ {
	name := TR_STR_PTR(tr_send2(tr_send2(self, "class"), "name"));
	return tr_sprintf(vm, "#<%s:%p>", name, (void*)self);
}

func TrObject_preinit(vm *RubyVM) {
	TR_INIT_CORE_CLASS(Object, /* ignored */ Object);
}

func TrObject_init(vm *RubyVM) {
	c := TR_CORE_CLASS(Object);
	tr_def(c, "class", TrObject_class, 0);
	tr_def(c, "method", TrObject_method, 1);
	tr_def(c, "method_missing", TrObject_method_missing, -1);
	tr_def(c, "send", TrObject_send, -1);
	tr_def(c, "object_id", TrObject_object_id, 0);
	tr_def(c, "instance_eval", TrObject_instance_eval, 1);
	tr_def(c, "to_s", TrObject_inspect, 0);
	tr_def(c, "inspect", TrObject_inspect, 0);
}