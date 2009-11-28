#include <stdlib.h>
#include <stdio.h>
#include <limits.h>
#include <assert.h>
#include <errno.h>

#include <gc.h>
#include <pcre.h>

#ifdef __unix__
  #include <unistd.h>
#else
  #include "freegetopt/getopt.h"
#endif

const (
	TR_VERSION		"0.0";
	TR_MAX_FRAMES	255;
)

/* allocation macros */
#define TR_REALLOC           GC_realloc

/* immediate values macros */
#define TR_IMMEDIATE(X)      (X==TR_NIL || X==TR_TRUE || X==TR_FALSE || X==TR_UNDEF || TR_IS_FIX(X))
#define TR_IS_FIX(F)         ((F) & 1)
#define TR_FIX2INT(F)        (((int)(F) >> 1))
#define TR_INT2FIX(I)        ((I) << 1 |  1)
#define TR_NIL               OBJ(0)
#define TR_FALSE             OBJ(2)
#define TR_TRUE              OBJ(4)
#define TR_UNDEF             OBJ(6)

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

const(
	TR_T_Object = iota;
	TR_T_Module;
	TR_T_Class;
	TR_T_Method;
	TR_T_Binding;
	TR_T_Symbol;
	TR_T_String;
	TR_T_Fixnum;
	TR_T_Range;
	TR_T_Regexp;
 	TR_T_NilClass;
	TR_T_TrueClass;
	TR_T_FalseClass,
	TR_T_Array;
	TR_T_Hash;
	TR_T_Node;
	TR_T_MAX;			// keep last
)

const(
	TR_THROW_EXCEPTION = iota;
	TR_THROW_RETURN;
	TR_THROW_BREAK;
)

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
	ivars			map[string] *RubyObject;
	ptr				*char;
	len				size_t;
	interned		bool;
}
type TrSymbol TrString

type TrRange struct {
	type			TR_T;
	class			*RubyObject;
	ivars			map[string] *RubyObject;
	first, last		*RubyObject;
	exclusive		int;
}

type TrHash struct {
	type			TR_T;
	class			*RubyObject;
	ivars			map[string] RubyObject;
	hash			map[string] RubyObject;
}

type TrRegexp struct {
	type			TR_T;
	class			*RubyObject;
	ivars			map[string] RubyObject;
  	re				*pcre;
}

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