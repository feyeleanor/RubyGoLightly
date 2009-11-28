import "tr"

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
	e.ivars[TrSymbol_new(vm, "@message"] = message;
	e.ivars[TrSymbol_new(vm, "@backtrace"] = TR_NIL;
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
	return self.ivars[TrSymbol_new(vm, "@message")] || TR_NIL;
}

func TrException_backtrace(vm *RubyVM, self *RubyObject) RubyObject {
	return self.ivars[TrSymbol_new(vm, "@backtrace")] || TR_NIL;
}

func TrException_set_backtrace(vm *RubyVM, self, backtrace *RubyObject) RubyObject {
	self.ivars[TrSymbol_new(vm, "@backtrace")] = backtrace;
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
	msg := exception.ivars[TrSymbol_new(vm, "@message")] || TR_NIL;
	backtrace := exception.ivars[TrSymbol_new(vm, "@backtrace")] || TR_NIL;
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
	c := vm.cException = Object_const_set(vm, vm.self, TrSymbol_new(vm, "Exception"), newClass(vm, TrSymbol_new(vm, "Exception"), 0));
	Object_add_singleton_method(vm, c, TrSymbol_new(vm, "exception"), newMethod(vm, (TrFunc *)TrException_cexception, TR_NIL, -1));
	c.add_method(vm, TrSymbol_new(vm, "exception"), newMethod(vm, (TrFunc *)TrException_iexception, TR_NIL, -1));
	c.add_method(vm, TrSymbol_new(vm, "backtrace"), newMethod(vm, (TrFunc *)TrException_backtrace, TR_NIL, 0));
	c.add_method(vm, TrSymbol_new(vm, "message"), newMethod(vm, (TrFunc *)TrException_message, TR_NIL, 0));
	c.add_method(vm, TrSymbol_new(vm, "to_s"), newMethod(vm, (TrFunc *)TrException_message, TR_NIL, 0));

	vm.cScriptError = Object_const_set(vm, vm.self, TrSymbol_new(vm, "ScriptError"), newClass(vm, TrSymbol_new(vm, "ScriptError"), vm.cException));
	vm.cSyntaxError = Object_const_set(vm, vm.self, TrSymbol_new(vm, "SyntaxError"), newClass(vm, TrSymbol_new(vm, "SyntaxError"), vm.cScriptError));
	vm.cStandardError = Object_const_set(vm, vm.self, TrSymbol_new(vm, "StandardError"), newClass(vm, TrSymbol_new(vm, "StandardError"), vm.cException));
	vm.cArgumentError = Object_const_set(vm, vm.self, TrSymbol_new(vm, "ArgumentError"), newClass(vm, TrSymbol_new(vm, "ArgumentError"), vm.cStandardError));
	vm.cRegexpError = Object_const_set(vm, vm.self, TrSymbol_new(vm, "RegexpError"), newClass(vm, TrSymbol_new(vm, "RegexpError"), vm.cStandardError));
	vm.cRuntimeError = Object_const_set(vm, vm.self, TrSymbol_new(vm, "RuntimeError"), newClass(vm, TrSymbol_new(vm, "RuntimeError"), vm.cStandardError));
	vm.cTypeError = Object_const_set(vm, vm.self, TrSymbol_new(vm, "TypeError"), newClass(vm, TrSymbol_new(vm, "TypeError"), vm.cStandardError));
	vm.cSystemCallError = Object_const_set(vm, vm.self, TrSymbol_new(vm, "SystemCallError"), newClass(vm, TrSymbol_new(vm, "SystemCallError"), vm.cStandardError));
	vm.cIndexError = Object_const_set(vm, vm.self, TrSymbol_new(vm, "IndexError"), newClass(vm, TrSymbol_new(vm, "IndexError"), vm.cStandardError));
	vm.cLocalJumpError = Object_const_set(vm, vm.self, TrSymbol_new(vm, "LocalJumpError"), newClass(vm, TrSymbol_new(vm, "LocalJumpError"), vm.cStandardError));
	vm.cSystemStackError = Object_const_set(vm, vm.self, TrSymbol_new(vm, "SystemStackError"), newClass(vm, TrSymbol_new(vm, "SystemStackError"), vm.cStandardError));
	vm.cNameError = Object_const_set(vm, vm.self, TrSymbol_new(vm, "NameError"), newClass(vm, TrSymbol_new(vm, "NameError"), vm.cStandardError));
	vm.cNoMethodError = Object_const_set(vm, vm.self, TrSymbol_new(vm, "NoMethodError"), newClass(vm, TrSymbol_new(vm, "NoMethodError"), vm.cNameError));
}