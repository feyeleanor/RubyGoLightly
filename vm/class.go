import (
	"tr";
)

type Method struct {
	type			TR_T;
	class			*RubyObject;
	ivars			map[string] RubyObject;
  	func			*TrFunc;
	data			*RubyObject;
	name			*RubyObject;
	arity			int;
}

type Module struct {
	type			TR_T;
	class			*RubyObject;
	ivars			map[string] RubyObject;
	name			*RubyObject;
	super			*RubyObject;
	methods			map[string] RubyObject;
	meta			bool;
}

func (vm *RubyVM) newModule(name *RubyObject) RubyObject {
	return Module{type: TR_T_Module, class: vm.classes[TR_T_Module], ivars: make(map[string] RubyObject), name: name, methods: make(map[string] RubyObject)};
}

func (vm *RubyVM) newIncludedModule(module, super *RubyObject) RubyObject {
	if !module.(Class) && !module.(Module) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + module));
		return TR_UNDEF;
	}
	m := Class *(module);
	return Module{type: TR_T_Module, class: vm.classes[TR_T_Module], ivars: make(map[string] RubyObject), name: m.name, methods: m.methods, super: super};
}

type Class struct {
	module			Module;
}

func (self *Module) instance_method(vm *RubyVM, name *RubyObject) RubyObject {
	if !self.(Class) && !self.(Module) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + self));
		return TR_UNDEF;
	}
	class := Class *(self);
	while (class) {
		if method := class.methods[name] { return method; }
		class = Class *(class).super;
	}
	return TR_NIL;
}

func (self *Module) add_method(vm *RubyVM, name, method *RubyObject) RubyObject {
	if !self.(Class) && !self.(Module) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + self));
		return TR_UNDEF;
	}
	m := Class *(self);
	m.methods[name] = method;
	method.name = name;
	return method;
}

func (self *Module) alias_method(vm *RubyVM, new_name, old_name *RubyObject) RubyObject {
	return self.instance_method(vm, old_name).add_method(vm, self, new_name);
}

func (self *Module) include(vm *RubyVM, module *RubyObject) RubyObject {
	if !self.(Class) && !self.(Module) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + self));
		return TR_UNDEF;
	}
	class := Class *(self);
	class.super = vm.newIncludedModule(module, class.super);
	return module;
}

func (self *Module) name(vm *RubyVM) RubyObject {
	if !self.(Class) && !self.(Module) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + self));
		return TR_UNDEF;
	}
	return Class *(self).name;
}

func TrModule_init(vm *RubyVM) {
	c := vm.classes[TR_T_Module] = Object_const_set(vm, vm.self, TrSymbol_new(vm, Module), newClass(vm, TrSymbol_new(vm, Module), vm.classes[TR_T_Object]));
	c.add_method(vm, TrSymbol_new(vm, "name"), newMethod(vm, (TrFunc *)TrModule_name, TR_NIL, 0));
	c.add_method(vm, TrSymbol_new(vm, "include"), newMethod(vm, (TrFunc *)TrModule_include, TR_NIL, 1));
	c.add_method(vm, TrSymbol_new(vm, "instance_method"), newMethod(vm, (TrFunc *)TrModule_instance_method, TR_NIL, 1));
	c.add_method(vm, TrSymbol_new(vm, "alias_method"), newMethod(vm, (TrFunc *)TrModule_instance_method, TR_NIL, 2));
	c.add_method(vm, TrSymbol_new(vm, "to_s"), newMethod(vm, (TrFunc *)TrModule_name, TR_NIL, 0));
}

/* class */

func newClass(vm *RubyVM, name, super *RubyObject) RubyObject {
	c := Class{type: TR_T_Class, class: vm.classes[TR_T_Class], ivars: make(map[string] RubyObject), name: name, methods: make(map[string] RubyObject), meta: false};
	if !super.(Class) && !super.(Module) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + super));
		return TR_UNDEF;
	}
 	// if VM is booting, those might not be set
	if (super && Class *(super).class) { c.class = newMetaClass(vm, Class *(super).class); }
	c.super = super
	return c;
}

func (self *Class) allocate(vm *RubyVM) RubyObject {
	return Object{type: TR_T_Object, class: vm.classes[TR_T_Object], ivars: make(map[string] RubyObject), class: self};
}

func (self *Class) superclass(vm *RubyVM) RubyObject {
	if !self.(Class) && !self.(Module) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + self));
		return TR_UNDEF;
	}
	super := Class *(self).super;
	while (super && !super.(Class)) {
		if !super.(Module) {
			vm.throw_reason = TR_THROW_EXCEPTION;
			vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + super));
			return TR_UNDEF;
		}
		super = Class *(super).super;
	}
	return super;
}

func TrClass_init(vm *RubyVM) {
	c := vm.classes[TR_T_Class] = Object_const_set(vm, vm.self, TrSymbol_new(vm, Class), newClass(vm, TrSymbol_new(vm, Class), vm.classes[TR_T_Module]));
	c.add_method(vm, TrSymbol_new(vm, "superclass"), newMethod(vm, (TrFunc *)TrClass_superclass, TR_NIL, 0));
	c.add_method(vm, TrSymbol_new(vm, "allocate"), newMethod(vm, (TrFunc *)TrClass_allocate, TR_NIL, 0));
}

/* metaclass */

func newMetaClass(vm *RubyVM, super *RubyObject) RubyObject {
	if !super.(Class) && !super.(Module) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + super));
		return TR_UNDEF;
	}
	class := Class *(super);
	if !class.name.(String) && !class.name.(Symbol) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + class.name));
		return TR_UNDEF;
	}
	name := tr_sprintf(vm, "Class:%s", class.name.ptr);
	if !name.(String) && !name.(Symbol) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + name));
		return TR_UNDEF;
	}
	name := TrSymbol_new(vm, name.ptr); /* symbolize */
	mc = newClass(vm, name, 0);
	mc.super = super;
	mc.meta = 1;
	return mc;
}

/* method */

func newMethod(vm *RubyVM, function *TrFunc, data *RubyObject, arity int) RubyObject {
	return Method{type: TR_T_Method, class: vm.classes[TR_T_Method], ivars: make(map[string] RubyObject), func: function, data: data, arity: arity};
}

func (self *Method) name(vm *RubyVM) RubyObject { return ((Method *) self).name; }
func (self *Method) arity(vm *RubyVM) RubyObject { return TR_INT2FIX(((Method *) self).arity); }

func (self *Method) dump(vm *RubyVM) RubyObject {
	m := (Method *) self;
	if !m.name.(String) && !m.name.(Symbol) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + m.name));
		return TR_UNDEF;
	}
	if m.name { println("<Method '%s':%p>", m.name.ptr, m); }
	if m.data {
		Block *(m.data).dump(vm, 0);
	} else {
		fmt.Println("<CFunction:%p>", m.func);
	}
	return TR_NIL;
}

func TrMethod_init(vm *RubyVM) {
	c := vm.classes[TR_T_Method] = Object_const_set(vm, vm.self, TrSymbol_new(vm, Method), newClass(vm, TrSymbol_new(vm, Method), vm.classes[TR_T_Object]));
	c.add_method(vm, TrSymbol_new(vm, "name"), newMethod(vm, (TrFunc *)TrMethod_name, TR_NIL, 0));
	c.add_method(vm, TrSymbol_new(vm, "arity"), newMethod(vm, (TrFunc *)TrMethod_arity, TR_NIL, 0));
	c.add_method(vm, TrSymbol_new(vm, "dump"), newMethod(vm, (TrFunc *)TrMethod_dump, TR_NIL, 0));
}