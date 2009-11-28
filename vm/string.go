#include <alloca.h>
#include <stdarg.h>
#include <stdio.h>

import (
	"bytes";
	"fmt";
	"tr";
)

// symbol

func TrSymbol_lookup(vm *RubyVM, name string) RubyObject {
	return vm.symbols[name] || TR_NIL;
}

func TrSymbol_add(vm *RubyVM, name *string, id *RubyObject) {
	vm.symbols[name] = id;
}

func TrSymbol_new(vm *RubyVM, str *string) RubyObject {
	id := TrSymbol_lookup(vm, str);
  
	if (!id) {
		s := Symbol{type: TR_T_Symbol, class: vm.classes[TR_T_Symbol], ivars: make(map[string] RubyObject), len: strlen(str), ptr: make([]byte, s.len + 1), interned: true};
		bytes.Copy(s.ptr, str[0:s.len - 1]);
		s.ptr[s.len] = '\0';
		id := s;
		TrSymbol_add(vm, s.ptr, id);
	}
	return id;
}

func TrSymbol_to_s(vm *RubyVM, self *RubyObject) RubyObject {
	if !self.(String) && !self.(Symbol) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + self));
		return TR_UNDEF;
	}
	return TrString_new(vm, self.ptr, self.len);
}

func TrSymbol_init(vm *RubyVM) {
	c := vm.classes[TR_T_Symbol] = Object_const_set(vm, vm.self, TrSymbol_new(vm, Symbol), newClass(vm, TrSymbol_new(vm, Symbol), vm.classes[TR_T_Object]));
	c.add_method(vm, TrSymbol_new(vm, "to_s"), newMethod(vm, (TrFunc *)TrSymbol_to_s, TR_NIL, 0));
}

// string

func TrString_to_s(vm *RubyVM, self *RubyObject) RubyObject {
	return self;
}

func TrString_size(vm *RubyVM, self) RubyObject {
	if !self.(String) && !self.(Symbol) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + self));
		return TR_UNDEF;
	}
	return TR_INT2FIX(self.len);
}

func TrString_new(vm *RubyVM, str *string, len size_t) RubyObject {
	s := String{type: TR_T_String, class: vm.classes[TR_T_String], ivars: make(map[string] RubyObject), len: len, ptr: make([]byte, s.len + 1)};
	bytes.Copy(s.ptr, str[0:s.len - 1]);
	s.ptr[s.len] = '\0';
	return s;
}

func TrString_new2(vm *RubyVM, str *string) RubyObject {
	return TrString_new(vm, str, strlen(str));
}

func TrString_new3(vm *RubyVM, len size_t) RubyObject {
	s := String{type: TR_T_String, class: vm.classes[TR_T_String], ivars: make(map[string] RubyObject), len: len, ptr: make([]byte, s.len + 1)};
	s.ptr[s.len] = '\0'
	return s;
}

func TrString_add(vm *RubyVM, self, other *RubyObject) RubyObject {
	if !self.(String) && !self.(Symbol) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + self));
		return TR_UNDEF;
	}
	if !other.(String) && !other.(Symbol) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + other));
		return TR_UNDEF;
	}
	return tr_sprintf(vm, "%s%s", self.ptr, other.ptr);
}

func TrString_push(vm *RubyVM, self, other *RubyObject) RubyObject {
	if !self.(String) && !self.(Symbol) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + self));
		return TR_UNDEF;
	}
	if !other.(String) && !other.(Symbol) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + other));
		return TR_UNDEF;
	}
	orginal_len := self.len;
	self.len += other.len;
	self.ptr := TR_REALLOC(self.ptr, self.len + 1);
	memcpy(self.ptr + original_len, other.ptr, sizeof(char) * other.len);
	self.ptr[self.len] = '\0';
	return self;
}

func TrString_replace(vm *RubyVM, self, other *RubyObject) RubyObject {
	if !self.(String) && !self.(Symbol) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + self));
		return TR_UNDEF;
	}
	if !other.(String) && !other.(Symbol) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + other));
		return TR_UNDEF;
	}
	self.ptr, self.len = other.ptr, other.len;
	return self;
}

func TrString_cmp(vm *RubyVM, self, other *RubyObject) RubyObject {
	if (!other.(String)) return TR_INT2FIX(-1);
	if !self.(String) && !self.(Symbol) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + self));
		return TR_UNDEF;
	}
	return TR_INT2FIX(strcmp(self.ptr, other.ptr));
}

func TrString_substring(vm *RubyVM, self, start, len *RubyObject) RubyObject {
	int s = TR_FIX2INT(start);
	int l = TR_FIX2INT(len);
	if !self.(String) && !self.(Symbol) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + self));
		return TR_UNDEF;
	}
	if s < 0 || (s + l) > self.len { return TR_NIL; }
	return TrString_new(vm, self.ptr + s, l);
}

func TrString_to_sym(vm *RubyVM, self *RubyObject) RubyObject {
	if !self.(String) && !self.(Symbol) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + self));
		return TR_UNDEF;
	}
	return TrSymbol_new(vm, self.ptr);
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
	c := vm.classes[TR_T_String] = Object_const_set(vm, vm.self, TrSymbol_new(vm, String), newClass(vm, TrSymbol_new(vm, String), vm.classes[TR_T_Object]));
	c.add_method(vm, TrSymbol_new(vm, "to_s"), newMethod(vm, (TrFunc *)TrString_to_s, TR_NIL, 0));
	c.add_method(vm, TrSymbol_new(vm, "to_sym"), newMethod(vm, (TrFunc *)TrString_to_sym, TR_NIL, 0));
	c.add_method(vm, TrSymbol_new(vm, "size"), newMethod(vm, (TrFunc *)TrString_size, TR_NIL, 0));
	c.add_method(vm, TrSymbol_new(vm, "replace"), newMethod(vm, (TrFunc *)TrString_replace, TR_NIL, 1));
	c.add_method(vm, TrSymbol_new(vm, "substring"), newMethod(vm, (TrFunc *)ToString_substring, TR_NIL, 2));
	c.add_method(vm, TrSymbol_new(vm, "+"), newMethod(vm, (TrFunc *)TrString_add, TR_NIL, 1));
	c.add_method(vm, TrSymbol_new(vm, "<<"), newMethod(vm, (TrFunc *)TrString_push, TR_NIL, 1));
	c.add_method(vm, TrSymbol_new(vm, "<=>"), newMethod(vm, (TrFunc *)TrString_cmp, TR_NIL, 1));
}