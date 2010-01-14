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
 * Last edited: 2007-08-31 13:55:23 by piumarta on emilia.local
 */

package pegleg

prev := 0
func yyl() {
  prev++;
  return prev;
}

func charClassSet(bits []byte, c int) int {
	bits[c >> 3] |=  (1 << (c & 7));
}

func charClassClear(bits []byte, c int)	int {
	bits[c >> 3] &= ~(1 << (c & 7));
}

typedef void (*setter)(unsigned char bits[], int c);

func *makeCharClass(unsigned char *cclass) *byte {
	bits	[32]byte;
	set		setter;
	c		int;
	prev := -1;

  static char	 string[256];
  char		*ptr;

	if ('^' == *cclass) {
		memset(bits, 255, 32);
		set := charClassClear;
		cclass++;
	} else {
		memset(bits, 0, 32);
		set := charClassSet;
	}

	for (c = *cclass++) {
		if ('-' == c && *cclass && prev >= 0) {
			for  c = *cclass++;  prev <= c;  ++prev {
				set(bits, prev);
			}
			prev= -1;
		} else if '\\' == c && *cclass {
			switch c = *cclass++ {
				case 'a':  c= '\a';	/* bel */
				case 'b':  c= '\b';	/* bs */
				case 'e':  c= '\e';	/* esc */
				case 'f':  c= '\f';	/* ff */
				case 'n':  c= '\n';	/* nl */
				case 'r':  c= '\r';	/* cr */
				case 't':  c= '\t';	/* ht */
				case 'v':  c= '\v';	/* vt */
			}
			set(bits, prev= c);
		} else {
			set(bits, prev= c);
		}
	}

	ptr = string;
	for c := 0;  c < 32;  c++ {
		ptr += sprintf(ptr, "\\%03o", bits[c]);
	}
	return string;
}

func begin() { fprintf(output, "\n  {"); }
func end() { fprintf(output, "\n  }"); }
func label(n int) { fprintf(output, "\n  l%d:;\t", n); }
func jump(n int) { fprintf(output, "  goto l%d;", n); }
func save(n int) { fprintf(output, "  int yypos%d= yypos, yythunkpos%d= yythunkpos;", n, n); }
func restore(n int) { fprintf(output,     "  yypos= yypos%d; yythunkpos= yythunkpos%d;", n, n); }

func Node_compile_c_ko(node *Node, ko int) {
	assert(node);
	switch (node.type) {
		case Rule:
			fprintf(stderr, "\ninternal error #1 (%s)\n", node.name);
			exit(1);

		case Dot:
			fprintf(output, "  if (!matchDot()) goto l%d;", ko);

		case Name:
			fprintf(output, "  if (!yy_%s()) goto l%d;", node.rule.name, ko);
			if node.variable { fprintf(output, "  Do(yySet, %d, 0);", node.variable.offset); }

		case Character, String:
			length := len(node.value);
			if 1 == length || (2 == length && '\\' == node.value[0]) {
				fprintf(output, "  if (!matchChar('%s')) goto l%d;", node.value, ko);
			} else {
				fprintf(output, "  if (!matchString(\"%s\")) goto l%d;", node.value, ko);
			}

		case Class:
			fprintf(output, "  if (!matchClass((unsigned char *)\"%s\")) goto l%d;", makeCharClass(node.value), ko);

		case Action:
			fprintf(output, "  Do(yy%s, yybegin, yyend);", node.name);

		case Predicate:
			fprintf(output, "  yyText(yybegin, yyend);  if (!(%s)) goto l%d;", node.text, ko);

		case Alternate:
			ok := yyl();
			begin();
			save(ok);
			for node := node->alternate.first;  node;  node = node->alternate.next {
				if node->alternate.next {
					next := yyl();
					Node_compile_c_ko(node, next);
					jump(ok);
					label(next);
					restore(ok);
				} else {
					Node_compile_c_ko(node, ko);
				}
			}
			end();
			label(ok);

		case Sequence:
			for node := node->sequence.first;  node;  node = node->sequence.next {
				Node_compile_c_ko(node, ko);
			}

		case PeekFor:
			ok := yyl();
			begin();
			save(ok);
			Node_compile_c_ko(node->peekFor.element, ko);
			restore(ok);
			end();

		case PeekNot:
			ok := yyl();
			begin();
			save(ok);
			Node_compile_c_ko(node->peekFor.element, ok);
			jump(ko);
			label(ok);
			restore(ok);
			end();

		case Query:
			qko := yyl();
			qok := yyl();
			begin();
			save(qko);
			Node_compile_c_ko(node->query.element, qko);
			jump(qok);
			label(qko);
			restore(qko);
			end();
			label(qok);

		case Star:
			again := yyl();
			out := yyl();
			label(again);
			begin();
			save(out);
			Node_compile_c_ko(node->star.element, out);
			jump(again);
			label(out);
			restore(out);
			end();

		case Plus:
			again := yyl();
			out:= yyl();
			Node_compile_c_ko(node->plus.element, ko);
			label(again);
			begin();
			save(out);
			Node_compile_c_ko(node->plus.element, out);
			jump(again);
			label(out);
			restore(out);
			end();

		default:
			fprintf(stderr, "\nNode_compile_c_ko: illegal node type %d\n", node->type);
			exit(1);
	}
}


func countVariables(Node *node) int {
	count := 0;
	for node {
		count++;
		node = node.next;
	}
	return count;
}

func defineVariables(Node *node) {
	count := 0;
	for node {
		count--;
		fprintf(output, "#define %s yyval[%d]\n", node.name, count);
		node.offset = count;
		node = node.next;
	}
}

func undefineVariables(Node *node) {
	for node {
		fprintf(output, "#undef %s\n", node.name);
		node = node.next;
	}
}

func Rule_compile_c2(Node *node) {
	assert(node);
	assert(Rule == node.type);

	if !node.expression {
		fprintf(stderr, "rule '%s' used but not defined\n", node.name);
	} else {
		ko := yyl()
		if !node.used && node != start { fprintf(stderr, "rule '%s' defined but not used\n", node.name) }
		safe := (Query == node.expression.type) || (Star == node.expression.type);
		fprintf(output, "\nYY_RULE(int) yy_%s()\n{", node.name);
		if !safe { save(0) }
		if node.variables { fprintf(output, "  Do(yyPush, %d, 0);", countVariables(node.variables)) }
		fprintf(output, "\n  yyprintf((stderr, \"%%s\\n\", \"%s\"));", node.name);
		Node_compile_c_ko(node.expression, ko);
		fprintf(output, "\n  yyprintf((stderr, \"  ok   %%s @ %%s\\n\", \"%s\", yybuf+yypos));", node.name);
		if node.variables { fprintf(output, "  Do(yyPop, %d, 0);", countVariables(node.variables)) }
		fprintf(output, "\n  return 1;");
		if !safe {
			label(ko);
			restore(0);
			fprintf(output, "\n  yyprintf((stderr, \"  fail %%s @ %%s\\n\", \"%s\", yybuf+yypos));", node.name);
			fprintf(output, "\n  return 0;");
		}
		fprintf(output, "\n}");
	}
	if node.next { Rule_compile_c2(node.next) }
}

header := "import ( "fmt" );"
preamble := "\
#ifndef YY_VARIABLE\n\
#define YY_VARIABLE(T)	static T\n\
#endif\n\
#ifndef YY_LOCAL\n\
#define YY_LOCAL(T)	static T\n\
#endif\n\
#ifndef YY_ACTION\n\
#define YY_ACTION(T)	static T\n\
#endif\n\
#ifndef YY_RULE\n\
#define YY_RULE(T)	static T\n\
#endif\n\
#ifndef YY_PARSE\n\
#define YY_PARSE(T)	T\n\
#endif\n\
#ifndef YYPARSE\n\
#define YYPARSE		yyparse\n\
#endif\n\
#ifndef YYPARSEFROM\n\
#define YYPARSEFROM	yyparsefrom\n\
#endif\n\
#ifndef YY_INPUT\n\
#define YY_INPUT(buf, result, max_size)			\\\n\
  {							\\\n\
    int yyc= getchar();					\\\n\
    result= (EOF == yyc) ? 0 : (*(buf)= yyc, 1);	\\\n\
    yyprintf((stderr, \"<%c>\", yyc));			\\\n\
  }\n\
#endif\n\
#ifndef YY_BEGIN\n\
#define YY_BEGIN	( yybegin= yypos, 1)\n\
#endif\n\
#ifndef YY_END\n\
#define YY_END		( yyend= yypos, 1)\n\
#endif\n\
#ifdef YY_DEBUG\n\
# define yyprintf(args)	fprintf args\n\
#else\n\
# define yyprintf(args)\n\
#endif\n\
#ifndef YYSTYPE\n\
#define YYSTYPE int\n\
#endif\n\
#ifndef YYMALLOC\n\
#define YYMALLOC malloc\n\
#endif\n\
#ifndef YYREALLOC\n\
#define YYREALLOC realloc\n\
#endif\n\
\n\
#ifndef YY_PART\n\
\n\
typedef void (*yyaction)(char *yytext, int yyleng);\n\
typedef struct _yythunk { int begin, end;  yyaction  action;  struct _yythunk *next; } yythunk;\n\
\n\
YY_VARIABLE(char *   ) yybuf= 0;\n\
YY_VARIABLE(int	     ) yybuflen= 0;\n\
YY_VARIABLE(int	     ) yypos= 0;\n\
YY_VARIABLE(int	     ) yylimit= 0;\n\
YY_VARIABLE(char *   ) yytext= 0;\n\
YY_VARIABLE(int	     ) yytextlen= 0;\n\
YY_VARIABLE(int	     ) yybegin= 0;\n\
YY_VARIABLE(int	     ) yyend= 0;\n\
YY_VARIABLE(int	     ) yytextmax= 0;\n\
YY_VARIABLE(yythunk *) yythunks= 0;\n\
YY_VARIABLE(int	     ) yythunkslen= 0;\n\
YY_VARIABLE(int      ) yythunkpos= 0;\n\
YY_VARIABLE(YYSTYPE  ) yy;\n\
YY_VARIABLE(YYSTYPE *) yyval= 0;\n\
YY_VARIABLE(YYSTYPE *) yyvals= 0;\n\
YY_VARIABLE(int      ) yyvalslen= 0;\n\
\n\
YY_LOCAL(int) yyrefill(void)\n\
{\n\
  int yyn;\n\
  while (yybuflen - yypos < 512)\n\
    {\n\
      yybuflen *= 2;\n\
      yybuf= YYREALLOC(yybuf, yybuflen);\n\
    }\n\
  int c= getc(input);\n\
  if ('\n' == c || '\r' == c) ++lineNumber;\n\
  yyn= (EOF == c) ? 0 : (*(yybuf + yypos)= c, 1);\n\
  if (!yyn) return 0;\n\
  yylimit += yyn;\n\
  return 1;\n\
}\n\
\n\
YY_LOCAL(int) matchDot(void)\n\
{\n\
  if (yypos >= yylimit && !yyrefill()) return 0;\n\
  ++yypos;\n\
  return 1;\n\
}\n\
\n\
YY_LOCAL(int) matchChar(int c)\n\
{\n\
  if (yypos >= yylimit && !yyrefill()) return 0;\n\
  if (yybuf[yypos] == c)\n\
    {\n\
      ++yypos;\n\
      yyprintf((stderr, \"  ok   matchChar(%c) @ %s\\n\", c, yybuf+yypos));\n\
      return 1;\n\
    }\n\
  yyprintf((stderr, \"  fail matchChar(%c) @ %s\\n\", c, yybuf+yypos));\n\
  return 0;\n\
}\n\
\n\
YY_LOCAL(int) matchString(char *s)\n\
{\n\
  int yysav= yypos;\n\
  while (*s)\n\
    {\n\
      if (yypos >= yylimit && !yyrefill()) return 0;\n\
      if (yybuf[yypos] != *s)\n\
        {\n\
          yypos= yysav;\n\
          return 0;\n\
        }\n\
      ++s;\n\
      ++yypos;\n\
    }\n\
  return 1;\n\
}\n\
\n\
YY_LOCAL(int) matchClass(unsigned char *bits)\n\
{\n\
  int c;\n\
  if (yypos >= yylimit && !yyrefill()) return 0;\n\
  c= yybuf[yypos];\n\
  if (bits[c >> 3] & (1 << (c & 7)))\n\
    {\n\
      ++yypos;\n\
      yyprintf((stderr, \"  ok   matchClass @ %s\\n\", yybuf+yypos));\n\
      return 1;\n\
    }\n\
  yyprintf((stderr, \"  fail matchClass @ %s\\n\", yybuf+yypos));\n\
  return 0;\n\
}\n\
\n\
YY_LOCAL(void) Do(yyaction action, int begin, int end)\n\
{\n\
  while (yythunkpos >= yythunkslen)\n\
    {\n\
      yythunkslen *= 2;\n\
      yythunks= YYREALLOC(yythunks, sizeof(yythunk) * yythunkslen);\n\
    }\n\
  yythunks[yythunkpos].begin=  begin;\n\
  yythunks[yythunkpos].end=    end;\n\
  yythunks[yythunkpos].action= action;\n\
  ++yythunkpos;\n\
}\n\
\n\
YY_LOCAL(int) yyText(int begin, int end)\n\
{\n\
  int yyleng= end - begin;\n\
  if (yyleng <= 0)\n\
    yyleng= 0;\n\
  else\n\
    {\n\
      while (yytextlen < (yyleng - 1))\n\
	{\n\
	  yytextlen *= 2;\n\
	  yytext= YYREALLOC(yytext, yytextlen);\n\
	}\n\
      memcpy(yytext, yybuf + begin, yyleng);\n\
    }\n\
  yytext[yyleng]= '\\0';\n\
  return yyleng;\n\
}\n\
\n\
YY_LOCAL(void) yyDone(void)\n\
{\n\
  int pos;\n\
  for (pos= 0;  pos < yythunkpos;  ++pos)\n\
    {\n\
      yythunk *thunk= &yythunks[pos];\n\
      int yyleng= thunk->end ? yyText(thunk->begin, thunk->end) : thunk->begin;\n\
      yyprintf((stderr, \"DO [%d] %p %s\\n\", pos, thunk->action, yytext));\n\
      thunk->action(yytext, yyleng);\n\
    }\n\
  yythunkpos= 0;\n\
}\n\
\n\
YY_LOCAL(void) yyCommit()\n\
{\n\
  if ((yylimit -= yypos))\n\
    {\n\
      memmove(yybuf, yybuf + yypos, yylimit);\n\
    }\n\
  yybegin -= yypos;\n\
  yyend -= yypos;\n\
  yypos= yythunkpos= 0;\n\
}\n\
\n\
YY_LOCAL(int) yyAccept(int tp0)\n\
{\n\
  if (tp0)\n\
    {\n\
      fprintf(stderr, \"accept denied at %d\\n\", tp0);\n\
      return 0;\n\
    }\n\
  else\n\
    {\n\
      yyDone();\n\
      yyCommit();\n\
    }\n\
  return 1;\n\
}\n\
\n\
YY_LOCAL(void) yyPush(char *text, int count)	{ (void)text; yyval += count; }\n\
YY_LOCAL(void) yyPop(char *text, int count)	{ (void)text; yyval -= count; }\n\
YY_LOCAL(void) yySet(char *text, int count)	{ (void)text; yyval[count]= yy; }\n\
\n\
#endif /* YY_PART */\n\
\n\
#define	YYACCEPT	yyAccept(yythunkpos0)\n\
\n\
";

footer := "\n\
\n\
#ifndef YY_PART\n\
\n\
typedef int (*yyrule)();\n\
\n\
YY_PARSE(int) YYPARSEFROM(yyrule yystart)\n\
{\n\
  int yyok;\n\
  if (!yybuflen)\n\
    {\n\
      yybuflen= 1024;\n\
      yybuf= YYMALLOC(yybuflen);\n\
      yytextlen= 1024;\n\
      yytext= YYMALLOC(yytextlen);\n\
      yythunkslen= 32;\n\
      yythunks= YYMALLOC(sizeof(yythunk) * yythunkslen);\n\
      yyvalslen= 32;\n\
      yyvals= YYMALLOC(sizeof(YYSTYPE) * yyvalslen);\n\
      yybegin= yyend= yypos= yylimit= yythunkpos= 0;\n\
    }\n\
  yybegin= yyend= yypos;\n\
  yythunkpos= 0;\n\
  yyval= yyvals;\n\
  yyok= yystart();\n\
  if (yyok) yyDone();\n\
  yyCommit();\n\
  return yyok;\n\
  (void)yyrefill;\n\
  (void)yyText;\n\
  (void)yyCommit;\n\
  (void)yyAccept;\n\
  (void)yyPush;\n\
  (void)yyPop;\n\
  (void)yySet;\n\
  (void)yytextmax;\n\
}\n\
\n\
YY_PARSE(int) YYPARSE(void)\n\
{\n\
  return YYPARSEFROM(yy_%s);\n\
}\n\
\n\
#endif\n\
";

func Rule_compile_c_header() {
	fprintf(output, "/* A recursive-descent parser generated by peg %d.%d.%d */\n", PEG_MAJOR, PEG_MINOR, PEG_LEVEL);
	fprintf(output, "\n");
	fprintf(output, "%s", header);
	fprintf(output, "#define YYRULECOUNT %d\n", ruleCount);
}

func consumesInput(Node *node) bool {
	if !node { return false }
	switch node.type {
		case Rule:
			result := false;
			if node.reached {
				fprintf(stderr, "possible infinite left recursion in rule '%s'\n", node.name);
			} else {
				node.reached = true;
				result := consumesInput(node.expression);
				node.reached = false;
			}
			return result;

		case Dot, Class:
			return true;

		case Name:
			return consumesInput(node.name.rule);

		case Character, String:
			return len(node.string.value) > 0;

		case Action, Predicate:
			return false;

		case Alternate:
			for n := node.alternate.first;  n;  n = n.alternate.next {
				if !consumesInput(n) { return false }
			}
			return true;

		case Sequence:
			for n := node.alternate.first;  n;  n = n.alternate.next {
				if consumesInput(n) { return true }
			}
			return false;

		case PeekFor, PeekNot, Query, Star, Plus:
			return consumesInput(node.plus.element);

		default:
			fprintf(stderr, "\nconsumesInput: illegal node type %d\n", node.type);
			exit(1);
		}
	return false;
}

func Rule_compile_c(Node *node) {
	for n := range rules { consumesInput(n) }
	fprintf(output, "%s", preamble);
	for n := node;  n;  n := n.rule.next) {
		fprintf(output, "YY_RULE(int) yy_%s(); /* %d */\n", n.rule.name, n.rule.id);
	}
	fprintf(output, "\n");
	for n := range actions {
		fprintf(output, "YY_ACTION(void) yy%s(char *yytext, int yyleng)\n{\n", n.name);
		fprintf(output, "  (void)yytext; (void)yyleng;\n");
		defineVariables(n.rule.variables);
		fprintf(output, "  yyprintf((stderr, \"do yy%s\\n\"));\n", n.name);
		fprintf(output, "  %s;\n", n.text);
		undefineVariables(n.rulevariables);
		fprintf(output, "}\n");
	}
	Rule_compile_c2(node);
	fprintf(output, footer, start.name);
}