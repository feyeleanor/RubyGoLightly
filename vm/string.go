#include <alloca.h>
#include <stdarg.h>
#include <stdio.h>

import (
	"fmt";
	"tr";
	"internal";
)

// symbol

func TrSymbol_lookup(vm *struct TrVM, str const char *) OBJ {
  khash_t(str) *kh = vm.symbols;
  khiter_t k = kh_get(str, kh, str);
  if (k != kh_end(kh)) return kh_value(kh, k);
  return TR_NIL;
}

func TrSymbol_add(vm *struct TrVM, str const char *, id OBJ) {
  int ret;
  khash_t(str) *kh = vm.symbols;
  khiter_t k = kh_put(str, kh, str, &ret);
  if (!ret) kh_del(str, kh, k);
  kh_value(kh, k) = id;
}

func TrSymbol_new(vm *struct TrVM, str const char *) OBJ {
  OBJ id = TrSymbol_lookup(vm, str);
  
  if (!id) {
    TrSymbol *s = TR_INIT_CORE_OBJECT(Symbol);
    s.len = strlen(str);
    s.ptr = TR_ALLOC_N(char, s.len+1);
    s.interned = 1;
    TR_MEMCPY_N(s.ptr, str, char, s.len);
    s.ptr[s.len] = '\0';
    
    id = OBJ(s);
    TrSymbol_add(vm, s.ptr, id);
  }
  return id;
}

func TrSymbol_to_s(vm *struct TrVM, self OBJ) OBJ {
  return TrString_new(vm, TR_STR_PTR(self), TR_STR_LEN(self));
}

func TrSymbol_init(vm *struct TrVM) {
  OBJ c = TR_INIT_CORE_CLASS(Symbol, Object);
  tr_def(c, "to_s", TrSymbol_to_s, 0);
}

// string

func TrString_to_s(vm *struct TrVM, self OBJ) OBJ {
	return self;
}

func TrString_size(vm *struct TrVM, self) OBJ {
  return TR_INT2FIX(TR_CSTRING(self).len);
}

func TrString_new(vm *struct TrVM, str const char *, len size_t) OBJ {
  TrString *s = TR_INIT_CORE_OBJECT(String);
  s.len = len;
  s.ptr = TR_ALLOC_N(char, s.len+1);
  s.interned = 0;
  TR_MEMCPY_N(s.ptr, str, char, s.len);
  s.ptr[s.len] = '\0';
  return OBJ(s);
}

func TrString_new2(vm *struct TrVM, str const char *) OBJ {
  return TrString_new(vm, str, strlen(str));
}

func TrString_new3(vm *struct TrVM, len size_t) OBJ {
  TrString *s = TR_INIT_CORE_OBJECT(String);
  s.len = len;
  s.ptr = TR_ALLOC_N(char, s.len+1);
  s.interned = 0;
  s.ptr[s.len] = '\0';
  return OBJ(s);
}

func TrString_add(vm *struct TrVM, self, other OBJ) OBJ {
  return tr_sprintf(vm, "%s%s", TR_STR_PTR(self), TR_STR_PTR(other));
}

func TrString_push(vm *struct TrVM, self, other OBJ) OBJ {
  TrString *s = TR_CSTRING(self);
  TrString *o = TR_CSTRING(other);
  
  size_t orginal_len = s.len;
  s.len += o.len;
  s.ptr = TR_REALLOC(s.ptr, s.len+1);
  TR_MEMCPY_N(s.ptr + orginal_len, o.ptr, char, o.len);
  s.ptr[s.len] = '\0';

  return self;
}

func TrString_replace(vm *struct TrVM, self, other OBJ) OBJ {
  TR_FREE(TR_STR_PTR(self));
  TR_STR_PTR(self) = TR_STR_PTR(other);
  TR_STR_LEN(self) = TR_STR_LEN(other);
  return self;
}

func TrString_cmp(vm *struct TrVM, self, other OBJ) OBJ {
  if (!other.(String)) return TR_INT2FIX(-1);
  return TR_INT2FIX(strcmp(TR_STR_PTR(self), TR_STR_PTR(other)));
}

func TrString_substring(vm *struct TrVM, self, start, len OBJ) OBJ {
  int s = TR_FIX2INT(start);
  int l = TR_FIX2INT(len);
  if (s < 0 || (s+l) > (int)TR_STR_LEN(self)) return TR_NIL;
  return TrString_new(vm, TR_STR_PTR(self)+s, l);
}

func TrString_to_sym(vm *struct TrVM, OBJ self) OBJ {
  return tr_intern(TR_STR_PTR(self));
}

// Uses variadic ... parameter which replaces the mechanism used by stdarg.h
func tr_sprintf(vm *struct TrVM, const char *fmt, args ...) OBJ {
  va_list arg;
  va_start(arg, fmt);
  int len = vsnprintf(NULL, 0, fmt, arg);
  char *ptr = alloca(sizeof(char) * len);
  va_end(arg);
  va_start(arg, fmt);
  vsprintf(ptr, fmt, arg);
  va_end(arg);
  OBJ str = TrString_new(vm, ptr, len);
  TR_FREE(ptr);
  return str;
}

func TrString_init(vm *struct TrVM) {
  OBJ c = TR_INIT_CORE_CLASS(String, Object);
  tr_def(c, "to_s", TrString_to_s, 0);
  tr_def(c, "to_sym", TrString_to_sym, 0);
  tr_def(c, "size", TrString_size, 0);
  tr_def(c, "replace", TrString_replace, 1);
  tr_def(c, "substring", TrString_substring, 2);
  tr_def(c, "+", TrString_add, 1);
  tr_def(c, "<<", TrString_push, 1);
  tr_def(c, "<=>", TrString_cmp, 1);
}