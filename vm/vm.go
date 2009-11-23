package vm

import (
	"os";						// operating system support
	"fmt";						// formatted I/O

// #include <sys/stat.h>
// #include <assert.h>

	"tr";
	"opcode";
	"internal";
	"call";
)

func TrVM_lookup(vm *struct TrVM, b *Block, receiver, msg OBJ, ip *TrInst) OBJ {
  OBJ method = TrObject_method(vm, receiver, msg);
  if method == TR_UNDEF { return TR_UNDEF }

  TrInst *boing = (ip-1);
  // TODO do not prealloc TrCallSite here, every one is a memory leak and a new one is created on polymorphic calls.
  TrCallSite *s = (kv_pushp(TrCallSite, b.sites));
  s.class = TR_CLASS(receiver);
  s.miss = 0;
  s.method = method;
  s.message = msg;
  if method == TR_NIL {
    s.method = TrObject_method(vm, receiver, tr_intern("method_missing"));
    s.method_missing = 1;
  }
  
  // Implement Monomorphic method cache by replacing the previous instruction (BOING) w/ CACHE that uses the CallSite to find the method instead of doing a full lookup.
  if GET_OPCODE(*boing) == TR_OP_CACHE {
    // Existing call site
    // TODO maybe take existing call site hit miss into consideration to replace it with this one. For now, we just don't replace it, the first one is always the cached one.
  } else {
    // New call site, we cache it fo shizzly!
    SET_OPCODE(*boing, TR_OP_CACHE);
    SETARG_A(*boing, GETARG_A(*ip)); 		// receiver register
    SETARG_B(*boing, 1);					// jmp
    SETARG_C(*boing, kv_size(b.sites)-1);	// CallSite index
  }
  
  return (OBJ)s;
}

func TrVM_defclass(vm *TrVM, name OBJ, b *Block, module int, super OBJ) OBJ {
  OBJ mod = TrObject_const_get(vm, vm.frame.class, name);
  if mod == TR_UNDEF { return TR_UNDEF }
  
  if !mod {
	// new module/class
    if module
      mod = newModule(vm, name);
    else
      mod = newClass(vm, name, super ? super : TR_CORE_CLASS(Object));
    if mod == TR_UNDEF { return TR_UNDEF }
    TrObject_const_set(vm, vm.frame.class, name, mod);
  }
  OBJ ret = TR_NIL;
  TR_WITH_FRAME(mod, mod, 0, {
    ret = TrVM_interpret(vm, vm.frame, b, 0, 0, 0, 0);
  });
  if ret == TR_UNDEF { return TR_UNDEF }
  return mod;
}

func TrVM_interpret_method(vm *struct TrVM, self OBJ, argc int, argv []OBJ) OBJ {
	assert(vm.frame.method);
	Block *b = (Block *)((Method*)vm.frame.method).data;
	if argc != (int)b.argc { tr_raise(ArgumentError, "wrong number of arguments (%d for %lu)", argc, b.argc); }
	return TrVM_interpret(vm, vm.frame, b, 0, argc, argv, 0);
}

func TrVM_interpret_method_with_defaults(vm *struct TrVM, self OBJ, argc int, argv []OBJ) OBJ {
	assert(vm.frame.method);
	Block *b = (Block *)((Method*)vm.frame.method).data;
	int req_argc = b.argc - b.defaults.Len;
	if argc < req_argc { tr_raise(ArgumentError, "wrong number of arguments (%d for %d)", argc, req_argc); }
	if argc > (int)b.argc { tr_raise(ArgumentError, "wrong number of arguments (%d for %lu)", argc, b.argc); }
	int defi = argc - req_argc - 1;		// index in defaults table or -1 for none
	return TrVM_interpret(vm, vm.frame, b, defi < 0 ? 0 : b.defaults.At(defi), argc, argv, 0);
}

func TrVM_interpret_method_with_splat(vm *struct TrVM, self OBJ, argc int, argv []OBJ) OBJ {
	assert(vm.frame.method);
	Block *b = (Block *)((Method*)vm.frame.method).data;
	// TODO support defaults
	assert(b.defaults.Len == 0 && "defaults with splat not supported for now");
	if argc < (int)b.argc-1 { tr_raise(ArgumentError, "wrong number of arguments (%d for %lu)", argc, b.argc-1); }
	argv[b.argc-1] = newArray3(vm, argc - b.argc + 1, &argv[b.argc-1]);
	return TrVM_interpret(vm, vm.frame, b, 0, b.argc, argv, 0);
}

static OBJ TrVM_defmethod(vm *struct TrVM, Frame *f, OBJ name, Block *b, int meta, OBJ receiver) {
  TrFunc *func;
  if b.arg_splat
    func = (TrFunc *) TrVM_interpret_method_with_splat;
  else if b.defaults.Len > 0)
    func = (TrFunc *) TrVM_interpret_method_with_defaults;
  else
    func = (TrFunc *) TrVM_interpret_method;
  OBJ method = newMethod(vm, func, (OBJ)b, -1);
	if method == TR_UNDEF { return TR_UNDEF }
  if (meta)
    TrObject_add_singleton_method(vm, receiver, name, method);
  else
    f.class.add_method(vm, name, method);
  return TR_NIL;
}

func TrVM_yield(vm *TrVM, f *Frame, argc int, argv []OBJ) OBJ {
  Closure *cl = f.closure;
  if !cl { tr_raise(LocalJumpError, "no block given"); }
  OBJ ret = TR_NIL;
  TR_WITH_FRAME(cl.self, cl.class, cl.parent, {
    ret = TrVM_interpret(vm, vm.frame, cl.block, 0, argc, argv, cl);
  });
  return ret;
}

// dispatch macros
#define DISPATCH       (i = *++ip); break

// register access macros
#define OPCODE GET_OPCODE(i)
#define A      GETARG_A(i)
#define B      GETARG_B(i)
#define C      GETARG_C(i)
#define nA     GETARG_A(*(ip+1))
#define nB     GETARG_B(*(ip+1))
#define R      stack
#define RK(X)  (X & (1 << (SIZE_B - 1)) ? k[X & ~0x100] : R[X])
#define Bx     GETARG_Bx(i)
#define sBx    GETARG_sBx(i)
#define SITE   (b.sites.a)

#define RETURN(V) \
  /* TODO GC release everything on the stack before returning */ \
  return (V)

// Interprets the code in b.code. Returns TR_UNDEF on error.
static OBJ TrVM_interpret(vm *TrVM, f *Frame, b *Block, start, argc int, argv []OBJ, closure *Closure) {
	f.stack = alloca(sizeof(OBJ) * b.regc);

	TrInst *ip = b.code.a + start;
	OBJ *stack = f.stack;

	TrInst i = *ip;
	OBJ *k = b.k.a;
	Block **blocks = b.blocks.a;
	f.line = b.line;
	f.filename = b.filename;
	TrUpval *upvals = closure ? closure.upvals : 0;
	TrCallSite *call = 0;

	// transfer locals
	if argc > 0 { 
		assert(argc <= (int)kv_size(b.locals) && "can't fit args in locals");
		TR_MEMCPY_N(stack, argv, OBJ, argc);
	}
  
	for {
		switch(OPCODE) {
			// no-op
			case TR_OP_BOING:

    		// register loading
			case TR_OP_MOVE:
				R[A] = R[B]

			case TR_OP_LOADK:
				R[A] = k[Bx]

			case TR_OP_STRING:
				R[A] = TrString_new2(vm, b.strings.At(Bx)

			case TR_OP_SELF:
				R[A] = f.self

			case TR_OP_NIL:
				R[A] = TR_NIL

			case TR_OP_BOOL:
				R[A] = B

			case TR_OP_NEWARRAY:
				R[A] = newArray3(vm, B, &R[A+1])

			case TR_OP_NEWHASH:
				R[A] = TrHash_new2(vm, B, &R[A+1])

			case TR_OP_NEWRANGE:
				R[A] = TrRange_new(vm, R[A], R[B], C)
    
			// return
			case TR_OP_RETURN:
				RETURN(R[A])

			case TR_OP_THROW:
				vm.throw_reason = A
				vm.throw_value = R[B]
				RETURN(TR_UNDEF)

			case TR_OP_YIELD:
				if (OBJ)(R[A] = TrVM_yield(vm, f, B, &R[A+1])) == TR_UNDEF { RETURN(TR_UNDEF) }
    
    		// variable and consts
    		case TR_OP_SETUPVAL:
				assert(upvals && upvals[B].value)
				*(upvals[B].value) = R[A]

    		case TR_OP_GETUPVAL:
				assert(upvals)
				R[A] = *(upvals[B].value)

    		case TR_OP_SETIVAR:
				TR_KH_SET(TR_COBJECT(f.self).ivars, k[Bx], R[A])

    		case TR_OP_GETIVAR:
				R[A] = TR_KH_GET(TR_COBJECT(f.self).ivars, k[Bx])

    		case TR_OP_SETCVAR:
				TR_KH_SET(TR_COBJECT(f.class).ivars, k[Bx], R[A])

    		case TR_OP_GETCVAR:
				R[A] = TR_KH_GET(TR_COBJECT(f.class).ivars, k[Bx])

    		case TR_OP_SETCONST:
				TrObject_const_set(vm, f.self, k[Bx], R[A])

    		case TR_OP_GETCONST:
				R[A] = TrObject_const_get(vm, f.self, k[Bx])

    		case TR_OP_SETGLOBAL:
				TR_KH_SET(vm.globals, k[Bx], R[A])

    		case TR_OP_GETGLOBAL:
				R[A] = TR_KH_GET(vm.globals, k[Bx])
    
    		// method calling
    		case TR_OP_LOOKUP:
				if (OBJ)(call = (TrCallSite*)TrVM_lookup(vm, b, R[A], k[Bx], ip)) == TR_UNDEF { RETURN(TR_UNDEF) }

    		case TR_OP_CACHE:
				// TODO how to expire cache?
				assert(&SITE[C] && "Method cached but no CallSite found");
				if SITE[C].class == TR_CLASS((R[A])) {
					call = &SITE[C]
					ip += B
				} else {
					// TODO invalidate CallSite if too much miss.
        			SITE[C].miss++
				}

			case TR_OP_CALL:
				Closure *cl = 0;
				TrInst ci = i;

				if C > 0 {
					// Get upvalues using the pseudo-instructions following the CALL instruction.
					//	Eg.: there's one upval to a local (x) to be passed:
					//	call    0  0  0
					//	move    0  0  0 ; this is not executed
					//	return  0

					cl = newClosure(vm, blocks[C-1], f.self, f.class, f.closure);
					size_t n, nupval = kv_size(cl.block.upvals);
					for (n = 0; n < nupval; ++n) {
						(i = *++ip)
						if OPCODE == TR_OP_MOVE {
							cl.upvals[n].value = &R[B];
						} else {
							assert(OPCODE == TR_OP_GETUPVAL);
							cl.upvals[n].value = upvals[B].value;
						}
					}
				}
				int argc = GETARG_B(ci) >> 1;
				OBJ *argv = &R[GETARG_A(ci)+2];
				if call.method_missing {
					argc++;
					*(--argv) = call.message;
				}
				OBJ ret = call.method.call(vm,
											R[GETARG_A(ci)],		// receiver
											argc, argv,
											GETARG_B(ci) & 1,		// splat
											cl						// closure
											);

				// Handle throw if some.
				// A "throw" is done by returning TR_UNDEF to exit a current call frame (Frame)
				// until one handle it by returning are real value or continuing execution.
				// Non-local returns and exception propagation are implemented this way.
				// Rubinius and Python do it this way. */

				if ret == TR_UNDEF {
					switch vm.throw_reason {
						case TR_THROW_EXCEPTION:
							// TODO run rescue and stop propagation if rescued
							// TODO run ensure block
            				RETURN(TR_UNDEF)

						case TR_THROW_RETURN:
							// TODO run ensure block
            				if f.closure { RETURN(TR_UNDEF) }
            				RETURN(vm.throw_value)

						case TR_THROW_BREAK:

          				default:
							assert(0 && "BUG: invalid throw_reason");
					}
				}
      			R[GETARG_A(ci)] = ret
    
			// definition
			case TR_OP_DEF:
				if (OBJ)TrVM_defmethod(vm, f, k[Bx], blocks[A], 0, 0) == TR_UNDEF { RETURN(TR_UNDEF) }

			case TR_OP_METADEF:
				if (OBJ)TrVM_defmethod(vm, f, k[Bx], blocks[A], 1, R[nA]) == TR_UNDEF { RETURN(TR_UNDEF) }
				ip++

			case TR_OP_CLASS:
				if (OBJ)TrVM_defclass(vm, k[Bx], blocks[A], 0, R[nA]) == TR_UNDEF { RETURN(TR_UNDEF) }
				ip++

			case TR_OP_MODULE:
				if (OBJ)TrVM_defclass(vm, k[Bx], blocks[A], 1, 0) == TR_UNDEF { RETURN(TR_UNDEF) }
    
			// jumps
			case TR_OP_JMP:
				ip += sBx

			case TR_OP_JMPIF:
				if ( TR_TEST(R[A])) ip += sBx

			case TR_OP_JMPUNLESS:
				if (!TR_TEST(R[A])) ip += sBx

    		// arithmetic optimizations
    		// TODO cache lookup in tr_send and force send if method was redefined
			case TR_OP_ADD:
				OBJ rb = RK(B)
				if TR_IS_FIX(rb) {
					R[A] = TR_INT2FIX(TR_FIX2INT(rb) + TR_FIX2INT(RK(C)))
				} else {
					R[A] = tr_send(rb, vm.sADD, RK(C))
				}

			case TR_OP_SUB:
				OBJ rb = RK(B)
				if TR_IS_FIX(rb) {
					R[A] = TR_INT2FIX(TR_FIX2INT(rb) - TR_FIX2INT(RK(C)))
				} else {
					R[A] = tr_send(rb, vm.sSUB, RK(C))
				}

			case TR_OP_LT:
				OBJ rb = RK(B)
				if TR_IS_FIX(rb) {
					R[A] = TR_BOOL(TR_FIX2INT(rb) < TR_FIX2INT(RK(C)))
				} else {
					R[A] = tr_send(rb, vm.sLT, RK(C))
				}

			case TR_OP_NEG:
				OBJ rb = RK(B)
				if TR_IS_FIX(rb) {
					R[A] = TR_INT2FIX(-TR_FIX2INT(rb))
				} else {
					R[A] = tr_send(rb, vm.sNEG, RK(C))
				}

			case TR_OP_NOT:
				OBJ rb = RK(B)
				R[A] = TR_BOOL(!TR_TEST(rb))

			default:
				// if there are unknown opcodes in the stream then halt the VM
				// TODO: we need a better error message
				fmt.Println("unknown opcode:", (int)OPCODE)
				os.Exit(1)
		}
		DISPATCH
	}
}

/* returns the backtrace of the current call frames */
OBJ TrVM_backtrace(vm *struct TrVM) {
  OBJ backtrace = newArray(vm);
  
  if (!vm.frame) return backtrace;
  
  /* skip a frame since it's the one doing the raising */
  Frame *f = vm.frame.previous;
  while (f) {
    OBJ str;
    char *filename = f.filename ? TR_STR_PTR(f.filename) : "?";
    if (f.method)
      str = tr_sprintf(vm, "\tfrom %s:%lu:in `%s'",
                       filename, f.line, TR_STR_PTR(((Method *)f.method).name));
    else
      str = tr_sprintf(vm, "\tfrom %s:%lu",
                       filename, f.line);
    TR_ARRAY_PUSH(backtrace, str);
    
    f = f.previous;
  }
  
  return backtrace;
}

OBJ TrVM_eval(vm *struct TrVM, char *code, char *filename) {
  Block *b = Block_compile(vm, code, filename, 0);
  if (!b) return TR_UNDEF;
  if (vm.debug) { b.dump(vm); }
  return TrVM_run(vm, b, vm.self, TR_CLASS(vm.self), 0, 0);
}

OBJ TrVM_load(vm *struct TrVM, char *filename) {
  FILE *fp;
  struct stat stats;
  
  if (stat(filename, &stats) == -1) tr_raise_errno(filename);
  fp = fopen(filename, "rb");
  if (!fp) tr_raise_errno(filename);
  
  char *string = TR_ALLOC_N(char, stats.st_size + 1);
  if (fread(string, 1, stats.st_size, fp) == stats.st_size)
    return TrVM_eval(vm, string, filename);
  
  tr_raise_errno(filename);
  return TR_NIL;
}

func TrVM_run(vm *TrVM, b *Block, self, class OBJ, argc int, argv []OBJ) OBJ {
  OBJ ret = TR_NIL;
  TR_WITH_FRAME(self, class, 0, {
    ret = TrVM_interpret(vm, vm.frame, b, 0, argc, argv, 0);
  });
  return ret;
}

TrVM *TrVM_new() {
  GC_INIT();

  TrVM *vm = TR_ALLOC(TrVM);
  vm.symbols = kh_init(str);
  vm.globals = kh_init(OBJ);
  vm.consts = kh_init(OBJ);
  vm.debug = 0;
  
  /* bootstrap core classes,
     order is important here, so careful, mkay? */
  TrMethod_init(vm);
  TrSymbol_init(vm);
  TrModule_init(vm);
  TrClass_init(vm);
  TrObject_preinit(vm);
  Class *symbolc = (Class*)TR_CORE_CLASS(Symbol);
  Class *modulec = (Class*)TR_CORE_CLASS(Module);
  Class *classc = (Class*)TR_CORE_CLASS(Class);
  Class *methodc = (Class*)TR_CORE_CLASS(Method);
  Class *objectc = (Class*)TR_CORE_CLASS(Object);
  /* set proper superclass has Object is defined last */
  symbolc.super = modulec.super = methodc.super = (OBJ)objectc;
  classc.super = (OBJ)modulec;
  /* inject core classes metaclass */
  symbolc.class = newMetaClass(vm, objectc.class);
  modulec.class = newMetaClass(vm, objectc.class);
  classc.class = newMetaClass(vm, objectc.class);
  methodc.class = newMetaClass(vm, objectc.class);
  objectc.class = newMetaClass(vm, objectc.class);
  
  /* Some symbols are created before Object, so make sure all have proper class. */
  TR_KH_EACH(vm.symbols, i, sym, {
    TR_COBJECT(sym).class = (OBJ)symbolc;
  });
  
  /* bootstrap rest of core classes, order is no longer important here */
  TrObject_init(vm);
  TrError_init(vm);
  TrBinding_init(vm);
  TrPrimitive_init(vm);
  TrKernel_init(vm);
  TrString_init(vm);
  TrFixnum_init(vm);
  TrArray_init(vm);
  TrHash_init(vm);
  TrRange_init(vm);
  TrRegexp_init(vm);
  
  vm.self = TrObject_alloc(vm, 0);
  vm.cf = -1;
  
  /* cache some commonly used values */
  vm.sADD = tr_intern("+");
  vm.sSUB = tr_intern("-");
  vm.sLT = tr_intern("<");
  vm.sNEG = tr_intern("@-");
  vm.sNOT = tr_intern("!");
  
  TR_FAILSAFE(TrVM_load(vm, "lib/boot.rb"));
  
  return vm;
}

void TrVM_destroy(TrVM *vm) {
  kh_destroy(str, vm.symbols);
  GC_gcollect();
}