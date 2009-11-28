import (
	"pcre";
	"tr";
	)

// Loosely based on http://vcs.pcre.org/viewvc/code/trunk/pcredemo.c

// Translate this to use Go's stdlib regexp package

func TrRegexp_new(vm *RubyVM, pattern *string, options int) RubyObject {
	r := Regexp{type: TR_T_Regexp, class: vm.classes[TR_T_Regexp], ivars: make(map[string] RubyObject)};
	error *string;
	erroffset int;
  
	r.re = pcre_compile(
		pattern,              /* the pattern */
		options,              /* default options */
		&error,               /* for error message */
		&erroffset,           /* for error offset */
		nil);                /* use default character tables */
  
	if (r.re == nil) {
		TrRegex_free(vm, r);
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cRegexpError, tr_sprintf(vm, "compilation failed at offset %d: %s", erroffset, error));
		return TR_UNDEF;
	}
	return r;
}

func TrRegexp_compile(vm *RubyVM, self, pattern *RubyObject) RubyObject {
	if !pattern.(String) && !pattern.(Symbol) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + pattern));
		return TR_UNDEF;
	}
	return TrRegexp_new(vm, pattern.ptr, 0);
}

#define OVECCOUNT 30    /* should be a multiple of 3 */

func TrRegexp_match(vm *RubyVM, self, str *RubyObject) RubyObject {
	if !self.(Regexp) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected Regexp"));
		return TR_UNDEF;
	}
	if !str.(String) && !str.(Symbol) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + str));
		return TR_UNDEF;
	}
	r := TrRegexp *(self);
	subject := str.ptr;
	rc int;
	ovector [OVECCOUNT]int;

	rc = pcre_exec(
		r.re,                /* the compiled pattern */
		NULL,                 /* no extra data - we didn't study the pattern */
		subject,              /* the subject string */
		str.len,      		/* the length of the subject */
		0,                    /* start at offset 0 in the subject */
		0,                    /* default options */
		ovector,              /* output vector for substring information */
		OVECCOUNT);           /* number of elements in the output vector */
  
	if (rc < 0) return TR_NIL;

	if (rc == 0) {
		rc = OVECCOUNT/3;
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cRegexpError, tr_sprintf(vm, "Too many matches, only %d supported for now", rc - 1));
		return TR_UNDEF;
	}
  
	// TODO should create a MatchData object
	data := vm.newArray();
	i int;
	for (i = 0; i < rc; i++) {
		substring_start := subject + ovector[2*i];
		substring_length := ovector[2*i+1] - ovector[2*i];
		data.Push(TrString_new(vm, substring_start, substring_length));
	}
	return data;
}

func TrRegex_free(vm *RubyVM, self *RubyObject) {
	r := TrRegexp *(self);
	pcre_free(r.re);
}

func TrRegexp_init(vm *RubyVM) {
	c := vm.classes[TR_T_Regexp] = Object_const_set(vm, vm.self, TrSymbol_new(vm, Regexp), newClass(vm, TrSymbol_new(vm, Regexp), vm.classes[TR_T_Object]));
	Object_add_singleton_method(vm, c, TrSymbol_new(vm, "new"), newMethod(vm, (TrFunc *)TrRegexp_compile, TR_NIL, 1));
	c.add_method(vm, TrSymbol_new(vm, "match"), newMethod(vm, (TrFunc *)TrRegexp_match, TR_NIL, 1));
}