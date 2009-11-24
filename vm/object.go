import (
	"tr";
	"internal";
	"call";
)

OBJ TrObject_alloc(vm *struct TrVM, OBJ class) {
  TrObject *o = TR_INIT_CORE_OBJECT(Object);
  if (class) o.class = class;
  return OBJ(o);
}

int TrObject_type(vm *struct TrVM, OBJ obj) {
	switch (obj) {
		case TR_NIL: return TR_T_NilClass;
		case TR_TRUE: return TR_T_TrueClass;
		case TR_FALSE: return TR_T_FalseClass;
	}
	if TR_IS_FIX(obj) { return TR_T_Fixnum }
	return TR_COBJECT(obj).type;
}

OBJ TrObject_method(vm *struct TrVM, OBJ self, OBJ name) {
  return TR_CLASS(self).instance_method(vm, name);
}

OBJ TrObject_method_missing(vm *struct TrVM, OBJ self, int argc, OBJ argv[]) {
	assert(argc > 0);
	tr_raise(NoMethodError, "Method not found: `%s'", TR_STR_PTR(argv[0]));
}

OBJ TrObject_send(vm *struct TrVM, OBJ self, int argc, OBJ argv[]) {
  if argc == 0 { tr_raise(ArgumentError, "wrong number of arguments (%d for 1)", argc); }
  OBJ method = TrObject_method(vm, self, argv[0]);
  if method == TR_NIL {
    method = TrObject_method(vm, self, tr_intern("method_missing"));
    return method.call(vm, self, argc, argv, 0, 0);
  } else {
    return method.call(vm, self, argc-1, argv+1, 0, 0);
  }
}

/* TODO respect namespace */
OBJ TrObject_const_get(vm *struct TrVM, OBJ self, OBJ name) {
	khiter_t k = kh_get(OBJ, vm.consts, name);
	if (k != kh_end(vm.consts)) return kh_value(vm.consts, k);
	return TR_NIL;
}

OBJ TrObject_const_set(vm *struct TrVM, OBJ self, OBJ name, OBJ value) {
	int ret;
	khiter_t k = kh_put(OBJ, vm.consts, name, &ret);
	if (!ret) kh_del(OBJ, vm.consts, k);
	kh_value(vm.consts, k) = value;
	return value;
}

OBJ TrObject_add_singleton_method(vm *struct TrVM, OBJ self, OBJ name, OBJ method) {
  TrObject *o = TR_COBJECT(self);
  if (!TR_CCLASS(o.class).meta)
    o.class = newMetaClass(vm, o.class);
  assert(TR_CCLASS(o.class).meta && "first class must be the metaclass");
  o.class.add_method(vm, name, method);
  return method;
}

static OBJ TrObject_class(vm *struct TrVM, OBJ self) {
  OBJ class = TR_CLASS(self);
  /* find the first non-metaclass */
  while (class && (!class.(Class) || TR_CCLASS(class).meta))
    class = TR_CCLASS(class).super;
  assert(class && "classless object");
  return class;
}

static OBJ TrObject_object_id(vm *struct TrVM, OBJ self) {
	return TR_INT2FIX((int)&self);
}

static OBJ TrObject_instance_eval(vm *struct TrVM, OBJ self, OBJ code) {
  Block *b = Block_compile(vm, TR_STR_PTR(code), "<eval>", 0);
  if (!b) return TR_UNDEF;
  return TrVM_run(vm, b, self, TR_COBJECT(self).class, 0, 0);
}

static OBJ TrObject_inspect(vm *struct TrVM, OBJ self) {
  const char *name;
  name = TR_STR_PTR(tr_send2(tr_send2(self, "class"), "name"));
  return tr_sprintf(vm, "#<%s:%p>", name, (void*)self);
}

void TrObject_preinit(vm *struct TrVM) {
  TR_INIT_CORE_CLASS(Object, /* ignored */ Object);
}

void TrObject_init(vm *struct TrVM) {
  OBJ c = TR_CORE_CLASS(Object);
  tr_def(c, "class", TrObject_class, 0);
  tr_def(c, "method", TrObject_method, 1);
  tr_def(c, "method_missing", TrObject_method_missing, -1);
  tr_def(c, "send", TrObject_send, -1);
  tr_def(c, "object_id", TrObject_object_id, 0);
  tr_def(c, "instance_eval", TrObject_instance_eval, 1);
  tr_def(c, "to_s", TrObject_inspect, 0);
  tr_def(c, "inspect", TrObject_inspect, 0);
}