#include "tr.h"

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
  if (argc == 0) return TrException_new(vm, self, TR_CCLASS(self).name);
  return TrException_new(vm, self, argv[0]);
}

func TrException_iexception(vm *RubyVM, self *RubyObject, argc int, argv []RubyObject) RubyObject {
  if (argc == 0) return self;
  return TrException_new(vm, TR_CLASS(self), argv[0]);
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
  Class *c = TR_CCLASS(TR_CLASS(exception));
  msg := tr_getivar(exception, "@message");
  backtrace := tr_getivar(exception, "@backtrace");
  
  printf("%s: %s\n", TR_CSTRING(c.name).ptr, TR_CSTRING(msg).ptr);
  if backtrace {
	for item := range backtrace.Iter() { println(TR_CSTRING(item).ptr); }
  }
  vm.destroy();
  exit(1);
}

void TrError_init(vm *RubyVM) {
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