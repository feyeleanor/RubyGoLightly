#include <stdlib.h>
#include <stdio.h>
#include <limits.h>
#include <assert.h>
#include <errno.h>

#include <gc.h>
#include <pcre.h>

#include "config.h"
#include "vendor/kvec.h"
#include "vendor/khash.h"

/* allocation macros */
#define TR_MALLOC            GC_malloc
#define TR_CALLOC(m,n)       GC_MALLOC((m)*(n))
#define TR_REALLOC           GC_realloc
#define TR_FREE(S)           do { (void)(S); } while (0)

/* type convertion macros */
#define TR_TYPE(X)           TrObject_type(vm, (X))
#define TR_CLASS(X)          (TR_IMMEDIATE(X) ? vm.classes[TR_TYPE(X)] : TR_COBJECT(X).class)
#define TR_IS_A(X,T)         (TR_TYPE(X) == TR_T_##T)
#define TR_COBJECT(X)        ((TrObject*)(X))
#define TR_TYPE_ERROR(T)     TR_THROW(EXCEPTION, TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " #T)))
#define TR_CTYPE(X,T)        ((X.(T) ? 0 : TR_TYPE_ERROR(T)),(Tr##T*)(X))
#define TR_CHASH(X)          TR_CTYPE(X,Hash)
#define TR_CRANGE(X)         TR_CTYPE(X,Range)
#define TR_CREGEXP(X)        TR_CTYPE(X,Regexp)
#define TR_CSTRING(X)        ((X.(String) || X.(Symbol) ? 0 : TR_TYPE_ERROR(T)),(TrString*)(X))
#define TR_CBINDING(X)       TR_CTYPE(X,Binding)

/* string macros */
#define TR_STR_PTR(S)        (TR_CSTRING(S).ptr)
#define TR_STR_LEN(S)        (TR_CSTRING(S).len)

/* raw hash macros */
#define TR_KH_GET(KH,K) ({ \
  OBJ key = (K); \
  khash_t(OBJ) *kh = (KH); \
  khiter_t k = kh_get(OBJ, kh, key); \
  k == kh_end(kh) ? TR_NIL : kh_value(kh, k); \
})
#define TR_KH_SET(KH,K,V) ({ \
  OBJ key = (K); \
  khash_t(OBJ) *kh = (KH); \
  int ret; \
  khiter_t k = kh_put(OBJ, kh, key, &ret); \
  kh_value(kh, k) = (V); \
})
#define TR_KH_EACH(H,I,V,B) ({ \
    khiter_t __k##V; \
    for (__k##V = kh_begin(H); __k##V != kh_end(H); ++__k##V) \
      if (kh_exist((H), __k##V)) { \
        OBJ V = kh_value((H), __k##V); \
        B \
      } \
  })

/* throw macros */
#define TR_THROW(R,V)        ({ \
                               vm.throw_reason = TR_THROW_##R; \
                               vm.throw_value = (V); \
                               return TR_UNDEF; \
                             })
#define TR_HAS_EXCEPTION(R)  ((R) == TR_UNDEF && vm.throw_reason == TR_THROW_EXCEPTION)
#define TR_FAILSAFE(R)       if (TR_HAS_EXCEPTION(R)) { \
                               TrException_default_handler(vm, TR_EXCEPTION); \
                               abort(); \
                             }
#define TR_EXCEPTION         (assert(vm.throw_reason == TR_THROW_EXCEPTION), vm.throw_value)

/* immediate values macros */
#define TR_IMMEDIATE(X)      (X==TR_NIL || X==TR_TRUE || X==TR_FALSE || X==TR_UNDEF || TR_IS_FIX(X))
#define TR_IS_FIX(F)         ((F) & 1)
#define TR_FIX2INT(F)        (((int)(F) >> 1))
#define TR_INT2FIX(I)        ((I) << 1 |  1)
#define TR_NIL               OBJ(0)
#define TR_FALSE             OBJ(2)
#define TR_TRUE              OBJ(4)
#define TR_UNDEF             OBJ(6)
#define TR_TEST(X)           ((X) == TR_NIL || (X) == TR_FALSE ? 0 : 1)
#define TR_BOOL(X)           ((X) ? TR_TRUE : TR_FALSE)

/* core classes macros */
#define TR_INIT_CORE_OBJECT(T) ({ \
  Tr##T *o = TR_ALLOC(Tr##T); \
  o.type  = TR_T_##T; \
  o.class = vm.classes[TR_T_##T]; \
  o.ivars = kh_init(OBJ); \
  o; \
})
#define TR_CORE_CLASS(T)     vm.classes[TR_T_##T]
#define TR_INIT_CORE_CLASS(T,S) \
  TR_CORE_CLASS(T) = TrObject_const_set(vm, vm.self, tr_intern(#T), newClass(vm, tr_intern(#T), TR_CORE_CLASS(S)))

/* API macros */
#define tr_getivar(O,N)      TR_KH_GET(TR_COBJECT(O).ivars, tr_intern(N))
#define tr_setivar(O,N,V)    TR_KH_SET(TR_COBJECT(O).ivars, tr_intern(N), V)
#define tr_getglobal(N)      TR_KH_GET(vm.globals, tr_intern(N))
#define tr_setglobal(N,V)    TR_KH_SET(vm.globals, tr_intern(N), V)
#define tr_intern(S)         TrSymbol_new(vm, (S))
#define tr_raise(T,M,...)    TR_THROW(EXCEPTION, TrException_new(vm, vm.c##T, tr_sprintf(vm, (M), ##__VA_ARGS__)))
#define tr_raise_errno(M)    tr_raise(SystemCallError, "%s: %s", strerror(errno), (M))
#define tr_def(C,N,F,A)      (C).add_method(vm, tr_intern(N), newMethod(vm, (TrFunc *)(F), TR_NIL, (A)))
#define tr_metadef(O,N,F,A)  TrObject_add_singleton_method(vm, (O), tr_intern(N), newMethod(vm, (TrFunc *)(F), TR_NIL, (A)))
#define tr_defclass(N,S)     TrObject_const_set(vm, vm.self, tr_intern(N), newClass(vm, tr_intern(N), S))
#define tr_defmodule(N)      TrObject_const_set(vm, vm.self, tr_intern(N), newModule(vm, tr_intern(N)))

#define tr_send(R,MSG,...)   ({ \
  OBJ __argv[] = { (MSG), ##__VA_ARGS__ }; \
  TrObject_send(vm, R, sizeof(__argv)/sizeof(OBJ), __argv); \
})
#define tr_send2(R,STR,...)  tr_send((R), tr_intern(STR), ##__VA_ARGS__)

typedef unsigned long OBJ;
typedef unsigned char u8;

type TrInst uint;

KHASH_MAP_INIT_STR(str, OBJ)
KHASH_MAP_INIT_INT(OBJ, OBJ)

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

struct TrVM;

type TrCallSite struct {
  OBJ class;
  OBJ method;
  OBJ message;
  int method_missing:1;
  size_t miss;
}

type TrUpval struct {
  OBJ *value;
  OBJ closed; /* value when closed */
}

typedef OBJ (TrFunc)(vm *struct TrVM, OBJ receiver, ...);

type TrBinding struct {
	type 			TR_T;
	class			OBJ;
	ivars			*khash_t(OBJ);
  	frame			*Frame;
}

type TrVM struct {
	symbols			*khash_t(str);
	globals			*khash_t(OBJ);
	consts			*khash_t(OBJ);           /* TODO this goes in modules */
	classes			[TR_T_MAX]OBJ;          /* core classes */
	top_frame		*Frame;             /* top level frame */
	frame			*Frame;                 /* current frame */
	cf				int;                         /* current frame number */
	self			OBJ;                       /* root object */
	debug			int;
	throw_reason	int;
	throw_value	OBJ;
  
	// exceptions
	cException			OBJ;
	cScriptError		OBJ;
	cSyntaxError		OBJ;
	cStandardError		OBJ;
	cArgumentError		OBJ;
	cRuntimeError		OBJ;
	cRegexpError		OBJ;
	cTypeError			OBJ;
	cSystemCallError	OBJ;
	cIndexError			OBJ;
	cLocalJumpError		OBJ;
	cSystemStackError	OBJ;
	cNameError			OBJ;
	cNoMethodError		OBJ;
  
	// cached objects
	sADD				OBJ;
	sSUB				OBJ;
	sLT					OBJ;
	sNEG				OBJ;
	sNOT				OBJ;
}

type TrObject struct {
  TR_T type;
  OBJ class;
  khash_t(OBJ) *ivars;
}

type TrString struct {
  TR_T type;
  OBJ class;
  khash_t(OBJ) *ivars;
  char *ptr;
  size_t len;
  int interned:1;
}
type TrSymbol TrString

type TrRange struct {
  TR_T type;
  OBJ class;
  khash_t(OBJ) *ivars;
  OBJ first;
  OBJ last;
  int exclusive;
}

type TrHash struct {
  TR_T type;
  OBJ class;
  khash_t(OBJ) *ivars;
  khash_t(OBJ) *kh;
}

type TrRegexp struct {
  TR_T type;
  OBJ class;
  khash_t(OBJ) *ivars;
  pcre *re;
}

/* vm */
TrVM *TrVM_new();
OBJ TrVM_backtrace(vm *struct TrVM);
OBJ TrVM_eval(vm *struct TrVM, char *code, char *filename);
OBJ TrVM_load(vm *struct TrVM, char *filename);
void TrVM_destroy(TrVM *vm);

/* string */
OBJ TrSymbol_new(vm *struct TrVM, const char *str);
OBJ TrString_new(vm *struct TrVM, const char *str, size_t len);
OBJ TrString_new2(vm *struct TrVM, const char *str);
OBJ TrString_new3(vm *struct TrVM, size_t len);
OBJ TrString_push(vm *struct TrVM, OBJ self, OBJ other);
OBJ tr_sprintf(vm *struct TrVM, const char *fmt, ...);
void TrSymbol_init(vm *struct TrVM);
void TrString_init(vm *struct TrVM);

/* number */
void TrFixnum_init(vm *struct TrVM);

/* hash */
OBJ TrHash_new(vm *struct TrVM);
OBJ TrHash_new2(vm *struct TrVM, size_t n, OBJ items[]);
void TrHash_init(vm *struct TrVM);

/* range */
OBJ TrRange_new(vm *struct TrVM, OBJ start, OBJ end, int exclusive);
void TrRange_init(vm *struct TrVM);

/* object */
OBJ TrObject_alloc(vm *struct TrVM, OBJ class);
int TrObject_type(vm *struct TrVM, OBJ obj);
OBJ TrObject_method(vm *struct TrVM, OBJ self, OBJ name);
OBJ TrObject_send(vm *struct TrVM, OBJ self, int argc, OBJ argv[]);
OBJ TrObject_const_set(vm *struct TrVM, OBJ self, OBJ name, OBJ value);
OBJ TrObject_const_get(vm *struct TrVM, OBJ self, OBJ name);
OBJ TrObject_add_singleton_method(vm *struct TrVM, OBJ self, OBJ name, OBJ method);
void TrObject_preinit(vm *struct TrVM);
void TrObject_init(vm *struct TrVM);

/* kernel */
void TrKernel_init(vm *struct TrVM);

/* primitive */
void TrPrimitive_init(vm *struct TrVM);

/* error */
OBJ TrException_new(vm *struct TrVM, OBJ class, OBJ message);
OBJ TrException_backtrace(vm *struct TrVM, OBJ self);
OBJ TrException_set_backtrace(vm *struct TrVM, OBJ self, OBJ backtrace);
OBJ TrException_default_handler(vm *struct TrVM, OBJ exception);
void TrError_init(vm *struct TrVM);

/* regexp */
OBJ TrRegexp_new(vm *struct TrVM, char *pattern, int options);
void TrRegex_free(vm *struct TrVM, OBJ self);
void TrRegexp_init(vm *struct TrVM);

#ifdef __unix__
  #include <unistd.h>
#else
  #include "freegetopt/getopt.h"
#endif

static int usage() {
  printf("usage: tinyrb [options] [file]\n"
         "options:\n"
         "  -e   eval code\n"
         "  -d   show debug info (multiple times for more)\n"
         "  -v   print version\n"
         "  -h   print this\n");
  return 1;
}

func main(argc int, argv *[]char) {
  int opt;
  TrVM *vm = TrVM_new();

  while((opt = getopt(argc, argv, "e:vdh")) != -1) {
    switch(opt) {
      case 'e':
        TR_FAILSAFE(TrVM_eval(vm, optarg, "<eval>"));
        return 0;
      case 'v':
        printf("tinyrb %s\n", TR_VERSION);
        return 1;
      case 'd':
        vm.debug++;
        continue;
      case 'h':
      default:
        return usage();
    }
  }

  /* These lines allow us to tread argc and argv as though 
   * any switches were not there */
  argc -= optind;
  argv += optind;
  
  if (argc > 0) {
    TR_FAILSAFE(TrVM_load(vm, argv[argc-1]));
    return 0;
  }
  
  TrVM_destroy(vm);
  
  return usage();
}