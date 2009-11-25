/* Inlined functions frequently used in the method calling process. */

#include <alloca.h>

type Frame struct {
	closure					*Closure;
	method					*Method;				// current called method
	stack, upvals			*OBJ;
	self, class, filename	OBJ;
	line					size_t;
	previous				*Frame;
}


#define TR_WITH_FRAME(SELF,CLASS,CLOS,BODY) ({ \
  /* push a frame */ \
  if (++vm.cf >= TR_MAX_FRAMES) { tr_raise(SystemStackError, "Stack overflow"); } \
  Frame __f; \
  Frame *__fp = &__f; \
  __f.previous = vm.frame; \
  if (vm.cf == 0) vm.top_frame = __fp; \
  vm.frame = __fp; \
  vm.throw_reason = vm.throw_value = 0; \
  __f.method = NULL; \
  __f.filename = __f.line = 0; \
  __f.self = SELF; \
  __f.class = CLASS; \
  __f.closure = CLOS; \
  /* execute BODY inside the frame */ \
  BODY \
  /* pop the frame */ \
  vm.cf--; \
  vm.frame = vm.frame.previous; \
})

func (self *Method) call(vm *TrVM, receiver OBJ, argc int, args *OBJ, splat int, cl *Closure) OBJ {
  OBJ ret = TR_NIL;
  Frame *f = nil;

  TR_WITH_FRAME(receiver, TR_CLASS(receiver), cl, {
    f = vm.frame;
    Method *m = f.method = (Method *) self;
    TrFunc *func = f.method.func;

    /* splat last arg is needed */
    if splat {
      splated := args[argc-1];
      splatedn := splated.kv.Len();
      OBJ *new_args = TR_ALLOC_N(OBJ, argc);
      TR_MEMCPY_N(new_args, args, OBJ, argc-1);
      TR_MEMCPY_N(new_args + argc-1, &splated.kv.At(0), OBJ, splatedn);
      argc += splatedn-1;
      args = new_args;
    }
  
    if (m.arity == -1) {
      ret = func(vm, receiver, argc, args);
    } else {
      if (m.arity != argc) tr_raise(ArgumentError, "Expected %d arguments, got %d.", f.method.arity, argc);
      switch (argc) {
        case 0:  ret = func(vm, receiver); break;
        case 1:  ret = func(vm, receiver, args[0]); break;
        case 2:  ret = func(vm, receiver, args[0], args[1]); break;
        case 3:  ret = func(vm, receiver, args[0], args[1], args[2]); break;
        case 4:  ret = func(vm, receiver, args[0], args[1], args[2], args[3]); break;
        case 5:  ret = func(vm, receiver, args[0], args[1], args[2], args[3], args[4]); break;
        case 6:  ret = func(vm, receiver, args[0], args[1], args[2], args[3], args[4], args[5]); break;
        case 7:  ret = func(vm, receiver, args[0], args[1], args[2], args[3], args[4], args[5], args[6]); break;
        case 8:  ret = func(vm, receiver, args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7]); break;
        case 9:  ret = func(vm, receiver, args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8]); break;
        case 10: ret = func(vm, receiver, args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8], args[9]); break;
        default: tr_raise(ArgumentError, "Too much arguments: %d, max is %d for now.", argc, 10);
      }
    }
  });
  
  return ret;
}
