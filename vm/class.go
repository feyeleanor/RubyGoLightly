import (
	"tr";
	"internal";
)

type Method struct {
	type			TR_T;
	class			OBJ;
	ivars			*khash_t(OBJ);
  	func			*TrFunc;
	data			OBJ;
	name			OBJ;
	arity			int;
}

type Class struct {
	type			TR_T;
	class			OBJ;
	ivars			*khash_t(OBJ);
	name			OBJ;
	super			OBJ;
	methods			*khash_t(OBJ);
	meta:1			int;
}

type Module Class;

/* included module proxy */

func newIModule(vm *struct TrVM, OBJ module, OBJ super) OBJ {
  Module *m = TR_CCLASS(module);
  Module *im = TR_INIT_CORE_OBJECT(Module);
  im.name = m.name;
  im.methods = m.methods;
  im.super = super;
  return (OBJ)im;
}

/* module */

func newModule(vm *struct TrVM, OBJ name) OBJ {
  Module *m = TR_INIT_CORE_OBJECT(Module);
  m.name = name;
  m.methods = kh_init(OBJ);
  m.meta = 0;
  return (OBJ)m;
}

func (self *Module) instance_method(vm *TrVM, name OBJ) OBJ {
  Class *class = TR_CCLASS(self);
  while (class) {
    OBJ method = TR_KH_GET(class.methods, name);
    if (method) return method;
    class = (Class *)class.super;
  }
  return TR_NIL;
}

func (self *Module) add_method(vm *TrVM, name, method OBJ) OBJ {
  Class *m = TR_CCLASS(self);
  TR_KH_SET(m.methods, name, method);
  ((Method *) method).name = name;
  return method;
}

func (self *Module) alias_method(vm *TrVM, new_name, old_name OBJ) OBJ {
  return self.instance_method(vm, old_name).add_method(vm, self, new_name);
}

func (self *Module) include(vm *TrVM, module OBJ) OBJ {
  Class *class = TR_CCLASS(self);
  class.super = newIModule(vm, module, class.super);
  return module;
}

func (self *Module) name(vm *TrVM) OBJ {
  return TR_CCLASS(self).name;
}

void TrModule_init(vm *TrVM) {
  OBJ c = TR_INIT_CORE_CLASS(Module, Object);
  tr_def(c, "name", TrModule_name, 0);
  tr_def(c, "include", TrModule_include, 1);
  tr_def(c, "instance_method", TrModule_instance_method, 1);
  tr_def(c, "alias_method", TrModule_alias_method, 2);
  tr_def(c, "to_s", TrModule_name, 0);
}

/* class */

func newClass(vm *TrVM, OBJ name, OBJ super) OBJ {
  Class *c = TR_INIT_CORE_OBJECT(Class);
  c.name = name;
  c.methods = kh_init(OBJ);
  c.meta = 0;

  /* if VM is booting, those might not be set */
  if (super && TR_CCLASS(super).class) c.class = newMetaClass(vm, TR_CCLASS(super).class);
  c.super = super;
  return (OBJ)c;
}

func (self *Class) allocate(vm *TrVM) OBJ {
  TrObject *o = TR_INIT_CORE_OBJECT(Object);
  o.class = self;
  return (OBJ)o;
}

func (self *Class) superclass(vm *TrVM) OBJ {
  OBJ super = TR_CCLASS(self).super;
  while (super && !TR_IS_A(super, Class))
    super = TR_CCLASS(super).super;
  return super;
}

void TrClass_init(vm *struct TrVM) {
  OBJ c = TR_INIT_CORE_CLASS(Class, Module);
  tr_def(c, "superclass", TrClass_superclass, 0);
  tr_def(c, "allocate", TrClass_allocate, 0);
}

/* metaclass */

func newMetaClass(vm *TrVM, OBJ super) OBJ {
  *c = TR_CCLASS(super);
  OBJ name = tr_sprintf(vm, "Class:%s", TR_STR_PTR(c.name));
  name = tr_intern(TR_STR_PTR(name)); /* symbolize */
  mc = newClass(vm, name, 0);
  mc.super = super;
  mc.meta = 1;
  return (OBJ)mc;
}

/* method */

func newMethod(vm *TrVM, function *TrFunc, data OBJ, arity int) OBJ {
  Method *m = TR_INIT_CORE_OBJECT(Method);
  m.func = function;
  m.data = data;
  m.arity = arity;
  return (OBJ)m;
}

OBJ (self *Method) name(vm *TrVM) { return ((Method *) self).name; }
OBJ (self *Method) arity(vm *TrVM) { return TR_INT2FIX(((Method *) self).arity); }

func (self *Method) dump(vm *TrVM) OBJ {
	Method *m = (Method *) self;
	if (m.name) printf("<Method '%s':%p>\n", TR_STR_PTR(m.name), m);
	if (m.data) {
		(Block*)m.data.dump(vm);
	} else {
		fmt.Println("<CFunction:%p>", m.func);
	}
	return TR_NIL;
}

void TrMethod_init(vm *TrVM) {
  OBJ c = TR_INIT_CORE_CLASS(Method, Object);
  tr_def(c, "name", TrMethod_name, 0);
  tr_def(c, "arity", TrMethod_arity, 0);
  tr_def(c, "dump", TrMethod_dump, 0);
}