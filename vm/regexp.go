import (
	"pcre";
	"tr";
	"internal";
	)

// Loosely based on http://vcs.pcre.org/viewvc/code/trunk/pcredemo.c

// Translate this to use Go's stdlib regexp package

func TrRegexp_new(vm *RubyVM, pattern *string, options int) RubyObject {
	r := Regexp{type: TR_T_Regexp, class: vm.classes[TR_T_Regexp], ivars: kh_init(RubyObject)};
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
		tr_raise(RegexpError, "compilation failed at offset %d: %s", erroffset, error);
	}
	return r;
}

func TrRegexp_compile(vm *RubyVM, self, pattern *RubyObject) RubyObject {
	return TrRegexp_new(vm, TR_STR_PTR(pattern), 0);
}

#define OVECCOUNT 30    /* should be a multiple of 3 */

func TrRegexp_match(vm *RubyVM, self, str *RubyObject) RubyObject {
	r := TR_CTYPE(self, Regexp);
	subject := *TR_STR_PTR(str);
	rc int;
	ovector [OVECCOUNT]int;

	rc = pcre_exec(
		r.re,                /* the compiled pattern */
		NULL,                 /* no extra data - we didn't study the pattern */
		subject,              /* the subject string */
		TR_STR_LEN(str),      /* the length of the subject */
		0,                    /* start at offset 0 in the subject */
		0,                    /* default options */
		ovector,              /* output vector for substring information */
		OVECCOUNT);           /* number of elements in the output vector */
  
	if (rc < 0) return TR_NIL;

	if (rc == 0) {
		rc = OVECCOUNT/3;
		tr_raise(RegexpError, "Too much matches, only %d supported for now", rc - 1);
	}
  
	// TODO should create a MatchData object
	data := newArray(vm);
	i int;
	for (i = 0; i < rc; i++) {
		substring_start := subject + ovector[2*i];
		substring_length := ovector[2*i+1] - ovector[2*i];
		data.kv.Push(TrString_new(vm, substring_start, substring_length));
	}
	return data;
}

func TrRegex_free(vm *RubyVM, self *RubyObject) {
	r := TrRegexp *(self);
	pcre_free(r.re);
}

func TrRegexp_init(vm *RubyVM) {
	c := vm.classes[TR_T_Regexp] = Object_const_set(vm, vm.self, tr_intern(Regexp), newClass(vm, tr_intern(Regexp), vm.classes[TR_T_Object]));
	tr_metadef(c, "new", TrRegexp_compile, 1);
	tr_def(c, "match", TrRegexp_match, 1);
}