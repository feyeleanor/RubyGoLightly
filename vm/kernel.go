import(
	"tr";
	"internal";
	)

func TrBinding_new(vm *RubyVM, frame *Frame) RubyObject {
	return Binding{type: TR_T_Binding, class: vm.classes[TR_T_Binding], ivars: kh_init(RubyObject), frame: frame};
}

void TrBinding_init(vm *RubyVM) {
	vm.classes[TR_T_Binding] = Object_const_set(vm, vm.self, tr_intern(Binding), newClass(vm, tr_intern(Binding), vm.classes[TR_T_Object]));
}

func TrKernel_puts(vm *RubyVM, self *RubyObject, argc int, argv []RubyObject) RubyObject {
	for object := range argv { fmt.println("%s", TR_STR_PTR(tr_send2(object, "to_s"))); }
	return TR_NIL;
}

func TrKernel_binding(vm *RubyVM, self RubyObject) RubyObject {
	return TrBinding_new(vm, vm.frame.previous);
}

func TrKernel_eval(vm *RubyVM, self *RubyObject, argc int, argv []RubyObject) RubyObject {
	if argc < 1 { tr_raise(ArgumentError, "string argument required") }
	if argc > 4 { tr_raise(ArgumentError, "Too much arguments") }
	code_string := argv[0];
	if argc > 1 && argv[1] {
		frame := TR_CTYPE(argv[1], Binding);
	} else {
		frame := vm.frame;
	}
	if argc > 2 && argv[1] {
		filename := TR_STR_PTR(argv[2]);
	} else {
		filename := "<eval>";
	}
	if argc > 3 {
		lineno := TR_FIX2INIT(argv[3]);
	} else {
		lineno := 0;
	}
	blk := Block_compile(vm, TR_STR_PTR(code_string), filename, lineno);
	if !blk { return TR_UNDEF }
	if vm.debug { blk.dump(vm) }
	return vm.run(blk, frame.self, frame.class, frame.stack[0:blk.locals.Len() - 1]);
}

func TrKernel_load(vm *RubyVM, self, filename *RubyObject) RubyObject {
	return vm.load(TR_STR_PTR(filename));
}

func TrKernel_raise(vm *RubyVM, self *RubyObject, argc int, argv []RubyObject) RubyObject {
	e := TR_NIL;
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
  m := tr_defmodule("Kernel");
  vm.classes[TR_T_Object].include(vm, m);
  tr_def(m, "puts", TrKernel_puts, -1);
  tr_def(m, "eval", TrKernel_eval, -1);
  tr_def(m, "load", TrKernel_load, 1);
  tr_def(m, "binding", TrKernel_binding, 0);
  tr_def(m, "raise", TrKernel_raise, -1);
}