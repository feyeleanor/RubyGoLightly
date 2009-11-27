import (
	"tr";
	"internal";
)

type Method struct {
	type			TR_T;
	class			*RubyObject;
	ivars			*map[string] RubyObject;
  	func			*TrFunc;
	data			*RubyObject;
	name			*RubyObject;
	arity			int;
}

type Module struct {
	type			TR_T;
	class			*RubyObject;
	ivars			*map[string] *RubyObject;
	name			*RubyObject;
	super			*RubyObject;
	methods			map[string] RubyObject;
	meta:1			int;
}

type Class struct {
	module			Module;
}


/* included module proxy */

func newIModule(vm *RubyVM, module, super *RubyObject) RubyObject {
	m := TR_CCLASS(module);
	return Module{type: TR_T_Module, class: vm.classes[TR_T_Module], ivars: kh_init(RubyObject), name: m.name, methods: m.methods, super: super};
}

/* module */

func newModule(vm *RubyVM, name *RubyObject) RubyObject {
	return Module{type: TR_T_Module, class: vm.classes[TR_T_Module], ivars: kh_init(RubyObject), name: name, methods: kh_init(RubyObject), meta: false};
}

#define TR_CCLASS(X)         ((X.(Class) || X.(Module) ? 0 : TR_TYPE_ERROR(T)), (Class *)(X))
func (self *Class) cclass() *Class {
	if self.(Class) || self.(Module) {
		nil
	} else {
		TR_TYPE_ERROR(T)
	}
	self
}

func (self *Module) instance_method(vm *RubyVM, name *RubyObject) RubyObject {
  Class *class = TR_CCLASS(self);
  while (class) {
    method := TR_KH_GET(class.methods, name);
    if (method) return method;
    class = (Class *)class.super;
  }
  return TR_NIL;
}

func (self *Module) add_method(vm *RubyVM, name, method *RubyObject) RubyObject {
  Class *m = TR_CCLASS(self);
  TR_KH_SET(m.methods, name, method);
  ((Method *) method).name = name;
  return method;
}

func (self *Module) alias_method(vm *RubyVM, new_name, old_name *RubyObject) RubyObject {
  return self.instance_method(vm, old_name).add_method(vm, self, new_name);
}

func (self *Module) include(vm *RubyVM, module *RubyObject) RubyObject {
  Class *class = TR_CCLASS(self);
  class.super = newIModule(vm, module, class.super);
  return module;
}

func (self *Module) name(vm *RubyVM) RubyObject {
  return TR_CCLASS(self).name;
}

void TrModule_init(vm *RubyVM) {
  c := vm.classes[TR_T_Module] = Object_const_set(vm, vm.self, tr_intern(Module), newClass(vm, tr_intern(Module), vm.classes[TR_T_Object]));
  tr_def(c, "name", TrModule_name, 0);
  tr_def(c, "include", TrModule_include, 1);
  tr_def(c, "instance_method", TrModule_instance_method, 1);
  tr_def(c, "alias_method", TrModule_alias_method, 2);
  tr_def(c, "to_s", TrModule_name, 0);
}

/* class */

func newClass(vm *RubyVM, name, super *RubyObject) RubyObject {
	c := Class{type: TR_T_Class, class: vm.classes[TR_T_Class], ivars: kh_init(RubyObject), name: name, methods: kh_init(RubyObject), meta: false};

 	// if VM is booting, those might not be set
	if (super && TR_CCLASS(super).class) { c.class = newMetaClass(vm, TR_CCLASS(super).class); }
	c.super = super
	return c;
}

func (self *Class) allocate(vm *RubyVM) RubyObject {
	return Object{type: TR_T_Object, class: vm.classes[TR_T_Object], ivars: kh_init(RubyObject), class: self};
}

func (self *Class) superclass(vm *RubyVM) RubyObject {
  super := TR_CCLASS(self).super;
  while (super && !super.(Class))
    super = TR_CCLASS(super).super;
  return super;
}

void TrClass_init(vm *RubyVM) {
  c := vm.classes[TR_T_Class] = Object_const_set(vm, vm.self, tr_intern(Class), newClass(vm, tr_intern(Class), vm.classes[TR_T_Module]));
  tr_def(c, "superclass", TrClass_superclass, 0);
  tr_def(c, "allocate", TrClass_allocate, 0);
}

/* metaclass */

func newMetaClass(vm *RubyVM, super *RubyObject) RubyObject {
  *c = TR_CCLASS(super);
  name := tr_sprintf(vm, "Class:%s", TR_STR_PTR(c.name));
  name := tr_intern(TR_STR_PTR(name)); /* symbolize */
  mc = newClass(vm, name, 0);
  mc.super = super;
  mc.meta = 1;
  return mc;
}

/* method */

func newMethod(vm *RubyVM, function *TrFunc, data *RubyObject, arity int) RubyObject {
	return Method{type: TR_T_Method, class: vm.classes[TR_T_Method], ivars: kh_init(RubyObject), func: function, data: data, arity: arity};
}

func (self *Method) name(vm *RubyVM) RubyObject { return ((Method *) self).name; }
func (self *Method) arity(vm *RubyVM) RubyObject { return TR_INT2FIX(((Method *) self).arity); }

func (self *Method) dump(vm *RubyVM) RubyObject {
	Method *m = (Method *) self;
	if (m.name) printf("<Method '%s':%p>\n", TR_STR_PTR(m.name), m);
	if (m.data) {
		(Block*)m.data.dump(vm);
	} else {
		fmt.Println("<CFunction:%p>", m.func);
	}
	return TR_NIL;
}

void TrMethod_init(vm *RubyVM) {
  c := vm.classes[TR_T_Method] = Object_const_set(vm, vm.self, tr_intern(Method), newClass(vm, tr_intern(Method), vm.classes[TR_T_Object]));
  tr_def(c, "name", TrMethod_name, 0);
  tr_def(c, "arity", TrMethod_arity, 0);
  tr_def(c, "dump", TrMethod_dump, 0);
}