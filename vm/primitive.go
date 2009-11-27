#include "tr.h"
#include "internal.h"

func TrNil_to_s(vm *RubyVM, self *RubyObject) RubyObject { return TrString_new2(vm, ""); }

func TrTrue_to_s(vm *RubyVM, self *RubyObject) RubyObject { return TrString_new2(vm, "true"); }

func TrFalse_to_s(vm *RubyVM, self *RubyObject) RubyObject { return TrString_new2(vm, "false"); }

void TrPrimitive_init(vm *RubyVM) {
  nilc := vm.classes[TR_T_NilClass] = Object_const_set(vm, vm.self, tr_intern(NilClass), newClass(vm, tr_intern(NilClass), vm.classes[TR_T_Object]));
  truec := vm.classes[TR_T_TrueClass] = Object_const_set(vm, vm.self, tr_intern(TrueClass), newClass(vm, tr_intern(TrueClass), vm.classes[TR_T_Object]));
  falsec := vm.classes[TR_T_FalseClass] = Object_const_set(vm, vm.self, tr_intern(FalseClass), newClass(vm, tr_intern(FalseClass), vm.classes[TR_T_Object]));
  tr_def(nilc, "to_s", TrNil_to_s, 0);
  tr_def(truec, "to_s", TrTrue_to_s, 0);
  tr_def(falsec, "to_s", TrFalse_to_s, 0);
}