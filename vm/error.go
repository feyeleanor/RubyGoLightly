import "tr"

/* Error management stuff */

/* Exception
 NoMemoryError
 ScriptError
   LoadError
   NotImplementedError
   SyntaxError
 SignalException
   Interrupt
 StandardError
   ArgumentError
   IOError
     EOFError
   IndexError
   LocalJumpError
   NameError
     NoMethodError
   RangeError
     FloatDomainError
   RegexpError
   RuntimeError
   SecurityError
   SystemCallError
   SystemStackError
   ThreadError
   TypeError
   ZeroDivisionError
 SystemExit
 fatal */

func TrException_new(vm *RubyVM, class, message *RubyObject) RubyObject {
	e := Object_alloc(vm, class);
	tr_setivar(e, "@message", message);
	tr_setivar(e, "@backtrace", TR_NIL);
	return e;
}

func TrException_cexception(vm *RubyVM, self *RubyObject, argc int, argv []RubyObject) {
	if !self.(Class) && !self.(Module) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + self));
		return TR_UNDEF;
	}
	if argc == 0 { return TrException_new(vm, self, self.name); }
	return TrException_new(vm, self, argv[0]);
}

func TrException_iexception(vm *RubyVM, self *RubyObject, argc int, argv []RubyObject) RubyObject {
	if (argc == 0) return self;
	if TR_IMMEDIATE(self) {
		class := vm.classes[Object_type(vm, self)];
	} else {
		class := Object *(self).class;
	}
	return TrException_new(vm, class, argv[0]);
}

func TrException_message(vm *RubyVM, self *RubyObject) RubyObject {
	return tr_getivar(self, "@message");
}

func TrException_backtrace(vm *RubyVM, self *RubyObject) RubyObject {
	return tr_getivar(self, "@backtrace");
}

func TrException_set_backtrace(vm *RubyVM, self, backtrace *RubyObject) RubyObject {
	return tr_setivar(self, "@backtrace", backtrace);
}

func TrException_default_handler(vm *RubyVM, exception *RubyObject) RubyObject {
	if TR_IMMEDIATE(exception) {
		exception_class := vm.classes[Object_type(vm, exception)];
	} else {
		exception_class := Object *(exception).class;
	}

	if !exception_class.(Class) && !exception_class.(Module) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + X));
		return TR_UNDEF;
	}
	msg := tr_getivar(exception, "@message");
	backtrace := tr_getivar(exception, "@backtrace");

	if !exception_class.name.(String) && !exception_class.name.(Symbol) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + exception_class.name));
		return TR_UNDEF;
	}
	if !msg.(String) && !msg.(Symbol) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + msg));
		return TR_UNDEF;
	}
	printf("%s: %s\n", exception_class.name.ptr, msg.ptr);
	if backtrace {
		for item := range backtrace.Iter() {
			if !item.(String) && !item.(Symbol) {
				vm.throw_reason = TR_THROW_EXCEPTION;
				vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + item));
				return TR_UNDEF;
			}
			println(item.ptr);
		}
	}
	vm.destroy();
	exit(1);
}

func TrError_init(vm *RubyVM) {
	c := vm.cException = tr_defclass("Exception", 0);
	tr_metadef(c, "exception", TrException_cexception, -1);
	tr_def(c, "exception", TrException_iexception, -1);
	tr_def(c, "backtrace", TrException_backtrace, 0);
	tr_def(c, "message", TrException_message, 0);
	tr_def(c, "to_s", TrException_message, 0);
  
	vm.cScriptError = tr_defclass("ScriptError", vm.cException);
	vm.cSyntaxError = tr_defclass("SyntaxError", vm.cScriptError);
	vm.cStandardError = tr_defclass("StandardError", vm.cException);
	vm.cArgumentError = tr_defclass("ArgumentError", vm.cStandardError);
	vm.cRegexpError = tr_defclass("RegexpError", vm.cStandardError);
	vm.cRuntimeError = tr_defclass("RuntimeError", vm.cStandardError);
	vm.cTypeError = tr_defclass("TypeError", vm.cStandardError);
	vm.cSystemCallError = tr_defclass("SystemCallError", vm.cStandardError);
	vm.cIndexError = tr_defclass("IndexError", vm.cStandardError);
	vm.cLocalJumpError = tr_defclass("LocalJumpError", vm.cStandardError);
	vm.cSystemStackError = tr_defclass("SystemStackError", vm.cStandardError);
	vm.cNameError = tr_defclass("NameError", vm.cStandardError);
	vm.cNoMethodError = tr_defclass("NoMethodError", vm.cNameError);
}