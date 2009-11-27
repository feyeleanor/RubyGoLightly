import(
	"tr";
	"internal";
	)

func TrBinding_new(vm *RubyVM, frame *Frame) RubyObject {
	return Binding{type: TR_T_Binding, class: vm.classes[TR_T_Binding], ivars: kh_init(RubyObject), frame: frame};
}

func TrBinding_init(vm *RubyVM) {
	vm.classes[TR_T_Binding] = Object_const_set(vm, vm.self, tr_intern(Binding), newClass(vm, tr_intern(Binding), vm.classes[TR_T_Object]));
}

func TrKernel_puts(vm *RubyVM, self *RubyObject, argc int, argv []RubyObject) RubyObject {
	object_as_string := tr_send2(object, "to_s");
	if !object_as_string.(String) && !object_as_string.(Symbol) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + object_as_string));
		return TR_UNDEF;
	}
	for object := range argv { fmt.println("%s", object_as_string.ptr); }
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
		if !argv[1].(Binding) {
			vm.throw_reason = TR_THROW_EXCEPTION;
			vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected Binding"));
			return TR_UNDEF;
		}
		frame := TrBinding *(argv[1]);
	} else {
		frame := vm.frame;
	}
	if argc > 2 && argv[1] {
		if !argv[2].(String) && !argv[2].(Symbol) {
			vm.throw_reason = TR_THROW_EXCEPTION;
			vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + argv[2]));
			return TR_UNDEF;
		}		
		filename := argv[2].ptr;
	} else {
		filename := "<eval>";
	}
	if argc > 3 {
		lineno := TR_FIX2INIT(argv[3]);
	} else {
		lineno := 0;
	}
	if !code_string.(String) && !code_string.(Symbol) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + code_string));
		return TR_UNDEF;
	}
	blk := Block_compile(vm, code_string.ptr, filename, lineno);
	if !blk { return TR_UNDEF }
	if vm.debug { blk.dump2(vm, 0) }
	return vm.run(blk, frame.self, frame.class, frame.stack[0:blk.locals.Len() - 1]);
}

func TrKernel_load(vm *RubyVM, self, filename *RubyObject) RubyObject {
	if !filename.(String) && !filename.(Symbol) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + filename));
		return TR_UNDEF;
	}
	return vm.load(filename.ptr);
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
	vm.throw_reason = TR_THROW_EXCEPTION;
	vm.throw_value = e;
	return TR_UNDEF;
}

func TrKernel_init(vm *RubyVM) {
	m := tr_defmodule("Kernel");
	vm.classes[TR_T_Object].include(vm, m);
	tr_def(m, "puts", TrKernel_puts, -1);
	tr_def(m, "eval", TrKernel_eval, -1);
	tr_def(m, "load", TrKernel_load, 1);
	tr_def(m, "binding", TrKernel_binding, 0);
	tr_def(m, "raise", TrKernel_raise, -1);
}