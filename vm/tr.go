#include <stdlib.h>
#include <stdio.h>
#include <limits.h>
#include <assert.h>
#include <errno.h>

#include <gc.h>
#include <pcre.h>

#include "config.h"
#include "vendor/khash.h"

/* allocation macros */
#define TR_REALLOC           GC_realloc

/* type convertion macros */
#define TR_CLASS(X)          (TR_IMMEDIATE(X) ? vm.classes[Object_type(vm, (X))] : Object *(X).class)
#define TR_TYPE_ERROR(T) {	\
	vm.throw_reason = TR_THROW_EXCEPTION; \
	vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " #T)); \
	return TR_UNDEF;	\
}

#define TR_CTYPE(X,T)        ((X.(T) ? 0 : TR_TYPE_ERROR(T)),(Tr##T*)(X))
#define TR_CSTRING(X)        ((X.(String) || X.(Symbol) ? 0 : TR_TYPE_ERROR(T)),(TrString*)(X))

/* raw hash macros */
#define TR_KH_GET(KH,K) ({ \
	key := (K); \
	hash := (KH); \
	k := kh_get(RubyObject, hash, key); \
	k == kh_end(hash) ? TR_NIL : kh_value(hash, k); \
})
#define TR_KH_SET(KH,K,V) ({ \
	key := (K); \
	hash := (KH); \
	int ret; \
	k := kh_put(RubyObject, hash, key, &ret); \
	kh_value(hash, k) = (V); \
})
#define TR_KH_EACH(H,I,V,B) ({ \
	khiter_t __k##V; \
	for (__k##V = kh_begin(H); __k##V != kh_end(H); ++__k##V) \
		if (kh_exist((H), __k##V)) { \
			V := kh_value((H), __k##V); \
			B \
		} \
	})

/* immediate values macros */
#define TR_IMMEDIATE(X)      (X==TR_NIL || X==TR_TRUE || X==TR_FALSE || X==TR_UNDEF || TR_IS_FIX(X))
#define TR_IS_FIX(F)         ((F) & 1)
#define TR_FIX2INT(F)        (((int)(F) >> 1))
#define TR_INT2FIX(I)        ((I) << 1 |  1)
#define TR_NIL               OBJ(0)
#define TR_FALSE             OBJ(2)
#define TR_TRUE              OBJ(4)
#define TR_UNDEF             OBJ(6)

/* API macros */
#define tr_getivar(O,N)      TR_KH_GET(Object*(O).ivars, tr_intern(N))
#define tr_setivar(O,N,V)    TR_KH_SET(Object*(O).ivars, tr_intern(N), V)
#define tr_getglobal(N)      TR_KH_GET(vm.globals, tr_intern(N))
#define tr_setglobal(N,V)    TR_KH_SET(vm.globals, tr_intern(N), V)
#define tr_intern(S)         TrSymbol_new(vm, (S))
#define tr_raise(T,M,...)    {	\
	vm.throw_reason = TR_THROW_EXCEPTION; \
	vm.throw_value = TrException_new(vm, vm.c##T, tr_sprintf(vm, (M), ##__VA_ARGS__)); \
	return TR_UNDEF; \
}

#define tr_raise_errno(M)    tr_raise(SystemCallError, "%s: %s", strerror(errno), (M))
#define tr_def(C,N,F,A)      (C).add_method(vm, tr_intern(N), newMethod(vm, (TrFunc *)(F), TR_NIL, (A)))
#define tr_metadef(O,N,F,A)  Object_add_singleton_method(vm, (O), tr_intern(N), newMethod(vm, (TrFunc *)(F), TR_NIL, (A)))
#define tr_defclass(N,S)     Object_const_set(vm, vm.self, tr_intern(N), newClass(vm, tr_intern(N), S))
#define tr_defmodule(N)      Object_const_set(vm, vm.self, tr_intern(N), newModule(vm, tr_intern(N)))

#define tr_send(R,MSG,...)   ({ \
	__argv[] := { (MSG), ##__VA_ARGS__ }; \
	Object_send(vm, R, sizeof(__argv)/sizeof(RubyObject), __argv); \
})
#define tr_send2(R,STR,...)  tr_send((R), tr_intern(STR), ##__VA_ARGS__)

typedef unsigned long OBJ;
typedef unsigned char u8;

KHASH_MAP_INIT_STR(str, RubyObject)
KHASH_MAP_INIT_INT(RubyObject, RubyObject)

typedef enum {
  /*  0 */ TR_T_Object, TR_T_Module, TR_T_Class, TR_T_Method, TR_T_Binding,
  /*  5 */ TR_T_Symbol, TR_T_String, TR_T_Fixnum, TR_T_Range, TR_T_Regexp,
  /* 10 */ TR_T_NilClass, TR_T_TrueClass, TR_T_FalseClass,
  /* 12 */ TR_T_Array, TR_T_Hash,
  /* 14 */ TR_T_Node,
  TR_T_MAX /* keep last */
} TR_T;

typedef enum {
	TR_THROW_EXCEPTION,
	TR_THROW_RETURN,
	TR_THROW_BREAK
} TR_THROW_REASON;

type TrCallSite struct {
	class			*RubyObject;
	method			*RubyObject;
	message			*RubyObject;
	method_missing	bool;
	miss			size_t;
}

type TrUpval struct {
	value			*RubyObject;
	closed			*RubyObject;		// value when closed
}

typedef OBJ (TrFunc)(vm *struct TrVM, receiver *RubyObject, ...);

type TrBinding struct {
	type 			TR_T;
	class			*RubyObject;
	ivars			map[string] *RubyObject;
  	frame			*Frame;
}

type TrString struct {
	type 			TR_T;
	class			*RubyObject;
	ivars			*map[string] *RubyObject;
	ptr				*char;
	len				size_t;
	interned		bool;
}
type TrSymbol TrString

type TrRange struct {
	type			TR_T;
	class			*RubyObject;
	ivars			*map[string] *RubyObject;
	first, last		*RubyObject;
	exclusive		int;
}

type TrHash struct {
	type			TR_T;
	class			*RubyObject;
	ivars			*map[string] *RubyObject;
	kh				*map[string] *RubyObject;
}

type TrRegexp struct {
	type			TR_T;
	class			*RubyObject;
	ivars			*map[string] RubyObject;
  	re				*pcre;
}

#ifdef __unix__
  #include <unistd.h>
#else
  #include "freegetopt/getopt.h"
#endif

func usage() {
	fmt.println("usage: tinyrb [options] [file]");
	fmt.println("options:";
	fmt.println("  -e   eval code");
	fmt.println("  -d   show debug info (multiple times for more)");
	fmt.println("  -v   print version");
	fmt.println("  -h   print this");
	return 1;
}

func main(argc int, argv *[]char) {
	int opt;
	vm := newRubyVM();

	while((opt = getopt(argc, argv, "e:vdh")) != -1) {
		switch(opt) {
			case 'e':
				if vm.eval(optarg, "<eval>") == TR_UNDEF && vm.throw_reason == TR_THROW_EXCEPTION {
					TrException_default_handler(vm, vm.throw_value));
					abort();
				}
				return 0;
			case 'v':
				fmt.println("tinyrb %s", TR_VERSION);
				return 1;
			case 'd':
				vm.debug++;
				continue;
			default:
				return usage();
		}
	}

	// These lines allow us to tread argc and argv as though any switches were not there
	argc -= optind;
	argv += optind;
  
	if (argc > 0) {
		if vm.load(argv[argc - 1])) == TR_UNDEF && vm.throw_reason == TR_THROW_EXCEPTION {
			TrException_default_handler(vm, vm.throw_value));
			abort();
		}
		return 0;
	}
	return usage();
}