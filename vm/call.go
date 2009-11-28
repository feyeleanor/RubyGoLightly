#include <alloca.h>

import (
	"bytes";
	)

func (self *Method) call(vm *RubyVM, receiver *RubyObject, argc int, args []RubyObject, splat int, closure *Closure) RubyObject {
	if TR_IMMEDIATE(receiver) {
		receiver_class := vm.classes[Object_type(vm, receiver)];
	} else {
		receiver_class := Object *(receiver).class;
	}

	// push a frame
	vm.cf++;
	if vm.cf >= TR_MAX_FRAMES {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cSystemStackError, tr_sprintf(vm, "Stack overflow"));
		return TR_UNDEF;
	}

	frame := newFrame(receiver, receiver_class, closure);
	if vm.cf == 0 { vm.top_frame = frame; }
	vm.frame = frame;
	vm.throw_reason = vm.throw_value = 0;

	// execute BODY inside the frame
	method := frame.method = self;
	function = frame.method.func;

	// splat last arg is needed
	if splat {
		splated := args[argc - 1];
		splatedn := splated.kv.Len();
		new_args := make([]OBJ, argc)
		memcpy(new_args, args, sizeof(OBJ) * (argc - 1));
		memcpy(new_args + argc - 1, &splated.values.At(0), sizeof(OBJ) * splatedn);
		argc += splatedn-1;
		args = new_args;
	}

	if (m.arity == -1) {
		result := function(vm, receiver, argc, args);
	} else {
		if method.arity != argc {
			vm.throw_reason = TR_THROW_EXCEPTION;
			vm.throw_value = TrException_new(vm, vm.cArgumentError, tr_sprintf(vm, "Expected %d arguments, got %d.", frame.method.arity, argc));
			return TR_UNDEF;
		}
		switch argc {
			case 0:  result := function(vm, receiver);
			case 1:  result := function(vm, receiver, args[0]);
			case 2:  result := function(vm, receiver, args[0], args[1]);
			case 3:  result := function(vm, receiver, args[0], args[1], args[2]);
			case 4:  result := function(vm, receiver, args[0], args[1], args[2], args[3]);
			case 5:  result := function(vm, receiver, args[0], args[1], args[2], args[3], args[4]);
			case 6:  result := function(vm, receiver, args[0], args[1], args[2], args[3], args[4], args[5]);
			case 7:  result := function(vm, receiver, args[0], args[1], args[2], args[3], args[4], args[5], args[6]);
			case 8:  result := function(vm, receiver, args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7]);
			case 9:  result := function(vm, receiver, args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8]);
			case 10: result := function(vm, receiver, args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8], args[9]);
			default:
				vm.throw_reason = TR_THROW_EXCEPTION;
				vm.throw_value = TrException_new(vm, vm.cArgumentError, tr_sprintf(vm, "Too many arguments: %d, max is %d for now.", argc, 10));
				return TR_UNDEF;
		}
	}

	// pop the frame
	vm.cf--;
	vm.frame = vm.frame.previous;
	return result;
}