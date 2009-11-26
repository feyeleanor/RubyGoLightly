package RubyVM

import (
	"os";						// operating system support
	"fmt";						// formatted I/O

// #include <sys/stat.h>
// #include <assert.h>
	"bytes";
	"tr";
	"opcode";
	"internal";
	"call";
)

type RubyVM struct {
	symbols				*map[string] string;
	globals				*map[string] OBJ;
	consts				*map[string] OBJ;				// TODO this goes in modules
	classes				[TR_T_MAX]OBJ;					// core classes
	top_frame			*Frame;							// top level frame
	frame				*Frame;							// current frame
	cf					int;							// current frame number
	self				OBJ;							// root object
	debug				int;
	throw_reason		int;
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

func (vm *RubyVM) lookup(block *Block, receiver, msg OBJ, ip *MachineOP) OBJ {
	method := TrObject_method(vm, receiver, msg);
	if method == TR_UNDEF { return TR_UNDEF }

	boing = *(ip - 1);
	// TODO do not prealloc TrCallSite here, every one is a memory leak and a new one is created on polymorphic calls.
	TrCallSite *s = (kv_pushp(TrCallSite, block.sites));
	s.class = TR_CLASS(receiver);
	s.miss = 0;
	s.method = method;
	s.message = msg;
	if method == TR_NIL {
		s.method = TrObject_method(vm, receiver, tr_intern("method_missing"));
		s.method_missing = 1;
	}
  
	// Implement Monomorphic method cache by replacing the previous instruction (BOING) w/ CACHE that uses the CallSite to find the method instead of doing a full lookup.
	if (*boing).OpCode == TR_OP_CACHE {
		// Existing call site
		// TODO maybe take existing call site hit miss into consideration to replace it with this one. For now, we just don't replace it, the first one is always the cached one.
	} else {
		// New call site, we cache it fo shizzly!
		(*boing).OpCode = TR_OP_CACHE;
		(*boing).A = (*ip).A; 						// receiver register
		(*boing).B = 1;								// jmp
		(*boing).C = block.sites.Len() - 1;			// CallSite index
	}
	return OBJ(s);
}

func (vm *RubyVM) defclass(name OBJ, block *Block, module int, super OBJ) OBJ {
	mod := TrObject_const_get(vm, vm.frame.class, name);
	if mod == TR_UNDEF { return TR_UNDEF }
  
	if !mod {
		// new module/class
		if module {
			mod := newModule(vm, name);
		} else {
			mod := newClass(vm, name, super ? super : TR_CORE_CLASS(Object));
		}
		if mod == TR_UNDEF { return TR_UNDEF }
		TrObject_const_set(vm, vm.frame.class, name, mod);
	}
	ret := TR_NIL;
	TR_WITH_FRAME(mod, mod, 0, { ret := vm.interpret(vm.frame, block, 0, 0, 0); });
	if ret == TR_UNDEF { return TR_UNDEF }
	return mod;
}

func (vm *RubyVM) interpret_method(self OBJ, args []OBJ) OBJ {
	assert(vm.frame.method);
	block := Block *(vm.frame.method.data);
	if args.Len() != block.argc { tr_raise(ArgumentError, "wrong number of arguments (%d for %lu)", args.Len(), block.argc); }
	return vm.interpret(vm, vm.frame, block, 0, args, 0);
}

func (vm *RubyVM) interpret_method_with_defaults(self OBJ, args []OBJ) OBJ {
	assert(vm.frame.method);
	block := Block *(vm.frame.method.data);
	req_argc := block.argc - block.defaults.Len();
	if args.Len() < req_argc { tr_raise(ArgumentError, "wrong number of arguments (%d for %d)", args.Len(), req_argc); }
	if args.Len() > block.argc { tr_raise(ArgumentError, "wrong number of arguments (%d for %lu)", args.Len(), b.argc); }
	// index in defaults table or -1 for none
	if (i := args.Len() - req_argc - 1) < 0 {
		return vm.interpret(vm.frame, block, 0, args, 0);
	} else {
		return vm.interpret(vm.frame, block, block.defaults.At(i), args, 0);
	}
}

func (vm *RubyVM) interpret_method_with_splat(self OBJ, args []OBJ) OBJ {
	assert(vm.frame.method);
	block := Block *(vm.frame.method.data);
	// TODO support defaults
	assert(block.defaults.Len() == 0 && "defaults with splat not supported for now");
	if args.Len() < b.argc - 1 { tr_raise(ArgumentError, "wrong number of arguments (%d for %lu)", args.Len(), block.argc - 1); }
	argv[block.argc - 1] = newArray3(vm, args.Len() - block.argc + 1, &argv[block.argc - 1]);
	return vm.interpret(vm.frame, block, 0, args[0:block.argc - 1], 0);
}

func (vm *RubyVM) defmethod(frame *Frame, name OBJ, block *Block, meta bool, receiver OBJ) OBJ {
	switch {
		case block.arg_splat:			interpreter := TrFunc *(TrVM_interpret_method_with_splat);
		case block.defaults.Len() > 0:	interpreter := TrFunc *(TrVM_interpret_method_with_defaults);
		default:						interpreter := TrFunc *(TrVM_interpret_method);
	}
	method := newMethod(vm, interpreter, OBJ(block), -1);
	if method == TR_UNDEF { return TR_UNDEF }
	if meta {
		TrObject_add_singleton_method(vm, receiver, name, method);
	} else {
		frame.class.add_method(vm, name, method);
	}
	return TR_NIL;
}

func (vm *RubyVM) yield(frame *Frame, args []OBJ) OBJ {
	closure := frame.closure;
	if !closure { tr_raise(LocalJumpError, "no block given"); }
	ret := TR_NIL;
	TR_WITH_FRAME(closure.self, closure.class, closure.parent, { ret = vm.interpret(vm.frame, closure.block, 0, args, closure); });
	return ret;
}

// Interprets the code in b.code. Returns TR_UNDEF on error.
func (vm *RubyVM) TrVM_interpret(frame *Frame, block *Block, start, args []OBJ, closure *Closure) OBJ {
	frame.stack := make([]OBJ, block.regc);
	ip := *block.code.a + start;
	stack := frame.stack;

	i := *ip;
	OBJ *k = block.k.a;
	Block **blocks = block.blocks.a;
	frame.line = block.line;
	frame.filename = block.filename;
	TrUpval *upvals = closure ? closure.upvals : 0;
	TrCallSite *call = 0;

	// transfer locals
	if args.Len() > 0 { 
		assert(args.Len() <= block.locals.Len() && "can't fit args in locals");
		bytes.Add(stack, args);
	}
  
	for {
		switch i.OpCode {
			// no-op
			case TR_OP_BOING:

    		// register loading
			case TR_OP_MOVE:
				stack[i.A] = stack[i.B]

			case TR_OP_LOADK:
				stack[i.A] = k[i.Get_Bx()]

			case TR_OP_STRING:
				stack[i.A] = TrString_new2(vm, block.strings.At(i.Get_Bx())

			case TR_OP_SELF:
				stack[i.A] = f.self

			case TR_OP_NIL:
				stack[i.A] = TR_NIL

			case TR_OP_BOOL:
				stack[i.A] = i.B

			case TR_OP_NEWARRAY:
				stack[i.A] = newArray3(vm, i.B, &stack[i.A + 1])

			case TR_OP_NEWHASH:
				stack[i.A] = TrHash_new2(vm, i.B, &stack[i.A + 1])

			case TR_OP_NEWRANGE:
				stack[i.A] = TrRange_new(vm, stack[i.A], stack[i.B], i.C)
    
			// return
			case TR_OP_RETURN:
				return stack[i.A];

			case TR_OP_THROW:
				vm.throw_reason = i.A;
				vm.throw_value = stack[i.B]
				return TR_UNDEF;

			case TR_OP_YIELD:
				if OBJ(stack[i.A] = vm.yield(frame, i.B, &stack[i.A + 1])) == TR_UNDEF { return TR_UNDEF; }
    
    		// variable and consts
    		case TR_OP_SETUPVAL:
				assert(upvals && upvals[i.B].value)
				*(upvals[i.B].value) = stack[i.A]

    		case TR_OP_GETUPVAL:
				assert(upvals)
				stack[i.A] = *(upvals[i.B].value)

    		case TR_OP_SETIVAR:
				TR_KH_SET(TR_COBJECT(frame.self).ivars, k[i.Get_Bx()], stack[i.A])

    		case TR_OP_GETIVAR:
				stack[i.A] = TR_KH_GET(TR_COBJECT(frame.self).ivars, k[i.Get_Bx()])

    		case TR_OP_SETCVAR:
				TR_KH_SET(TR_COBJECT(frame.class).ivars, k[i.Get_Bx()], stack[i.A])

    		case TR_OP_GETCVAR:
				stack[i.A] = TR_KH_GET(TR_COBJECT(frame.class).ivars, k[i.Get_Bx()])

    		case TR_OP_SETCONST:
				TrObject_const_set(vm, frame.self, k[i.Get_Bx()], stack[i.A])

    		case TR_OP_GETCONST:
				stack[i.A] = TrObject_const_get(vm, frame.self, k[i.Get_Bx()])

    		case TR_OP_SETGLOBAL:
				TR_KH_SET(vm.globals, k[i.Get_Bx()], stack[i.A])

    		case TR_OP_GETGLOBAL:
				stack[i.A] = TR_KH_GET(vm.globals, k[i.Get_Bx()])
    
    		// method calling
    		case TR_OP_LOOKUP:
				if OBJ(call = TrCallSite *(vm.lookup(block, stack[i.A], k[i.Get_Bx()], ip))) == TR_UNDEF { return TR_UNDEF; }

    		case TR_OP_CACHE:
				// TODO how to expire cache?
				assert(&block.sites.a[i.C] && "Method cached but no CallSite found");
				if block.sites.a[i.C].class == TR_CLASS((stack[i.A])) {
					call = &block.sites.a[i.C]
					ip += i.B;
				} else {
					// TODO invalidate CallSite if too much miss.
        			block.sites.a[i.C].miss++
				}

			case TR_OP_CALL:
				Closure *cl = 0;
				ci := i;

				if i.C > 0 {
					// Get upvalues using the pseudo-instructions following the CALL instruction.
					//	Eg.: there's one upval to a local (x) to be passed:
					//	call    0  0  0
					//	move    0  0  0 ; this is not executed
					//	return  0

					cl = newClosure(vm, blocks[i.C - 1], frame.self, frame.class, frame.closure);
					size_t n, nupval = cl.block.upvals.Len();
					for (n = 0; n < nupval; ++n) {
						(i = *++ip)
						if i.OpCode == TR_OP_MOVE {
							cl.upvals[n].value = &stack[i.B];
						} else {
							assert(i.OpCode == TR_OP_GETUPVAL);
							cl.upvals[n].value = upvals[i.B].value;
						}
					}
				}
				argc := ci.B >> 1;
				argv := &stack[ci.A + 2];
				if call.method_missing {
					argc++;
					*(--argv) = call.message;
				}
				ret := call.method.call(vm,
										stack[ci.A],		// receiver
										argc, argv,
										ci.B & 1,		// splat
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
            				return TR_UNDEF;

						case TR_THROW_RETURN:
							// TODO run ensure block
            				if frame.closure { return TR_UNDEF; }
            				return vm.throw_value;

						case TR_THROW_BREAK:

          				default:
							assert(0 && "BUG: invalid throw_reason");
					}
				}
      			stack[ci.A] = ret
    
			// definition
			case TR_OP_DEF:
				if OBJ(vm.defmethod(frame, k[i.Get_Bx()], blocks[i.A], 0, 0)) == TR_UNDEF { return TR_UNDEF; }

			case TR_OP_METADEF:
				if OBJ(vm.defmethod(frame, k[i.Get_Bx()], blocks[i.A], 1, stack[(*(ip + 1)).A])) == TR_UNDEF { return TR_UNDEF; }
				ip++

			case TR_OP_CLASS:
				if OBJ(vm.defclass(k[i.Get_Bx()], blocks[i.A], 0, stack[(*(ip + 1)).A])) == TR_UNDEF { return TR_UNDEF; }
				ip++

			case TR_OP_MODULE:
				if OBJ(vm.defclass(k[i.Get_Bx()], blocks[i.A], 1, 0)) == TR_UNDEF { return TR_UNDEF; }
    
			// jumps
			case TR_OP_JMP:
				ip += i.Get_sBx();

			case TR_OP_JMPIF:
				if TR_TEST(stack[i.A]) { ip += i.Get_sBx(); }

			case TR_OP_JMPUNLESS:
				if !TR_TEST(stack[i.A]) { ip += i.Get_sBx(); }

    		// arithmetic optimizations
    		// TODO cache lookup in tr_send and force send if method was redefined
			case TR_OP_ADD:
				if i.B & (1 << (SIZE_B - 1) {
					rb := k[i.B & ~0x100]
				} else {
					rb := stack[i.B]
				}

				if i.C & (1 << (SIZE_B - 1) {
					rc := k[i.C & ~0x100]
				} else {
					rc := stack[i.C]
				}

				if TR_IS_FIX(rb) {
					stack[i.A] = TR_INT2FIX(TR_FIX2INT(rb) + TR_FIX2INT(rc))
				} else {
					stack[i.A] = tr_send(rb, vm.sADD, rc)
				}

			case TR_OP_SUB:
				if i.B & (1 << (SIZE_B - 1) {
					rb := k[i.B & ~0x100]
				} else {
					rb := stack[i.B]
				}

				if i.C & (1 << (SIZE_B - 1) {
					rc := k[i.C & ~0x100]
				} else {
					rc := stack[i.C]
				}

				if TR_IS_FIX(rb) {
					stack[i.A] = TR_INT2FIX(TR_FIX2INT(rb) - TR_FIX2INT(rc))
				} else {
					stack[i.A] = tr_send(rb, vm.sSUB, rc)
				}

			case TR_OP_LT:
				if i.B & (1 << (SIZE_B - 1) {
					rb := k[i.B & ~0x100]
				} else {
					rb := stack[i.B]
				}

				if i.C & (1 << (SIZE_B - 1) {
					rc := k[i.C & ~0x100]
				} else {
					rc := stack[i.C]
				}

				if TR_IS_FIX(rb) {
					stack[i.A] = TR_BOOL(TR_FIX2INT(rb) < TR_FIX2INT(rc))
				} else {
					stack[i.A] = tr_send(rb, vm.sLT, rc)
				}

			case TR_OP_NEG:
				if i.B & (1 << (SIZE_B - 1) {
					rb := k[i.B & ~0x100]
				} else {
					rb := stack[i.B]
				}
				if TR_IS_FIX(rb) {
					stack[i.A] = TR_INT2FIX(-TR_FIX2INT(rb))
				} else {
					if i.C & (1 << (SIZE_B - 1) {
						rc := k[i.C & ~0x100]
					} else {
						rc := stack[i.C]
					}
					stack[i.A] = tr_send(rb, vm.sNEG, rc)
				}

			case TR_OP_NOT:
				if i.B & (1 << (SIZE_B - 1) {
					rb := k[i.B & ~0x100]
				} else {
					rb := stack[i.B]
				}
				stack[i.A] = TR_BOOL(!TR_TEST(rb))

			default:
				// if there are unknown opcodes in the stream then halt the VM
				// TODO: we need a better error message
				fmt.Println("unknown opcode:", i.OpCode)
				os.Exit(1)
		}
		i = *++ip;
		break;
	}
}

/* returns the backtrace of the current call frames */
func (vm *RubyVM) backtrace() OBJ {
	backtrace := newArray(vm);
	if vm.frame {
		// skip a frame since it's the one doing the raising
		frame := vm.frame.previous;
		while (frame) {
			if frame.filename {
				filename := TR_STR_PTR(f.filename);
			} else {
				filename := "?"
			}
			if frame.method {
				context := tr_sprintf(vm, "\tfrom %s:%lu:in `%s'", filename, f.line, TR_STR_PTR(((Method *)f.method).name));
			} else {
				context := tr_sprintf(vm, "\tfrom %s:%lu", filename, f.line);
			}
			backtrace.kv.Push(context);
			frame = frame.previous;
		}
	}
	return backtrace;
}

func (vm *RubyVM) eval(code *string, filename *string) OBJ {
	if block := Block_compile(vm, code, filename, 0) {
		if (vm.debug) { block.dump(vm); }
		return vm.run(block, vm.self, TR_CLASS(vm.self), nil);
	} else {
		return TR_UNDEF;
	}
}

func (vm *RubyVM) load(filename *string) OBJ {
	stats = new(stat);
	if stat(filename, &stats) == -1 { tr_raise_errno(filename); }

	if fp := fopen(filename, "rb") {
		s := make([]byte, stats.st_size + 1);
		if fread(s, 1, stats.st_size, fp) == stats.st_size { return vm.eval(s, filename); }
		tr_raise_errno(filename);
	} else {
		tr_raise_errno(filename);
	}
	return TR_NIL;
}

func (vm *RubyVM) run(block *Block, self, class OBJ, args []OBJ) OBJ {
 	ret := TR_NIL;
	TR_WITH_FRAME(self, class, 0, { ret := vm.interpret(vm.frame, block, 0, args, 0); });
	return ret;
}

func newRubyVM() *RubyVM {
	vm := new(RubyVM);
	vm.symbols = kh_init(str);
	vm.globals = kh_init(OBJ);
	vm.consts = kh_init(OBJ);
	vm.debug = 0;
  
	// bootstrap core classes, order is important here, so careful, mkay?
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
 	// set proper superclass has Object is defined last
	symbolc.super = modulec.super = methodc.super = OBJ(objectc);
	classc.super = OBJ(modulec);
	// inject core classes metaclass
	symbolc.class = newMetaClass(vm, objectc.class);
	modulec.class = newMetaClass(vm, objectc.class);
	classc.class = newMetaClass(vm, objectc.class);
	methodc.class = newMetaClass(vm, objectc.class);
	objectc.class = newMetaClass(vm, objectc.class);
  
 	// Some symbols are created before Object, so make sure all have proper class.
	TR_KH_EACH(vm.symbols, i, sym, { TR_COBJECT(sym).class = OBJ(symbolc); });
  
	// bootstrap rest of core classes, order is no longer important here
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
  
 	// cache some commonly used values
	vm.sADD = tr_intern("+");
	vm.sSUB = tr_intern("-");
	vm.sLT = tr_intern("<");
	vm.sNEG = tr_intern("@-");
	vm.sNOT = tr_intern("!");
  
	TR_FAILSAFE(vm.load("lib/boot.rb"));
	return vm;
}