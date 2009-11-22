#include "tr.h"
#include "internal.h"

OBJ TrBinding_new(vm *struct TrVM, TrFrame *f) {
  TrBinding *b = TR_INIT_CORE_OBJECT(Binding);
  b.frame = f;
  return (OBJ)b;
}

void TrBinding_init(vm *struct TrVM) {
  TR_INIT_CORE_CLASS(Binding, Object);
}

static OBJ TrKernel_puts(vm *struct TrVM, OBJ self, int argc, OBJ argv[]) {
	int i;
	for (i = 0; i < argc; ++i) printf("%s\n", TR_STR_PTR(tr_send2(argv[i], "to_s")));
	return TR_NIL;
}

static OBJ TrKernel_binding(vm *struct TrVM, OBJ self) {
	return TrBinding_new(vm, vm.frame.previous);
}

static OBJ TrKernel_eval(vm *struct TrVM, OBJ self, int argc, OBJ argv[]) {
	if argc < 1 { tr_raise(ArgumentError, "string argument required") }
	if argc > 4 { tr_raise(ArgumentError, "Too much arguments") }
	OBJ string = argv[0];
	TrFrame *f = (argc > 1 && argv[1]) ? TR_CBINDING(argv[1]).frame : vm.frame;
	char *filename = (argc > 2 && argv[1]) ? TR_STR_PTR(argv[2]) : "<eval>";
	size_t lineno = argc > 3 ? TR_FIX2INT(argv[3]) : 0;
	Block *blk = Block_compile(vm, TR_STR_PTR(string), filename, lineno);
	if !blk { return TR_UNDEF }
	if vm.debug { blk.dump(vm) }
	return TrVM_run(vm, blk, f.self, f.class, kv_size(blk.locals), f.stack);
}

static OBJ TrKernel_load(vm *struct TrVM, OBJ self, OBJ filename) {
	return TrVM_load(vm, TR_STR_PTR(filename));
}

static OBJ TrKernel_raise(vm *struct TrVM, OBJ self, int argc, OBJ argv[]) {
	OBJ e = TR_NIL;
	switch (argc) {
		case 0:
			e = tr_getglobal("$!");

		case 1:
			if TR_IS_A(argv[0], String) {
				e = TrException_new(vm, vm.cRuntimeError, argv[0]);
			} else {
				e = tr_send2(argv[0], "exception");
			}

		case 2:
			e = tr_send2(argv[0], "exception", argv[1]);

		default:
			tr_raise(ArgumentError, "wrong number of arguments (%d for 2)", argc);
	}
	TrException_set_backtrace(vm, e, TrVM_backtrace(vm));
	TR_THROW(EXCEPTION, e);
}

void TrKernel_init(vm *struct TrVM) {
  OBJ m = tr_defmodule("Kernel");
  TrModule_include(vm, TR_CORE_CLASS(Object), m);
  tr_def(m, "puts", TrKernel_puts, -1);
  tr_def(m, "eval", TrKernel_eval, -1);
  tr_def(m, "load", TrKernel_load, 1);
  tr_def(m, "binding", TrKernel_binding, 0);
  tr_def(m, "raise", TrKernel_raise, -1);
}