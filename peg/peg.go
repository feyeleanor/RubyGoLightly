/* Copyright (c) 2007 by Ian Piumarta
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
 * 
 * Last edited: 2007-09-12 00:27:30 by piumarta on vps2.piumarta.com
 */

package pegleg

import (
	"os";
	"fmt";
	"flag";
	"container/vector";
)

const (
	MAJOR = 0;
	MINOR = 1;
	LEVEL = 2;
)

var help_flag = flag.Bool("h", false, "print this help information")
var output_file_name = flag.String("o", "", "write output to <ofile>")
var verbose_flag = flag.Bool("v", false, "verbose output")
var version_flag = flag.Bool("V", false, "print version number and exit")

func init() {
	input := os.Stdin;
	output := os.Stdout;
	lineNumber := 1;
	fileName := "<stdin>";
}

func main() {
	flag.Parse();
	if *version_flag { version(); }
	if *help_flag { usage(); }
	if *output_file_name != "" {
		output, error := os.Open(*output_file_name, os.O_WRONLY, 0777);
		if error {
			fmt.Fprintln(os.Stderr, *output_file_name, ":", error.String);
			os.Exit(1);
		}
	}

	for arg := range flag.Args {
		if arg == "-" {
			input = os.Stdin;
			fileName = "<stdin>";
		} else {
			input, error := os.Open(arg, os.O_RDONLY, 0777);
			if error {
				fmt.Fprintln(os.Stderr, arg, ":", error.String());
				os.Exit(1);
			}
			fileName = arg;
		}
		lineNumber = 1;
		if !yyparse() { yyerror("syntax error"); }
		if input != os.Stdin { input.Close(); }
	}

	if *verbose_flag { for rule := range Rules.Iter() { rule.print()); } }
	Rule_compile_c_header();
	if Rules { Rule_compile_c(Rules); }
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

func YY_INPUT(buf, result, max) {
	c := getc(input);
	if '\n' == c || '\r' == c { lineNumber++; }
	result := (EOF == c) ? 0 : (*(buf)= c, 1);
}

typedef void (*yyaction)(char *yytext, int yyleng);

type struct yyThunk {
	begin, end		int;
	position		int;
	action			yyaction;
}

func (self yyThunk) save(position int) *yyThunk {
	return &yyThunk{begin: self.begin, end: self.end, action: self.action, position: position};
}

func (self yyThunk) restore() (thunk yyThunk, position int) {
	return yyThunk{begin: self.begin, end: self.end, action: self.action}, self.position;
}

yybuf : = "";
yybuflen := 0;
yypos := 0;
yylimit := 0;
yytext := "";
yytextlen := 0;
yybegin := 0;
yyend := 0;
yytextmax := 0;

Thunks := vector.New(100);
ThunkCursor := 0;

yy int;
yyval *int;
yyvals *int;
yyvalslen := 0;

func yyerror(message string) {
	fmt.Fprintf(os.Stderr, "%s:%d: %s", fileName, lineNumber, message);
	if yytext[0] { os.Fprintf(os.Stderr, " near token '%s'\n", yytext); }
	if yypos < yylimit || !feof(input) {
		yybuf[yylimit]= '\0';
		os.Stderr.WriteString(" before text \"");
		for yypos < yylimit {
			if '\n' == yybuf[yypos] || '\r' == yybuf[yypos] { break; }
			0s.Stderr.WriteString(yybuf[yypos++]);
		}
		if yypos == yylimit {
			for EOF != (c := fgetc(input)) && '\n' != c && '\r' != c { os.Stderr.WriteString(c); }
		}
		os.Stderr.WriteString('\"');
	}
	fmt.Fprintln(os.Stderr, "");
	os.Exit(1);
}

func yyrefill() bool {
	if yybuflen - yypos < 512 {
		yybuflen *= 2;
		yybuf = realloc(yybuf, yybuflen);
	}
	c := getc(input);
	if '\n' == c || '\r' == c { lineNumber++; }
	yyn := (EOF == c) ? false : (*(yybuf + yypos) = c, true);
	if !yyn { return false; }
	yylimit += yyn;
	return true;
}

func yymatchDot() bool {
	if yypos >= yylimit && !yyrefill() { return false; }
	yypos++;
	return true;
}

func yymatchChar(c int) bool {
	if yypos >= yylimit && !yyrefill() { return 0; }
	if yybuf[yypos] == c {
		yypos++;
		fmt.Fprintf(os.Stderr, "  ok   yymatchChar(%c) @ %s\n", c, yybuf + yypos);
		return true;
	}
	fmt.Fprintf(os.Stderr, "  fail yymatchChar(%c) @ %s\n", c, yybuf + yypos);
	return false;
}

func yymatchString(s string) bool {
	yysav := yypos;
	for *s {
		if yypos >= yylimit && !yyrefill() { return 0; }
		if yybuf[yypos] != *s {
			yypos= yysav;
			return false;
		}
		s++;
		yypos++;
	}
	return true;
}

func yymatchClass(bits []byte) bool {
	if yypos >= yylimit && !yyrefill() { return false; }
	c := yybuf[yypos];
	if bits[c >> 3] & (1 << (c & 7)) {
		yypos++;
		fmt.Fprintln(os.Stderr, "  ok   yymatchClass @ ", yybuf + yypos);
		return true;
	}
	fmt.Fprintln(os.Stderr, "  fail yymatchClass @ ", yybuf + yypos);
	return false;
}

func yyDo(action yyaction, begin, end int) {
	Thunks[ThunkCursor] = yyThunk{begin: begin, end: end, action: action};
	ThunkCursor++;
}

func yyText(begin, end int) int {
	yyleng := end - begin;
	if yyleng <= 0 {
		yyleng= 0;
	} else {
		if yytextlen < (yyleng - 1) {
			yytextlen *= 2;
			yytext= realloc(yytext, yytextlen);
		}
		memcpy(yytext, yybuf + begin, yyleng);
	}
	yytext[yyleng]= '\0';
	return yyleng;
}

func yyDone() {
	for thunk := range Thunks.Iter() {
		yyleng := thunk.end ? yyText(thunk.begin, thunk.end) : thunk.begin;
		fmt.Fprintf(os.Stderr, "DO [%d] %p %s\n", pos, thunk.action, yytext);
		thunk.action(yytext, yyleng);
	}
	ThunkCursor = 0;
}

func yyCommit() {
	if (yylimit -= yypos) > 0 { memmove(yybuf, yybuf + yypos, yylimit); }
	yybegin -= yypos;
	yyend -= yypos;
	yypos = ThunkCursor = 0;
}

func yyAccept(tp0 int) bool {
	if tp0 {
		fmt.Fprintf(os.Stderr, "accept denied at %d\n", tp0);
		return false;
	} else {
		yyDone();
		yyCommit();
	}
	return true;
}

func yyPush(text string, count int)	{ yyval += count; }
func yyPop(text string, count int)	{ yyval -= count; }
func yySet(text string, count int)	{ yyval[count]= yy; }

func yy_7_Primary(yytext string, yyleng int) {
	fmt.Fprintln(os.Stderr, "do yy_7_Primary");
	push(Predicate{text: "yyend = yypos"});
}

func yy_6_Primary(yytext string, yyleng int) {
	fmt.Fprintln(os.Stderr, "do yy_6_Primary");
	push(Predicate{text: "yybegin = yypos"});
}

func yy_5_Primary(yytext string, yyleng int) {
	fmt.Fprintln(os.Stderr, "do yy_5_Primary");
	push(makeAction(yytext));
}

func yy_4_Primary(yytext string, yyleng int) {
	fmt.Fprintln(os.Stderr, "do yy_4_Primary");
	push(Dot{});
}

func yy_3_Primary(yytext string, yyleng int) {
	fmt.Fprintln(os.Stderr, "do yy_3_Primary");
	push(Class{cclass: yytext});
}

func yy_2_Primary(yytext string, yyleng int) {
	fmt.Fprintln(os.Stderr, "do yy_2_Primary");
	push(String{value: yytext});
}

func yy_1_Primary(yytext string, yyleng int) {
	fmt.Fprintln(os.Stderr, "do yy_1_Primary");
	push(Name{used: true, variable: nil, rule: findRule(yytext)});
}

func yy_3_Suffix(yytext string, yyleng int) {
	fmt.Fprintln(os.Stderr, "do yy_3_Suffix");
	push(Plus{element: pop()});
}

func yy_2_Suffix(yytext string, yyleng int) {
	fmt.Fprintln(os.Stderr, "do yy_2_Suffix");
	push(Star{element: pop()});
}

func yy_1_Suffix(yytext string, yyleng int) {
	fmt.Fprintln(os.Stderr, "do yy_1_Suffix");
	push(Query{element: pop()});
}

func yy_3_Prefix(yytext string, yyleng int) {
	fmt.Fprintln(os.Stderr, "do yy_3_Prefix");
	push(PeekNot{element: pop()});
}

func yy_2_Prefix(yytext string, yyleng int) {
	fmt.Fprintln(os.Stderr, "do yy_2_Prefix");
	push(PeekFor{element: pop()});
}

func yy_1_Prefix(yytext string, yyleng int) {
	fmt.Fprintln(os.Stderr, "do yy_1_Prefix");
	push(Predicate{text: yytext});
}

func yy_2_Sequence(yytext string, yyleng int) {
	fmt.Fprintln(os.Stderr, "do yy_2_Sequence");
	push(Predicate{text: "1"});
}

func yy_1_Sequence(yytext string, yyleng int) {
	fmt.Fprintln(os.Stderr, "do yy_1_Sequence");
	f := pop();
	push(Sequence_append(pop(), f));
}

func yy_1_Expression(yytext string, yyleng int) {
	fmt.Fprintln(os.Stderr, "do yy_1_Expression");
	f := pop();
	push(Alternate_append(pop(), f));
}

func yy_2_Definition(yytext string, yyleng int) {
	fmt.Fprintln(os.Stderr, "do yy_2_Definition");
	e := pop();
	Rule_setExpression(pop(), e);
}

func yy_1_Definition(yytext string, yyleng int) {
	fmt.Fprintln(os.Stderr, "do yy_1_Definition");
	if push(beginRule(findRule(yytext))).expression { fmt.Fprintf(os.Stderr, "rule '%s' redefined\n", yytext); }
}

func yy_EndOfLine() bool {
	position := ThunkCursor.save(yypos);
	fmt.Fprintln(os.Stderr, "EndOfLine");
	if !yymatchString("\r\n") {
		ThunkCursor, yypos = position.restore();
		if !yymatchChar('\n') {
			ThunkCursor, yypos = position.restore();
			if !yymatchChar('\r') {
				ThunkCursor, yypos = position.restore();
				fmt.Fprintf(os.Stderr, "  fail %s @ %s\n", "EndOfLine", yybuf + yypos);
				return false;
			}
		}
	}
	fmt.Fprintf(os.Stderr, "  ok   %s @ %s\n", "EndOfLine", yybuf + yypos);
	return true;
}

func yy_Comment() bool {
	position := ThunkCursor.save(yypos);
	fmt.Fprintln(os.Stderr, "Comment");
	if yymatchChar('#') {
		for {
			position_1 := ThunkCursor.save(yypos);
			if !yy_EndOfLine() {
				ThunkCursor, yypos = position_1.restore();
				if yymatchDot() { continue; }
			}
			break;
		}
		ThunkCursor, yypos = position_1.restore();
		if yy_EndOfLine() {
			fmt.Fprintf(os.Stderr, "  ok   %s @ %s\n", "Comment", yybuf + yypos);
			return true;
		}
	}
	ThunkCursor, yypos = position.restore();
	fmt.Fprintf(os.Stderr, "  fail %s @ %s\n", "Comment", yybuf + yypos);
	return false;
}

func yy_Space() bool {
	position := ThunkCursor.save(yypos);
	fmt.Fprintln(os.Stderr, "Space");
	if !yymatchChar(' ') {
		ThunkCursor, yypos = position.restore();
		if !yymatchChar('\t') {
			ThunkCursor, yypos = position.restore();
			if !yy_EndOfLine() {
				ThunkCursor, yypos = position.restore();
				fmt.Fprintf(os.Stderr, "  fail %s @ %s\n", "Space", yybuf + yypos);
				return false;
			}
		}
	}
	fmt.Fprintf(os.Stderr, "  ok   %s @ %s\n", "Space", yybuf + yypos);
	return true;
}

func yy_Range() bool {
	position := ThunkCursor.save(yypos);
	fmt.Fprintln(os.Stderr, "Range");
	if !yy_Char() || !yymatchChar('-') || !yy_Char() {
		ThunkCursor, yypos = position.restore();
		if !yy_Char() {
			ThunkCursor, yypos = position.restore();
			fmt.Fprintf(os.Stderr, "  fail %s @ %s\n", "Range", yybuf + yypos);
			return false;
		}
	}
	fmt.Fprintf(os.Stderr, "  ok   %s @ %s\n", "Range", yybuf + yypos);
	return true;
}

func yy_Char() bool {
	position := ThunkCursor.save(yypos);
	fmt.Fprintln(os.Stderr, "Char");
	if !yymatchChar('\\') || !yymatchClass("\000\000\000\000\204\000\000\000\000\000\000\070\146\100\124\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000") {
		ThunkCursor, yypos = position.restore();
		if !yymatchChar('\\') || !yymatchClass("\000\000\000\000\000\000\017\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000") || !yymatchClass("\000\000\000\000\000\000\377\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000") || !yymatchClass("\000\000\000\000\000\000\377\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000") {
			ThunkCursor, yypos = position.restore();
			if yymatchChar('\\') && yymatchClass("\000\000\000\000\000\000\377\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000") {
				position_1 := ThunkCursor.save(yypos);
				if !yymatchClass("\000\000\000\000\000\000\377\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000") {
					ThunkCursor, yypos = position_1.restore();
				}
			} else {
				ThunkCursor, yypos = position.restore();
				if !yymatchChar('\\') || !yymatchChar('-') {
					ThunkCursor, yypos = position.restore();
					position_1 := ThunkCursor.save(yypos);
					if yymatchChar('\\') { goto bad_char; }
					ThunkCursor, yypos = position_1.restore();
					if !yymatchDot() { goto bad_char; }
				}
			}
		}
	}
	fmt.Fprintf(os.Stderr, "  ok   %s @ %s\n", "Char", yybuf + yypos);
	return true;

bad_char:
	ThunkCursor, yypos = position.restore();
	fmt.Fprintf(os.Stderr, "  fail %s @ %s\n", "Char", yybuf + yypos);
	return false;
}

func yy_IdentCont() bool {
	position := ThunkCursor.save(yypos);
	fmt.Fprintln(os.Stderr, "IdentCont");
	if !yy_IdentStart() {
		ThunkCursor, yypos = position.restore();
		if !yymatchClass("\000\000\000\000\000\000\377\003\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000") {
			ThunkCursor, yypos = position.restore();
			fmt.Fprintf(os.Stderr, "  fail %s @ %s\n", "IdentCont", yybuf + yypos);
			return false;
		}
	}
	fmt.Fprintf(os.Stderr, "  ok   %s @ %s\n", "IdentCont", yybuf + yypos);
	return true;
}

func yy_IdentStart() bool {
	position := ThunkCursor.save(yypos);
	fmt.Fprintln(os.Stderr, "IdentStart");
	if yymatchClass("\000\000\000\000\000\000\000\000\376\377\377\207\376\377\377\007\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000") {
		fmt.Fprintf(os.Stderr, "  ok   %s @ %s\n", "IdentStart", yybuf + yypos);
		return true;
	}
	ThunkCursor, yypos = position.restore();
	fmt.Fprintf(os.Stderr, "  fail %s @ %s\n", "IdentStart", yybuf + yypos);
	return false;
}

func yy_END() bool {
	position := ThunkCursor.save(yypos);
	fmt.Fprintln(os.Stderr, "END");
	if yymatchChar('>') && yy_Spacing() {
		fmt.Fprintf(os.Stderr, "  ok   %s @ %s\n", "END", yybuf + yypos);
		return true;
	}
	ThunkCursor, yypos = position.restore();
	fmt.Fprintf(os.Stderr, "  fail %s @ %s\n", "END", yybuf + yypos);
	return false;
}

func yy_BEGIN() bool {
	position := ThunkCursor.save(yypos);
	fmt.Fprintln(os.Stderr, "BEGIN");
	if yymatchChar('<') && yy_Spacing() {
		fmt.Fprintf(os.Stderr, "  ok   %s @ %s\n", "BEGIN", yybuf + yypos);
		return true;
	}
	ThunkCursor, yypos = position.restore();
	fmt.Fprintf(os.Stderr, "  fail %s @ %s\n", "BEGIN", yybuf + yypos);
	return false;
}

func yy_DOT() bool {
	position := ThunkCursor.save(yypos);
	fmt.Fprintln(os.Stderr, "DOT");
	if yymatchChar('.') && yy_Spacing() {
		fmt.Fprintf(os.Stderr, "  ok   %s @ %s\n", "DOT", yybuf + yypos);
		return true;
	}
	ThunkCursor, yypos = position.restore();
	fmt.Fprintf(os.Stderr, "  fail %s @ %s\n", "DOT", yybuf + yypos);
	return false;
}

func yy_Class() bool {
	position := ThunkCursor.save(yypos);
	fmt.Fprintln(os.Stderr, "Class");
	if yymatchChar('[') {
		yyText(yybegin, yyend);
		yybegin = yypos;
		for {
			position_1 := ThunkCursor.save(yypos);
			if yymatchChar(']') { break; }
			ThunkCursor, yypos = position_1.restore();
			if !yy_Range() { break; }
		}
		ThunkCursor, yypos = position_1.restore();
		yyText(yybegin, yyend);
		yyend = yypos;
		if yymatchChar(']') && yy_Spacing() {
			fmt.Fprintf(os.Stderr, "  ok   %s @ %s\n", "Class", yybuf + yypos);
			return true;
		}
	}
	ThunkCursor, yypos = position.restore();
	fmt.Fprintf(os.Stderr, "  fail %s @ %s\n", "Class", yybuf + yypos);
	return false;
}

func yy_Literal() bool {
	position := ThunkCursor.save(yypos);
	fmt.Fprintln(os.Stderr, "Literal");
	if yymatchClass("\000\000\000\000\200\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000") {
		yyText(yybegin, yyend);
		yybegin = yypos;
		for {
			position_1 := ThunkCursor.save(yypos);
			if yymatchClass("\000\000\000\000\200\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000") {
				break;
			}
			ThunkCursor, yypos = position_1.restore();
			if !yy_Char() { break; }
		}
		ThunkCursor, yypos = position_1.restore();
		yyText(yybegin, yyend);
		yyend = yypos;
		if yymatchClass("\000\000\000\000\200\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000") {
			if yy_Spacing() {
				goto good_literal;
			}
		}
	}

	ThunkCursor, yypos = position.restore();
	if !yymatchClass("\000\000\000\000\004\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000") {
		goto bad_literal;
	}
	yyText(yybegin, yyend);
	yybegin = yypos;
	for {
		position_1 := ThunkCursor.save(yypos);
		if !yymatchClass("\000\000\000\000\004\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000") {
			ThunkCursor, yypos = position_1.restore();
			if yy_Char() { continue; }
		}
		break;
	}
	ThunkCursor, yypos = position_1.restore();
	yyText(yybegin, yyend);
	yyend = yypos;
	if !yymatchClass("\000\000\000\000\004\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000") {
		goto bad_literal;
	}
	if !yy_Spacing() { goto bad_literal; }

good_literal:
	fmt.Fprintf(os.Stderr, "  ok   %s @ %s\n", "Literal", yybuf + yypos);
	return true;

bad_literal:
	ThunkCursor, yypos = position.restore();
	fmt.Fprintf(os.Stderr, "  fail %s @ %s\n", "Literal", yybuf + yypos);
	return false;
}

func yy_CLOSE() bool {
	position := ThunkCursor.save(yypos);
	fmt.Fprintln(os.Stderr, "CLOSE");
	if yymatchChar(')') && yy_Spacing() {
		fmt.Fprintf(os.Stderr, "  ok   %s @ %s\n", "CLOSE", yybuf + yypos);
		return true;
	}
	ThunkCursor, yypos = position.restore();
	fmt.Fprintf(os.Stderr, "  fail %s @ %s\n", "CLOSE", yybuf + yypos);
	return false;
}

func yy_OPEN() bool {
	position := ThunkCursor.save(yypos);
	fmt.Fprintln(os.Stderr, "OPEN");
	if yymatchChar('(') && yy_Spacing() {
		fmt.Fprintf(os.Stderr, "  ok   %s @ %s\n", "OPEN", yybuf + yypos);
		return true;
	}
	ThunkCursor, yypos = position.restore();
	fmt.Fprintf(os.Stderr, "  fail %s @ %s\n", "OPEN", yybuf + yypos);
	return false;
}

func yy_PLUS() bool {
	position := ThunkCursor.save(yypos);
	fmt.Fprintln(os.Stderr, "PLUS");
	if yymatchChar('+') && yy_Spacing() {
		fmt.Fprintf(os.Stderr, "  ok   %s @ %s\n", "PLUS", yybuf + yypos);
		return true;
	}
	ThunkCursor, yypos = position.restore();
	fmt.Fprintf(os.Stderr, "  fail %s @ %s\n", "PLUS", yybuf + yypos);
	return false;
}

func yy_STAR() bool {
	position := ThunkCursor.save(yypos);
	fmt.Fprintln(os.Stderr, "STAR");
	if yymatchChar('*') && yy_Spacing() {
		fmt.Fprintf(os.Stderr, "  ok   %s @ %s\n", "STAR", yybuf + yypos);
		return true;
	}
	ThunkCursor, yypos = position.restore();
	fmt.Fprintf(os.Stderr, "  fail %s @ %s\n", "STAR", yybuf + yypos);
	return false;
}

func yy_QUESTION() bool {
	position := ThunkCursor.save(yypos);
	fmt.Fprintln(os.Stderr, "QUESTION");
	if yymatchChar('?') && yy_Spacing() {
		fmt.Fprintf(os.Stderr, "  ok   %s @ %s\n", "QUESTION", yybuf + yypos);
		return true;
	}
	ThunkCursor, yypos = position.restore();
	fmt.Fprintf(os.Stderr, "  fail %s @ %s\n", "QUESTION", yybuf + yypos);
	return false;
}

func yy_Primary() bool {
	position := ThunkCursor.save(yypos);
	fmt.Fprintln(os.Stderr, "Primary");
	if yy_Identifier() {
		position_1 := ThunkCursor.save(yypos);
		if !yy_LEFTARROW() {
			ThunkCursor, yypos = position_1.restore();
			yyDo(yy_1_Primary, yybegin, yyend);
			goto good_primary;
		}
	}

	ThunkCursor, yypos = position.restore();
	if !yy_OPEN() || !yy_Expression() || !yy_CLOSE() {
		ThunkCursor, yypos = position.restore();
		if yy_literal() {
			yyDo(yy_2_Primary, yybegin, yyend);
		} else {
			ThunkCursor, yypos = position.restore();
			if yy_Class() {
				yyDo(yy_3_Primary, yybegin, yyend);
			} else {
				ThunkCursor, yypos = position.restore();
				if yy_DOT() {
					yyDo(yy_4_Primary, yybegin, yyend);
				} else {
					ThunkCursor, yypos = position.restore();
					if yy_Action() {
						yyDo(yy_5_Primary, yybegin, yyend);
					} else {
						ThunkCursor, yypos = position.restore();
						if yy_BEGIN() {
							yyDo(yy_6_Primary, yybegin, yyend);
						} else {
							ThunkCursor, yypos = position.restore();
							if yy_END() {
								yyDo(yy_7_Primary, yybegin, yyend);
							} else {
								ThunkCursor, yypos = position.restore();
								fmt.Fprintf(os.Stderr, "  fail %s @ %s\n", "Primary", yybuf + yypos);
								return false;
							}
						}
					}
				}
			}
		}
	}

good_primary:
	fmt.Fprintf(os.Stderr, "  ok   %s @ %s\n", "Primary", yybuf + yypos);
	return true;
}

func yy_NOT() bool {
	position := ThunkCursor.save(yypos);
	fmt.Fprintln(os.Stderr, "NOT");
	if yymatchChar('!') && yy_Spacing() {
		fmt.Fprintf(os.Stderr, "  ok   %s @ %s\n", "NOT", yybuf + yypos);
		return true;
	}
	ThunkCursor, yypos = position.restore();
	fmt.Fprintf(os.Stderr, "  fail %s @ %s\n", "NOT", yybuf + yypos);
	return false;
}

func yy_Suffix() bool {
	position := ThunkCursor.save(yypos);
	fmt.Fprintln(os.Stderr, "Suffix");
	if yy_Primary() {
		position_1 := ThunkCursor.save(yypos);
		if yy_QUESTION() {
			yyDo(yy_1_Suffix, yybegin, yyend);
		} else {
			ThunkCursor, yypos = position_1.restore();
			if yy_STAR() {
				yyDo(yy_2_Suffix, yybegin, yyend);
			} else {
				ThunkCursor, yypos = position_1.restore();
				if yy_PLUS() {
					yyDo(yy_3_Suffix, yybegin, yyend);
				} else {
					ThunkCursor, yypos = position_1.restore();
				}
			}
		}
		fmt.Fprintf(os.Stderr, "  ok   %s @ %s\n", "Suffix", yybuf + yypos);
		return true;
	}
	ThunkCursor, yypos = position.restore();
	fmt.Fprintf(os.Stderr, "  fail %s @ %s\n", "Suffix", yybuf + yypos);
	return false;
}

func yy_Action() bool {
	position := ThunkCursor.save(yypos);
	fmt.Fprintln(os.Stderr, "Action");
	if yymatchChar('{') {
		yyText(yybegin, yyend);
		yybegin = yypos;
		for {
			position_1 := ThunkCursor.save(yypos);
			if !yymatchClass("\377\377\377\377\377\377\377\377\377\377\377\377\377\377\377\337\377\377\377\377\377\377\377\377\377\377\377\377\377\377\377\377") { break; }
		}
		ThunkCursor, yypos = position_1.restore();
		yyText(yybegin, yyend);
		yyend = yypos;
		if yymatchChar('}') && yy_Spacing() {
			fmt.Fprintf(os.Stderr, "  ok   %s @ %s\n", "Action", yybuf + yypos);
			return true;
		}
	}
	ThunkCursor, yypos = position.restore();
	fmt.Fprintf(os.Stderr, "  fail %s @ %s\n", "Action", yybuf + yypos);
	return false;
}

func yy_AND() bool {
	position := ThunkCursor.save(yypos);
	fmt.Fprintln(os.Stderr, "AND");
	if yymatchChar('&') && yy_Spacing() {
		fmt.Fprintf(os.Stderr, "  ok   %s @ %s\n", "AND", yybuf + yypos);
		return true;
	}
	ThunkCursor, yypos = position.restore();
	fmt.Fprintf(os.Stderr, "  fail %s @ %s\n", "AND", yybuf + yypos);
	return false;
}

func yy_Prefix() bool {
	position := ThunkCursor.save(yypos);
	fmt.Fprintln(os.Stderr, "Prefix");
	if yy_AND() && yy_Action() {
		yyDo(yy_1_Prefix, yybegin, yyend);
	} else {
		ThunkCursor, yypos = position.restore();
		if !yy_AND() || !yy_Suffix() {
			ThunkCursor, yypos = position.restore();
			if !yy_NOT() || !yy_Suffix() {
				ThunkCursor, yypos = position.restore();
				if !yy_Suffix() {
					ThunkCursor, yypos = position.restore();
					fmt.Fprintf(os.Stderr, "  fail %s @ %s\n", "Prefix", yybuf + yypos);
					return false;
				}
			}
			yyDo(yy_3_Prefix, yybegin, yyend);
		} else {
			yyDo(yy_2_Prefix, yybegin, yyend);
		}
	}
	fmt.Fprintf(os.Stderr, "  ok   %s @ %s\n", "Prefix", yybuf + yypos);
	return true;
}

func yy_SLASH() bool {
	position := ThunkCursor.save(yypos);
	fmt.Fprintln(os.Stderr, "SLASH");
	if yymatchChar('/') && yy_Spacing() {
		fmt.Fprintf(os.Stderr, "  ok   %s @ %s\n", "SLASH", yybuf + yypos);
		return true;
	}
	ThunkCursor, yypos = position.restore();
	fmt.Fprintf(os.Stderr, "  fail %s @ %s\n", "SLASH", yybuf + yypos);
	return false;
}

func yy_Sequence() bool {
	position := ThunkCursor.save(yypos);
	fmt.Fprintln(os.Stderr, "Sequence");
	if yy_Prefix() {
		for {
			position_1 := ThunkCursor.save(yypos);
			if !yy_Prefix() { break; }
			yyDo(yy_1_Sequence, yybegin, yyend);
		}
		ThunkCursor, yypos = position_1.restore();
	} else {
		ThunkCursor, yypos = position.restore();
		yyDo(yy_2_Sequence, yybegin, yyend);
	}
	fmt.Fprintf(os.Stderr, "  ok   %s @ %s\n", "Sequence", yybuf + yypos);
	return true;

bad_sequence:
	// Apparently never gets here, which is troubling...
	ThunkCursor, yypos = position.restore();
	fmt.Fprintf(os.Stderr, "  fail %s @ %s\n", "Sequence", yybuf + yypos);
	return false;
}

func yy_Expression() bool {
	position := ThunkCursor.save(yypos);
	fmt.Fprintln(os.Stderr, "Expression");
	if yy_Sequence() {
		for {
			position_1 := ThunkCursor.save(yypos);
			if !yy_SLASH() || !yy_Sequence { break; }
			yyDo(yy_1_Expression, yybegin, yyend);
		}
		ThunkCursor, yypos = position_1.restore();
		fmt.Fprintf(os.Stderr, "  ok   %s @ %s\n", "Expression", yybuf + yypos);
		return true;
	}
	ThunkCursor, yypos = position.restore();
	fmt.Fprintf(os.Stderr, "  fail %s @ %s\n", "Expression", yybuf + yypos);
	return false;
}

func yy_LEFTARROW() bool {
	position := ThunkCursor.save(yypos);
	fmt.Fprintln(os.Stderr, "LEFTARROW");
	if yymatchString("<-") && yy_Spacing() {
		fmt.Fprintf(os.Stderr, "  ok   %s @ %s\n", "LEFTARROW", yybuf + yypos);
		return true;
	}
	ThunkCursor, yypos = position.restore();
	fmt.Fprintf(os.Stderr, "  fail %s @ %s\n", "LEFTARROW", yybuf + yypos);
	return false;
}

func yy_Identifier() bool {
	position := ThunkCursor.save(yypos);
	fmt.Fprintln(os.Stderr, "Identifier");
	yyText(yybegin, yyend);
	yybegin = yypos;
	if yy_IdentStart() {
		for {
			position_1 := ThunkCursor.save(yypos);
			if !yy_IdentCont() {
				ThunkCursor, yypos = position_1.restore();
				yyText(yybegin, yyend);
				yyend = yypos;
				if !yy_Spacing() { break; }
				fmt.Fprintf(os.Stderr, "  ok   %s @ %s\n", "Identifier", yybuf + yypos);
				return true;
			}
		}
	}
	ThunkCursor, yypos = position.restore();
	fmt.Fprintf(os.Stderr, "  fail %s @ %s\n", "Identifier", yybuf + yypos);
	return false;
}

func yy_EndOfFile() bool {
	position := ThunkCursor.save(yypos);
	fmt.Fprintln(os.Stderr, "EndOfFile");
	if !yymatchDot() {
		ThunkCursor, yypos = position.restore();
		fmt.Fprintf(os.Stderr, "  ok   %s @ %s\n", "EndOfFile", yybuf + yypos);
		return true;
	}
	ThunkCursor, yypos = position.restore();
	fmt.Fprintf(os.Stderr, "  fail %s @ %s\n", "EndOfFile", yybuf + yypos);
	return false;
}

func yy_Definition() bool {
	position := ThunkCursor.save(yypos);
	fmt.Fprintln(os.Stderr, "Definition");
	if yy_Identifier() {
		yyDo(yy_1_Definition, yybegin, yyend);
		if yy_LEFTARROW() && yy_Expression() {
			yyDo(yy_2_Definition, yybegin, yyend);
			yyText(yybegin, yyend);
			if yyAccept(yythunkpos0) {
				fmt.Fprintf(os.Stderr, "  ok   %s @ %s\n", "Definition", yybuf + yypos);
				return true;
			}
		}
	}
	ThunkCursor, yypos = position.restore();
	fmt.Fprintf(os.Stderr, "  fail %s @ %s\n", "Definition", yybuf + yypos);
	return false;
}

func yy_Spacing() bool {
	fmt.Fprintln(os.Stderr, "Spacing");
	for {
		position := ThunkCursor.save(yypos);
		if !yy_Space() {
			ThunkCursor = position.restore();
			if !yy_Comment() { break; }
		}
	}
	ThunkCursor, yypos = position.restore();
	fmt.Fprintf(os.Stderr, "  ok   %s @ %s\n", "Spacing", yybuf + yypos);
	return true;
}

func yy_Grammar() bool {
	position := ThunkCursor.save(yypos);
	fmt.Fprintln(os.Stderr, "Grammar");
	if yy_Spacing() && yy_Definition {
		for {
			position_1 := ThunkCursor.save(yypos);
			if !yy_Definition() { break; }
		}

		ThunkCursor, yypos := position_1.restore();
		if yy_EndOfFile() {
			fmt.Fprintf(os.Stderr, "  ok   %s @ %s\n", "Grammar", yybuf + yypos);
			return true;
		}
	}
	ThunkCursor := position.restore();
	fmt.Fprintf(os.Stderr, "  fail %s @ %s\n", "Grammar", yybuf + yypos);
	return false;
}