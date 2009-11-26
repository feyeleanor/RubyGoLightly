#include "tr.h"
#include "internal.h"

func TrBinding_new(vm *RubyVM, f *Frame) OBJ {
  TrBinding *b = TR_INIT_CORE_OBJECT(Binding);
  b.frame = f;
  return OBJ(b);
}

void TrBinding_init(vm *RubyVM) {
  TR_INIT_CORE_CLASS(Binding, Object);
}

func TrKernel_puts(vm *RubyVM, self OBJ, argc int, argv []OBJ) OBJ {
	for object := range argv { fmt.println("%s", TR_STR_PTR(tr_send2(object, "to_s"))); }
	return TR_NIL;
}

func TrKernel_binding(vm *RubyVM, self OBJ) OBJ {
	return TrBinding_new(vm, vm.frame.previous);
}

func TrKernel_eval(vm *RubyVM, self OBJ, argc int, argv []OBJ) OBJ {
	if argc < 1 { tr_raise(ArgumentError, "string argument required") }
	if argc > 4 { tr_raise(ArgumentError, "Too much arguments") }
	code_string := argv[0];
	frame := (argc > 1 && argv[1]) ? TR_CBINDING(argv[1]).frame : vm.frame;
	filename := (argc > 2 && argv[1]) ? TR_STR_PTR(argv[2]) : "<eval>";
	lineno := argc > 3 ? TR_FIX2INT(argv[3]) : 0;
	blk := Block_compile(vm, TR_STR_PTR(code_string), filename, lineno);
	if !blk { return TR_UNDEF }
	if vm.debug { blk.dump(vm) }
	return vm.run(blk, frame.self, frame.class, frame.stack[0:blk.locals.Len() - 1]);
}

func TrKernel_load(vm *RubyVM, self, filename OBJ) OBJ {
	return vm.load(TR_STR_PTR(filename));
}

static OBJ TrKernel_raise(vm *RubyVM, OBJ self, int argc, OBJ argv[]) {
	OBJ e = TR_NIL;
	switch (argc) {
		case 0:
			e = tr_getglobal("$!");

		case 1:
			if argv[0].(String) {
				e = TrException_new(vm, vm.cRuntimeError, argv[0]);
			} else {
				e = tr_send2(argv[0], "exception");
			}

		case 2:
			e = tr_send2(argv[0], "exception", argv[1]);

		default:
			tr_raise(ArgumentError, "wrong number of arguments (%d for 2)", argc);
	}
	TrException_set_backtrace(vm, e, vm.backtrace());
	TR_THROW(EXCEPTION, e);
}

void TrKernel_init(vm *RubyVM) {
  OBJ m = tr_defmodule("Kernel");
  TR_CORE_CLASS(Object).include(vm, m);
  tr_def(m, "puts", TrKernel_puts, -1);
  tr_def(m, "eval", TrKernel_eval, -1);
  tr_def(m, "load", TrKernel_load, 1);
  tr_def(m, "binding", TrKernel_binding, 0);
  tr_def(m, "raise", TrKernel_raise, -1);
}