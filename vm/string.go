#include <alloca.h>
#include <stdarg.h>
#include <stdio.h>

import (
	"bytes";
	"fmt";
	"tr";
	"internal";
)

// symbol

func TrSymbol_lookup(vm *RubyVM, str string) RubyObject {
	symbols := vm.symbols;
	k := kh_get(str, symbols, str);
	if (k != kh_end(symbols)) return kh_value(symbols, k);
	return TR_NIL;
}

func TrSymbol_add(vm *RubyVM, str *string, id *RubyObject) {
	ret int;
	symbols := vm.symbols;
	k := kh_put(str, symbols, str, &ret);
	if (!ret) kh_del(str, symbols, k);
	kh_value(symbols, k) = id;
}

func TrSymbol_new(vm *RubyVM, str *string) RubyObject {
	id := TrSymbol_lookup(vm, str);
  
	if (!id) {
		s := Symbol{type: TR_T_Symbol, class: vm.classes[TR_T_Symbol], ivars: kh_init(RubyObject), len: strlen(str), ptr: make([]byte, s.len + 1), interned: true};
		bytes.Copy(s.ptr, str[0:s.len - 1]);
		s.ptr[s.len] = '\0';
		id := s;
		TrSymbol_add(vm, s.ptr, id);
	}
	return id;
}

func TrSymbol_to_s(vm *RubyVM, self *RubyObject) RubyObject {
	return TrString_new(vm, TR_STR_PTR(self), TR_STR_LEN(self));
}

func TrSymbol_init(vm *RubyVM) {
  c := vm.classes[TR_T_Symbol] = Object_const_set(vm, vm.self, tr_intern(Symbol), newClass(vm, tr_intern(Symbol), vm.classes[TR_T_Object]));
  tr_def(c, "to_s", TrSymbol_to_s, 0);
}

// string

func TrString_to_s(vm *RubyVM, self *RubyObject) RubyObject {
	return self;
}

func TrString_size(vm *RubyVM, self) RubyObject {
  return TR_INT2FIX(TR_CSTRING(self).len);
}

func TrString_new(vm *RubyVM, str *string, len size_t) RubyObject {
	s := String{type: TR_T_String, class: vm.classes[TR_T_String], ivars: kh_init(RubyObject), len: len, ptr: make([]byte, s.len + 1)};
	bytes.Copy(s.ptr, str[0:s.len - 1]);
	s.ptr[s.len] = '\0';
	return s;
}

func TrString_new2(vm *RubyVM, str *string) RubyObject {
	return TrString_new(vm, str, strlen(str));
}

func TrString_new3(vm *RubyVM, len size_t) RubyObject {
	s := String{type: TR_T_String, class: vm.classes[TR_T_String], ivars: kh_init(RubyObject), len: len, ptr: make([]byte, s.len + 1)};
	s.ptr[s.len] = '\0'
	return s;
}

func TrString_add(vm *RubyVM, self, other *RubyObject) RubyObject {
	return tr_sprintf(vm, "%s%s", TR_STR_PTR(self), TR_STR_PTR(other));
}

func TrString_push(vm *RubyVM, self, other *RubyObject) RubyObject {
	s := TR_CSTRING(self);
	o := TR_CSTRING(other);
  
	orginal_len := s.len;
	s.len += o.len;
	s.ptr := TR_REALLOC(s.ptr, s.len+1);
	memcpy(s.ptr + original_len, o.ptr, sizeof(char) * o.len);
	s.ptr[s.len] = '\0';
	return self;
}

func TrString_replace(vm *RubyVM, self, other *RubyObject) RubyObject {
	TR_STR_PTR(self) = TR_STR_PTR(other);
	TR_STR_LEN(self) = TR_STR_LEN(other);
	return self;
}

func TrString_cmp(vm *RubyVM, self, other *RubyObject) RubyObject {
	if (!other.(String)) return TR_INT2FIX(-1);
	return TR_INT2FIX(strcmp(TR_STR_PTR(self), TR_STR_PTR(other)));
}

func TrString_substring(vm *RubyVM, self, start, len *RubyObject) RubyObject {
	int s = TR_FIX2INT(start);
	int l = TR_FIX2INT(len);
	if (s < 0 || (s+l) > (int)TR_STR_LEN(self)) return TR_NIL;
	return TrString_new(vm, TR_STR_PTR(self)+s, l);
}

func TrString_to_sym(vm *RubyVM, self *RubyObject) RubyObject {
	return tr_intern(TR_STR_PTR(self));
}

// Uses variadic ... parameter which replaces the mechanism used by stdarg.h
func tr_sprintf(vm *RubyVM, fmt *string, args ...) RubyObject {
	arg va_list;
	va_start(arg, fmt);
	len := vsnprintf(NULL, 0, fmt, arg);
	char *ptr = alloca(sizeof(char) * len);
	va_end(arg);
	va_start(arg, fmt);
	vsprintf(ptr, fmt, arg);
	va_end(arg);
	str := TrString_new(vm, ptr, len);
	return str;
}

func TrString_init(vm *RubyVM) {
	c := vm.classes[TR_T_String] = Object_const_set(vm, vm.self, tr_intern(String), newClass(vm, tr_intern(String), vm.classes[TR_T_Object]));
	tr_def(c, "to_s", TrString_to_s, 0);
	tr_def(c, "to_sym", TrString_to_sym, 0);
	tr_def(c, "size", TrString_size, 0);
	tr_def(c, "replace", TrString_replace, 1);
	tr_def(c, "substring", TrString_substring, 2);
	tr_def(c, "+", TrString_add, 1);
	tr_def(c, "<<", TrString_push, 1);
	tr_def(c, "<=>", TrString_cmp, 1);
}