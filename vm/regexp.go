#include <pcre.h>
#include "tr.h"
#include "internal.h"

// Loosely based on http://vcs.pcre.org/viewvc/code/trunk/pcredemo.c

// Translate this to use Go's stdlib regexp package

func TrRegexp_new(vm *RubyVM, pattern *string, int options) OBJ {
	r := *TR_INIT_CORE_OBJECT(Regexp);
	error *string;
	erroffset int;
  
	r.re = pcre_compile(
		pattern,              /* the pattern */
		options,              /* default options */
		&error,               /* for error message */
		&erroffset,           /* for error offset */
		nil);                /* use default character tables */
  
	if (r.re == nil) {
		TrRegex_free(vm, OBJ(r));
		tr_raise(RegexpError, "compilation failed at offset %d: %s", erroffset, error);
	}
	return OBJ(r);
}

OBJ TrRegexp_compile(vm *RubyVM, OBJ self, OBJ pattern) {
	return TrRegexp_new(vm, TR_STR_PTR(pattern), 0);
}

#define OVECCOUNT 30    /* should be a multiple of 3 */

func TrRegexp_match(vm *RubyVM, OBJ self, OBJ str) OBJ {
	r := *TR_CREGEXP(self);
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

func TrRegex_free(vm *RubyVM, OBJ self) {
	TrRegexp *r = (TrRegexp*)self;
	pcre_free(r.re);
	TR_FREE(r);
}

func TrRegexp_init(vm *RubyVM) {
	c := TR_INIT_CORE_CLASS(Regexp, Object);
	tr_metadef(c, "new", TrRegexp_compile, 1);
	tr_def(c, "match", TrRegexp_match, 1);
}