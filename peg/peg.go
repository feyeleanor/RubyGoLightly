/* Copyright (c) 2009 by Eleanor McHugh
 * Derived from C sources copyright (c) 2007 by Ian Piumarta
 * All rights reserved.
 * 
 * Permission is hereby granted, free of charge, to any person obtaining a
 * copy of this software and associated documentation files (the 'Software'),
 * to deal in the Software without restriction, including without limitation
 * the rights to use, copy, modify, merge, publish, distribute, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, provided that the above copyright notice(s) and this
 * permission notice appear in all copies of the Software.  Acknowledgement
 * of the use of this Software in supporting documentation would be
 * appreciated but is not required.
 * 
 * THE SOFTWARE IS PROVIDED 'AS IS'.  USE ENTIRELY AT YOUR OWN RISK.
 */

package pegleg

import (
	"os";
	"fmt";
	"flag";
	"container/vector";
	"bytes";
)

const (
	MAJOR = 0;
	MINOR = 1;
	LEVEL = 0;
)

help_flag := flag.Bool("h", false, "print this help information")
output_file_name := flag.String("o", "", "write output to <ofile>")
verbose_flag := flag.Bool("v", false, "verbose output")
version_flag := flag.Bool("V", false, "print version number and exit")

var input		*os.File;
var output		*os.File;
var lineNumber	int;

yybegin := 0;
yyend := 0;

var TextBuffer	[]byte;
var Token		string;
TextCursor := 0;

Thunks := vector.New(100);
ThunkCursor := 0;

func main() {
	flag.Parse();
	if *version_flag { version(); }
	if *help_flag { usage(); }
	if *output_file_name == "" {
		output = os.Stdout;
	} else {
		output, error := os.Open(*output_file_name, os.O_WRONLY, 0777);
		if output == nil { critical_file_error(output, error); }
	}

	for arg := range flag.Args {
		if arg == "-" {
			input = os.Stdin;
		} else {
			input, error := os.Open(arg, os.O_RDONLY, 0777);
			if input == nil { critical_file_error(arg, error); }
		}
		lineNumber = 1;
		TextBuffer, error := io.ReadAll(input);
		if TextBuffer == nil { critical_file_error(input.Name(), error); }
		TextCursor = 0;
		if !parse() { parser_error("syntax error"); }
		if input != os.Stdin { input.Close(); }
	}

	if *verbose_flag { for rule := range Rules.Iter() { rule.print()); } }
	Rule_compile_c_header();
	if Rules { Rule_compile_c(Rules); }
}

func critical_file_error(name string, error os.Error) {
	fmt.Fprintf(os.Stderr, "%s: error accessing '%s': %s\n", os.Args[0], name, error);
	os.Exit(1);
}

func version() {
	fmt.Printf("%s version %d.%d.%d\n", os.Args[0], MAJOR, MINOR, LEVEL);
}

func usage() {
	version();
	fmt.Fprintln(os.Stderr, "usage:", os.Args[0], "[<option>...] [<file>...]");
	fmt.Fprintln(os.Stderr, "where <option> can be");
	fmt.Fprintln(os.Stderr, "  -h          print this help information");
	fmt.Fprintln(os.Stderr, "  -o <ofile>  write output to <ofile>");
	fmt.Fprintln(os.Stderr, "  -v          be verbose");
	fmt.Fprintln(os.Stderr, "  -V          print version number and exit");
	fmt.Fprintln(os.Stderr, "if no <file> is given, input is read from stdin");
	fmt.Fprintln(os.Stderr, "if no <ofile> is given, output is written to stdout");
	os.Exit(1);
}

type Action func(*string, int);

func (self *Action) store(begin, end int) {
	for ThunkCursor > Thunks.Len() { Thunks.Push(Thunk{}) }
	Thunks[ThunkCursor] = Thunk{begin: begin, end: end, action: self};
	ThunkCursor++;
}

type Thunk struct {
	begin, end		int;
	position		int;
	action			Action;
}

func (identifier String) succeeded(position int) bool {
	fmt.Fprintln(os.Stderr, "  ok   ", identifier, "@", position);
	return true;
}

func (identifier String) failed(position int) bool {
	fmt.Fprintln(os.Stderr, "  fail ", identifier, "@", position);
	return false;
}

type Cursor struct {
	text			int;
	thunk			int;
}

func cursor_checkpoint() *Cursor {
	return &Cursor{text: TextCursor, thunk: ThunkCursor};
}

func (self *Cursor) restore() {
	ThunkCursor, TextCursor = Cursor.thunk, Cursor.text;
}

func (self *[]byte) found_token(begin, end int) {
	Token = string(self[begin : end]);
}

func parser_error(message string) {
	fmt.Fprintf(os.Stderr, "%s:%d: %s", input.Name(), lineNumber, message);
	if len(Token) > 0 { os.Fprintf(os.Stderr, " near token '%s'\n", Token); }
	if TextCursor < len(TextBuffer) {
		TextBuffer[len(TextBuffer)] = 0;
		os.Stderr.WriteString(" before text \"");
		for TextCursor < len(TextBuffer) {
			switch TextBuffer[TextCursor] {
				case '\n', '\r':
					break;

				default:
					os.Stderr.Write(TextBuffer[TextCursor]);
					TextCursor++;
			}
		}
		if TextCursor == len(TextBuffer) {
			for EOF != (c := fgetc(input)) && '\n' != c && '\r' != c {
				os.Stderr.WriteString(c);
			}
		}
		os.Stderr.WriteString('\"');
	}
	fmt.Fprintln(os.Stderr, "");
	os.Exit(1);
}

func parse() bool {
	yybegin = yyend = TextCursor;
	ThunkCursor = 0;
	if ok := parse_grammar() { done(); }
	commit();
	return ok;
}

func parse_grammar() bool {
	position := cursor_checkpoint();
	if spacing() && definition() {
		for {
			position_1 := cursor_checkpoint();
			if !definition() { break; }
		}
		position_1.restore();
		if yy_EndOfFile() { return "Grammar".succeeded(TextCursor); }
	}
	position.restore();
	return "Grammar".failed(TextCursor);
}

func done() {
	for thunk := range Thunks.Iter() {
		if thunk.end {
			TextBuffer.found_token(thunk.begin, thunk.end);
			token_length := len(Token);
		} else {
			token_length := thunk.begin;
		}
		fmt.Fprintf(os.Stderr, "DO [%d] %p %s\n", pos, thunk.action, Token);
		thunk.action(Token, token_length);
	}
	ThunkCursor = 0;
}

func commit() {
	if (len(TextBuffer) - TextCursor) > 0 { TextBuffer = TextBuffer[TextCursor : len(TextBuffer)]; }
	yybegin -= TextCursor;
	yyend -= TextCursor;
	TextCursor = ThunkCursor = 0;
}

func matchDot() bool {
	TextCursor++;
	return true;
}

func matchChar(c int) bool {
	if TextBuffer[TextCursor] == c {
		TextCursor++;
		return ("matchChar(" + c + ")").succeeded(TextCursor);
	} else {
		return ("matchChar(" + c + ")").failed(TextCursor);
	}
}

func matchString(s string) bool {
	yysav := TextCursor;
	for *s {
		if TextBuffer[TextCursor] != *s {
			TextCursor= yysav;
			return false;
		}
		s++;
		TextCursor++;
	}
	return true;
}

func matchClass(bits []byte) bool {
	c := TextBuffer[TextCursor];
	if bits[c >> 3] & (1 << (c & 7)) {
		TextCursor++;
		return "matchClass".succeeded(TextCursor);
	}
	return "matchClass".failed(TextCursor);
}

func yy_3_Suffix(text string, yyleng int) {
	push(Plus{element: pop()});
}

func yy_2_Suffix(text string, yyleng int) {
	push(Star{element: pop()});
}

func yy_1_Suffix(text string, yyleng int) {
	push(Query{element: pop()});
}

func yy_3_Prefix(text string, yyleng int) {
	push(PeekNot{element: pop()});
}

func yy_2_Prefix(text string, yyleng int) {
	push(PeekFor{element: pop()});
}

func yy_1_Prefix(text string, yyleng int) {
	push(Predicate{text: text});
}

func yy_2_Sequence(text string, yyleng int) {
	push(Predicate{text: "1"});
}

func yy_1_Sequence(text string, yyleng int) {
	f := pop();
	push(Sequence_append(pop(), f));
}

func yy_1_Expression(text string, yyleng int) {
	f := pop();
	push(Alternate_append(pop(), f));
}

func yy_2_Definition(text string, yyleng int) {
	e := pop();
	pop().setExpression(e);
}

func yy_1_Definition(text string, yyleng int) {
	if push(beginRule(findRule(text))).expression { fmt.Fprintf(os.Stderr, "rule '%s' redefined\n", text); }
}

func end_of_line() bool {
	position := cursor_checkpoint();
	if !matchString("\r\n") {
		position.restore();
		if !matchChar('\n') {
			position.restore();
			if !matchChar('\r') {
				position.restore();
				return "EndOfLine".failed(TextCursor);
			}
		}
	}
	return "EndOfLine".succeeded(TextCursor);
}

func yy_Comment() bool {
	position := cursor_checkpoint();
	if matchChar('#') {
		for {
			position_1 := cursor_checkpoint();
			if !end_of_line() {
				position_1.restore();
				if matchDot() { continue; }
			}
			break;
		}
		position_1.restore();
		if end_of_line() {
			return "Comment".succeeded(TextCursor);
		}
	}
	position.restore();
	return "Comment".failed(TextCursor);
}

func space() bool {
	position := cursor_checkpoint();
	if !matchChar(' ') {
		position.restore();
		if !matchChar('\t') {
			position.restore();
			if !end_of_line() {
				position.restore();
				return "Space".failed(TextCursor);
			}
		}
	}
	return "Space".succeeded(TextCursor);
}

func yy_Range() bool {
	position := cursor_checkpoint();
	if !yy_Char() || !matchChar('-') || !yy_Char() {
		position.restore();
		if !yy_Char() {
			position.restore();
			return "Range".failed(TextCursor);
		}
	}
	return "Range".succeeded(TextCursor);
}

func yy_Char() bool {
	position := cursor_checkpoint();
	if !matchChar('\\') || !matchClass("\000\000\000\000\204\000\000\000\000\000\000\070\146\100\124\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000") {
		position.restore();
		if !matchChar('\\') || !matchClass("\000\000\000\000\000\000\017\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000") || !matchClass("\000\000\000\000\000\000\377\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000") || !matchClass("\000\000\000\000\000\000\377\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000") {
			position.restore();
			if matchChar('\\') && matchClass("\000\000\000\000\000\000\377\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000") {
				position_1 := cursor_checkpoint();
				if !matchClass("\000\000\000\000\000\000\377\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000") {
					position_1.restore();
				}
			} else {
				position.restore();
				if !matchChar('\\') || !matchChar('-') {
					position.restore();
					position_1 := cursor_checkpoint();
					if matchChar('\\') { goto bad_char; }
					position_1.restore();
					if !matchDot() { goto bad_char; }
				}
			}
		}
	}
	return "Char".succeeded(TextCursor);

bad_char:
	position.restore();
	return "Char".failed(TextCursor);
}

func yy_IdentCont() bool {
	position := cursor_checkpoint();
	if !yy_IdentStart() {
		position.restore();
		if !matchClass("\000\000\000\000\000\000\377\003\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000") {
			position.restore();
			return "IdentCont".failed(TextCursor);
		}
	}
	return "IdentCont".succeeded(TextCursor);
}

func yy_IdentStart() bool {
	position := cursor_checkpoint();
	if matchClass("\000\000\000\000\000\000\000\000\376\377\377\207\376\377\377\007\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000") {
		return "IdentStart".succeeded(TextCursor);
	}
	position.restore();
	return "IdentStart".failed(TextCursor);
}

func yy_END() bool {
	position := cursor_checkpoint();
	if matchChar('>') && spacing() { return "END".succeeded(TextCursor); }
	position.restore();
	return "END".failed(TextCursor);
}

func yy_BEGIN() bool {
	position := cursor_checkpoint();
	if matchChar('<') && spacing() {
		return "BEGIN".succeeded(TextCursor);
	}
	position.restore();
	return "BEGIN".failed(TextCursor);
}

func yy_DOT() bool {
	position := cursor_checkpoint();
	if matchChar('.') && spacing() {
		return "DOT".succeeded(TextCursor);
	}
	position.restore();
	return "DOT".failed(TextCursor);
}

func yy_Class() bool {
	position := cursor_checkpoint();
	if matchChar('[') {
		TextBuffer.found_token(yybegin, yyend);
		yybegin = TextCursor;
		for {
			position_1 := cursor_checkpoint();
			if matchChar(']') { break; }
			position_1.restore();
			if !yy_Range() { break; }
		}
		position_1.restore();
		TextBuffer.found_token(yybegin, yyend);
		yyend = TextCursor;
		if matchChar(']') && spacing() {
			return "Class".succeeded(TextCursor);
		}
	}
	position.restore();
	return "Class".failed(TextCursor);
}

func yy_Literal() bool {
	position := cursor_checkpoint();
	if matchClass("\000\000\000\000\200\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000") {
		TextBuffer.found_token(yybegin, yyend);
		yybegin = TextCursor;
		for {
			position_1 := cursor_checkpoint();
			if matchClass("\000\000\000\000\200\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000") {
				break;
			}
			position_1.restore();
			if !yy_Char() { break; }
		}
		position_1.restore();
		TextBuffer.found_token(yybegin, yyend);
		yyend = TextCursor;
		if matchClass("\000\000\000\000\200\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000") {
			if spacing() {
				goto good_literal;
			}
		}
	}

	position.restore();
	if !matchClass("\000\000\000\000\004\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000") {
		goto bad_literal;
	}
	TextBuffer.found_token(yybegin, yyend);
	yybegin = TextCursor;
	for {
		position_1 := cursor_checkpoint();
		if !matchClass("\000\000\000\000\004\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000") {
			position_1.restore();
			if yy_Char() { continue; }
		}
		break;
	}
	position_1.restore();
	TextBuffer.found_token(yybegin, yyend);
	yyend = TextCursor;
	if !matchClass("\000\000\000\000\004\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000") {
		goto bad_literal;
	}
	if !spacing() { goto bad_literal; }

good_literal:
	return "Literal".succeeded(TextCursor);

bad_literal:
	position.restore();
	return "Literal".failed(TextCursor);
}

func yy_CLOSE() bool {
	position := cursor_checkpoint();
	if matchChar(')') && spacing() {
		return "CLOSE".succeeded(TextCursor);
	}
	position.restore();
	return "CLOSE".failed(TextCursor);
}

func yy_OPEN() bool {
	position := cursor_checkpoint();
	if matchChar('(') && spacing() {
		return "OPEN".succeeded(TextCursor);
	}
	position.restore();
	return "OPEN".failed(TextCursor);
}

func yy_PLUS() bool {
	position := cursor_checkpoint();
	if matchChar('+') && spacing() {
		return "PLUS".succeeded(TextCursor);
	}
	position.restore();
	return "PLUS".failed(TextCursor);
}

func yy_STAR() bool {
	position := cursor_checkpoint();
	if matchChar('*') && spacing() {
		return "STAR".succeeded(TextCursor);
	}
	position.restore();
	return "STAR".failed(TextCursor);
}

func yy_QUESTION() bool {
	position := cursor_checkpoint();
	if matchChar('?') && spacing() {
		return "QUESTION".succeeded(TextCursor);
	}
	position.restore();
	return "QUESTION".failed(TextCursor);
}

func yy_Primary() bool {
	position := cursor_checkpoint();
	if yy_Identifier() {
		position_1 := cursor_checkpoint();
		if !yy_LEFTARROW() {
			position_1.restore();
			func (text string, yyleng int) { push(Name{used: true, variable: nil, rule: findRule(text)}); }.store(yybegin, yyend);
			goto good_primary;
		}
	}

	position.restore();
	if !yy_OPEN() || !yy_Expression() || !yy_CLOSE() {
		position.restore();
		if yy_literal() {
			func (text string, yyleng int) { push(String{value: text}); }.store(yybegin, yyend);
		} else {
			position.restore();
			if yy_Class() {
				func (text string, yyleng int) { push(Class{cclass: text}); }.store(yybegin, yyend);
			} else {
				position.restore();
				if yy_DOT() {
					func (text string, yyleng int) { push(Dot{}); }.store(yybegin, yyend);
				} else {
					position.restore();
					if yy_Action() {
						func (text string, yyleng int) { push(makeAction(text)); }.store(yybegin, yyend);
					} else {
						position.restore();
						if yy_BEGIN() {
							func (text string, yyleng int) { push(Predicate{text: "yybegin = TextCursor"}); }.store(yybegin, yyend);
						} else {
							position.restore();
							if yy_END() {
								func(text string, yyleng int) { push(Predicate{text: "yyend = TextCursor"}); }.store(yybegin, yyend);
							} else {
								position.restore();
								return "Primary".failed(TextCursor);
							}
						}
					}
				}
			}
		}
	}

good_primary:
	return "Primary".succeeded(TextCursor);
}

func yy_NOT() bool {
	position := cursor_checkpoint();
	if matchChar('!') && spacing() {
		return "NOT".succeeded(TextCursor);
	}
	position.restore();
	return "NOT".failed(TextCursor);
}

func yy_Suffix() bool {
	position := cursor_checkpoint();
	if yy_Primary() {
		position_1 := cursor_checkpoint();
		if yy_QUESTION() {
			yy_1_Suffix.store(yybegin, yyend);
		} else {
			position_1.restore();
			if yy_STAR() {
				yy_2_Suffix.store(yybegin, yyend);
			} else {
				position_1.restore();
				if yy_PLUS() {
					yy_3_Suffix.store(yybegin, yyend);
				} else {
					position_1.restore();
				}
			}
		}
		return "Suffix".succeeded(TextCursor);
	}
	position.restore();
	return "Suffix".failed(TextCursor);
}

func yy_Action() bool {
	position := cursor_checkpoint();
	if matchChar('{') {
		TextBuffer.found_token(yybegin, yyend);
		yybegin = TextCursor;
		for {
			position_1 := Curosr{TextCursor, ThunkCursor};
			if !matchClass("\377\377\377\377\377\377\377\377\377\377\377\377\377\377\377\337\377\377\377\377\377\377\377\377\377\377\377\377\377\377\377\377") { break; }
		}
		position_1.restore();
		TextBuffer.found_token(yybegin, yyend);
		yyend = TextCursor;
		if matchChar('}') && spacing() {
			return "Action".succeeded(TextCursor);
		}
	}
	position.restore();
	return "Action".failed(TextCursor);
}

func yy_AND() bool {
	position := cursor_checkpoint();
	if matchChar('&') && spacing() {
		return "AND".succeeded(TextCursor);
	}
	position.restore();
	return "AND".failed(TextCursor);
}

func yy_Prefix() bool {
	position := cursor_checkpoint();
	if yy_AND() && yy_Action() {
		yy_1_Prefix.store(yybegin, yyend);
	} else {
		position.restore();
		if !yy_AND() || !yy_Suffix() {
			position.restore();
			if !yy_NOT() || !yy_Suffix() {
				position.restore();
				if !yy_Suffix() {
					position.restore();
					return "Prefix".failed(TextCursor);
				}
			}
			yy_3_Prefix.store(yybegin, yyend);
		} else {
			yy_2_Prefix.store(yybegin, yyend);
		}
	}
	return "Prefix".succeeded(TextCursor);
}

func yy_SLASH() bool {
	position := cursor_checkpoint();
	if matchChar('/') && spacing() {
		return "SLASH".succeeded(TextCursor);
	}
	position.restore();
	return "SLASH".failed(TextCursor);
}

func yy_Sequence() bool {
	position := cursor_checkpoint();
	if yy_Prefix() {
		for {
			position_1 := cursor_checkpoint();
			if !yy_Prefix() { break; }
			yy_1_Sequence.store(yybegin, yyend);
		}
		position_1.restore();
	} else {
		position.restore();
		yy_2_Sequence.store(yybegin, yyend);
	}
	return "Sequence".succeeded(TextCursor);

bad_sequence:
	// Apparently never gets here, which is troubling...
	position.restore();
	return "Sequence".failed(TextCursor);
}

func yy_Expression() bool {
	position := cursor_checkpoint();
	if yy_Sequence() {
		for {
			position_1 := cursor_checkpoint();
			if !yy_SLASH() || !yy_Sequence { break; }
			yy_1_Expression.store(yybegin, yyend);
		}
		position_1.restore();
		return "Expression".succeeded(TextCursor);
	}
	position.restore();
	return "Expression".failed(TextCursor);
}

func yy_LEFTARROW() bool {
	position := cursor_checkpoint();
	if matchString("<-") && spacing() {
		return "LEFTARROW".succeeded(TextCursor);
	}
	position.restore();
	return "LEFTARROW".failed(TextCursor);
}

func yy_Identifier() bool {
	position := cursor_checkpoint();
	TextBuffer.found_token(yybegin, yyend);
	yybegin = TextCursor;
	if yy_IdentStart() {
		for {
			position_1 := cursor_checkpoint();
			if !yy_IdentCont() {
				position_1.restore();
				TextBuffer.found_token(yybegin, yyend);
				yyend = TextCursor;
				if !spacing() { break; }
				return "Identifier".succeeded(TextCursor);
			}
		}
	}
	position.restore();
	return "Identifier".failed(TextCursor);
}

func yy_EndOfFile() bool {
	position := cursor_checkpoint();
	if !matchDot() {
		position.restore();
		return "EndOfFile".succeeded(TextCursor);
	}
	position.restore();
	return "EndOfFile".failed(TextCursor);
}

func definition() bool {
	position := cursor_checkpoint();
	if yy_Identifier() {
		yy_1_Definition.store(yybegin, yyend);
		if yy_LEFTARROW() && yy_Expression() {
			yy_2_Definition.store(yybegin, yyend);
			TextBuffer.found_token(yybegin, yyend);
			if ThunkCursor != 0 {
				fmt.Fprintf(os.Stderr, "accept denied at %d\n", ThunkCursor);
			} else {
				done();
				commit();
				return "Definition".succeeded(TextCursor);
			}
		}
	}
	position.restore();
	return "Definition".failed(TextCursor);
}

func spacing() bool {
	for {
		position := cursor_checkpoint();
		if !space() {
			position.restore();
			if !yy_Comment() { break; }
		}
	}
	position.restore();
	return "Spacing".succeeded(TextCursor);
}