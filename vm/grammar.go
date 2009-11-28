%{
#include <stdlib.h>
#include "tr.h"

/*#define YY_DEBUG 1*/

#define YYSTYPE   RubyObject
#define yyvm      compiler.vm

charbuf *string;
sbuf *string;
nbuf size_t;
compiler *Compiler;

#define YY_INPUT(buf, result, max_size) {	\
	yyc int;	\
	if charbuf && *charbuf != '\0' {	\
		yyc= *charbuf++;	\
	} else {	\
		yyc= EOF;	\
	}	\
	if EOF == yyc {	\
		result := 0;	\
	} else {	\
		(*(buf)= yyc, 1);	\
	}	\
}

// TODO grow buffer
#define STRING_START sbuf = make([]byte, 4096); nbuf = 0
%}

Root      = s:Stmts EOF                     { compiler.node = newASTNode(compiler.vm, NODE_ROOT, s, 0, 0, compiler.line) }

Stmts     = SEP*
            - head:Stmt Comment?            { head = compiler.vm.newArray2(1, head) }
            ( SEP - tail:Stmt Comment?      { head.Push(tail) }
            | SEP - Comment
            )* SEP?                         { $$ = head }
          | SEP+                            { $$ = compiler.vm.newArray2(0) }

OptStmts  = Stmts
          | - SEP?                          { $$ = compiler.vm.newArray2(0) }

Stmt      = While
          | Until
          | If
          | Unless
          | Def
          | Class
          | Module
          | Expr

Expr      = Assign
          | AsgnCall
          | UnaryOp
          | BinOp
          | SpecCall
          | Call
          | Range
          | Yield
          | Return
          | Break
          | Value

Comment   = - '#' (!EOL .)*

Call      =                                 { block = rcv = 0 }
            ( rcv:Value '.'
            )? ( rmsg:Message '.'           { rcv = newASTNode(compiler.vm, NODE_SEND, rcv, rmsg, 0, compiler.line) }
               )* msg:Message
                  - block:Block?            { $$ = newASTNode(compiler.vm, NODE_SEND, rcv, msg, block, compiler.line) }

# TODO refactor head part w/ Call maybe eh?
AsgnCall   =                                { rcv = 0 }
            ( rcv:Value '.'
            )? ( rmsg:Message '.'           { rcv = newASTNode(compiler.vm, NODE_SEND, rcv, rmsg, 0, compiler.line) }
               )* msg:ID - asg:ASSIGN
                  - val:Stmt                { vm = RubyVM *(yyvm); $$ = newASTNode(compiler.vm, NODE_SEND, rcv, newASTNode(compiler.vm, NODE_MSG, TrSymbol_new(vm, TrString *(msg).ptr, TrString *(asg).ptr), compiler.vm.newArray2(1, newASTNode(compiler.vm, NODE_ARG, val, 0, 0, compiler.line)), 0, compiler.line), 0, compiler.line) }

Receiver  = (                               { rcv = 0 }
              rcv:Call
            | rcv:Value
            )                               { $$ = rcv }

SpecCall  = rcv:Receiver '[' args:Args ']'  
            - ASSIGN - val:Stmt             { args.Push(newASTNode(compiler.vm, NODE_ARG, val, 0, 0, compiler.line)); $$ = newASTNode(compiler.vm, NODE_SEND, rcv, newASTNode(compiler.vm, NODE_MSG, TrSymbol_new(yyvm, "[]="), args, 0, compiler.line), 0, compiler.line) }
          | rcv:Receiver '[' args:Args ']'  { $$ = newASTNode(compiler.vm, NODE_SEND, rcv, newASTNode(compiler.vm, NODE_MSG, TrSymbol_new(yyvm, "[]"), args, 0, compiler.line), 0, compiler.line) }

BinOp     = ( rcv:SpecCall | rcv:Receiver )
            -
            (
              '&&' - arg:Expr               { $$ = newASTNode(compiler.vm, NODE_AND, rcv, arg, 0, compiler.line) }
            | '||' - arg:Expr               { $$ = newASTNode(compiler.vm, NODE_OR, rcv, arg, 0, compiler.line) }
            | '+' - arg:Expr                { $$ = newASTNode(compiler.vm, NODE_ADD, rcv, arg, 0, compiler.line) }
            | '-' - arg:Expr                { $$ = newASTNode(compiler.vm, NODE_SUB, rcv, arg, 0, compiler.line) }
            | '<' - arg:Expr                { $$ = newASTNode(compiler.vm, NODE_LT, rcv, arg, 0, compiler.line) }
            | op:BINOP - arg:Expr           { $$ = newASTNode(compiler.vm, NODE_SEND, rcv, newASTNode(compiler.vm, NODE_MSG, op, compiler.vm.newArray2(1, newASTNode(compiler.vm, NODE_ARG, arg, 0, 0, compiler.line)), 0, compiler.line), 0, compiler.line) }
            ) 

UnaryOp   = '-' rcv:Expr                    { $$ = newASTNode(compiler.vm, NODE_NEG, rcv, 0, 0, compiler.line) }
          | '!' rcv:Expr                    { $$ = newASTNode(compiler.vm, NODE_NOT, rcv, 0, 0, compiler.line) }

Message   = name:ID                         { args = 0 }
              ( '(' args:Args? ')'
              | SPACE args:Args
              )?                            { $$ = newASTNode(compiler.vm, NODE_MSG, name, args, 0, compiler.line) }

Args      = - head:Expr -                   { head = compiler.vm.newArray2(1, newASTNode(compiler.vm, NODE_ARG, head, 0, 0, compiler.line)) }
            ( ',' - tail:Expr -             { head.Push(newASTNode(compiler.vm, NODE_ARG, tail, 0, 0, compiler.line)) }
            )* ( ',' - '*' splat:Expr -     { head.Push(newASTNode(compiler.vm, NODE_ARG, splat, 1, 0, compiler.line)) }
               )?                           { $$ = head }
          | - '*' splat:Expr -              { $$ = compiler.vm.newArray2(1, newASTNode(compiler.vm, NODE_ARG, splat, 1, 0, compiler.line)) }

Block     = 'do' SEP
              - body:OptStmts -
            'end'                           { $$ = newASTNode(compiler.vm, NODE_BLOCK, body, 0, 0, compiler.line) }
          | 'do' - '|' params:Params '|' SEP
              - body:OptStmts -
            'end'                           { $$ = newASTNode(compiler.vm, NODE_BLOCK, body, params, 0, compiler.line) }
          # FIXME this might hang the parser and is very slow.
          # Clash with Hash for sure.
          #| '{' - body:OptStmts - '}'       { $$ = newASTNode(compiler.vm, NODE_BLOCK, body, 0, 0, compiler.line) }
          #| '{' - '|' params:Params '|'
          #  - body:OptStmts - '}'           { $$ = newASTNode(compiler.vm, NODE_BLOCK, body, params, 0, compiler.line) }

Assign    = name:ID - ASSIGN - val:Stmt     { $$ = newASTNode(compiler.vm, NODE_ASSIGN, name, val, 0, compiler.line) }
          | name:CONST - ASSIGN - val:Stmt  { $$ = newASTNode(compiler.vm, NODE_SETCONST, name, val, 0, compiler.line) }
          | name:IVAR - ASSIGN - val:Stmt   { $$ = newASTNode(compiler.vm, NODE_SETIVAR, name, val, 0, compiler.line) }
          | name:CVAR - ASSIGN - val:Stmt   { $$ = newASTNode(compiler.vm, NODE_SETCVAR, name, val, 0, compiler.line) }
          | name:GLOBAL - ASSIGN - val:Stmt { $$ = newASTNode(compiler.vm, NODE_SETGLOBAL, name, val, 0, compiler.line) }

While     = 'while' SPACE cond:Expr SEP
              body:Stmts -
            'end'                           { $$ = newASTNode(compiler.vm, NODE_WHILE, cond, body, 0, compiler.line) }

Until     = 'until' SPACE cond:Expr SEP
              body:Stmts -
            'end'                           { $$ = newASTNode(compiler.vm, NODE_UNTIL, cond, body, 0, compiler.line) }

If        = 'if' SPACE cond:Expr SEP        { else_body = 0 }
              body:Stmts -
            else_body:Else?
            'end'                           { $$ = newASTNode(compiler.vm, NODE_IF, cond, body, else_body, compiler.line) }
          | body:Expr - 'if' - cond:Expr    { $$ = newASTNode(compiler.vm, NODE_IF, cond, compiler.vm.newArray2(1, body), 0, compiler.line) }

Unless    = 'unless' SPACE cond:Expr SEP    { else_body = 0 }
              body:Stmts -
            else_body:Else?
            'end'                           { $$ = newASTNode(compiler.vm, NODE_UNLESS, cond, body, else_body, compiler.line) }
          | body:Expr -
              'unless' - cond:Expr          { $$ = newASTNode(compiler.vm, NODE_UNLESS, cond, compiler.vm.newArray2(1, body), 0, compiler.line) }

Else      = 'else' SEP - body:Stmts -       { $$ = body }

Method    = rcv:ID '.' name:METHOD          { $$ = newASTNode(compiler.vm, NODE_METHOD, newASTNode(compiler.vm, NODE_SEND, 0, newASTNode(compiler.vm, NODE_MSG, rcv, 0, 0, compiler.line), 0, compiler.line), name, 0, compiler.line) }
          | rcv:Value '.' name:METHOD       { $$ = newASTNode(compiler.vm, NODE_METHOD, rcv, name, 0, compiler.line) }
          | name:METHOD                     { $$ = newASTNode(compiler.vm, NODE_METHOD, 0, name, 0, compiler.line) }

Def       = 'def' SPACE method:Method       { params = 0 }
            (- '(' params:Params? ')')? SEP
              body:OptStmts -
            'end'                           {	if params > 0 {
													$$ := newASTNode(compiler.vm, NODE_DEF, method, params, body, compiler.line);
												} else {
													$$ := newASTNode(compiler.vm, NODE_DEF, method, compiler.vm.newArray2(0), body, compiler.line);
												}
											}

Params    = head:Param                      { head = compiler.vm.newArray2(1, head) }
            ( ',' tail:Param                { head.Push(tail) }
            )*                              { $$ = head }

Param     = - name:ID - '=' - def:Expr      { $$ = newASTNode(compiler.vm, NODE_PARAM, name, 0, def, compiler.line) }
          | - name:ID -                     { $$ = newASTNode(compiler.vm, NODE_PARAM, name, 0, 0, compiler.line) }
          | - '*' name:ID -                 { $$ = newASTNode(compiler.vm, NODE_PARAM, name, 1, 0, compiler.line) }

Class     = 'class' SPACE name:CONST        { super = 0 }
            (- '<' - super:CONST)? SEP
              body:OptStmts -
            'end'                           { $$ = newASTNode(compiler.vm, NODE_CLASS, name, super, body, compiler.line) }

Module    = 'module' SPACE name:CONST SEP
              body:OptStmts -
            'end'                           { $$ = newASTNode(compiler.vm, NODE_MODULE, name, 0, body, compiler.line) }

Range     = s:Receiver - '..' - e:Expr      { $$ = newASTNode(compiler.vm, NODE_RANGE, s, e, 0, compiler.line) }
          | s:Receiver - '...' - e:Expr     { $$ = newASTNode(compiler.vm, NODE_RANGE, s, e, 1, compiler.line) }

Yield     = 'yield' SPACE args:AryItems     { $$ = newASTNode(compiler.vm, NODE_YIELD, args, 0, 0, compiler.line) }
          | 'yield' '(' args:AryItems ')'   { $$ = newASTNode(compiler.vm, NODE_YIELD, args, 0, 0, compiler.line) }
          | 'yield'                         { $$ = newASTNode(compiler.vm, NODE_YIELD, compiler.vm.newArray2(0), 0, 0, compiler.line) }

Return    = 'return' SPACE arg:Expr - !','  { $$ = newASTNode(compiler.vm, NODE_RETURN, arg, 0, 0, compiler.line) }
          | 'return' '(' arg:Expr ')' - !','{ $$ = newASTNode(compiler.vm, NODE_RETURN, arg, 0, 0, compiler.line) }
          | 'return' SPACE args:AryItems    { $$ = newASTNode(compiler.vm, NODE_RETURN, newASTNode(compiler.vm, NODE_ARRAY, args, 0, 0, compiler.line), 0, 0, compiler.line) }
          | 'return' '(' args:AryItems ')'  { $$ = newASTNode(compiler.vm, NODE_RETURN, newASTNode(compiler.vm, NODE_ARRAY, args, 0, 0, compiler.line), 0, 0, compiler.line) }
          | 'return'                        { $$ = newASTNode(compiler.vm, NODE_RETURN, 0, 0, 0, compiler.line) }

Break     = 'break'                         { $$ = newASTNode(compiler.vm, NODE_BREAK, 0, 0, 0, compiler.line) }

Value     = v:NUMBER                        { $$ = newASTNode(compiler.vm, NODE_VALUE, v, 0, 0, compiler.line) }
          | v:SYMBOL                        { $$ = newASTNode(compiler.vm, NODE_VALUE, v, 0, 0, compiler.line) }
          | v:REGEXP                        { $$ = newASTNode(compiler.vm, NODE_VALUE, v, 0, 0, compiler.line) }
          | v:STRING1                       { $$ = newASTNode(compiler.vm, NODE_STRING, v, 0, 0, compiler.line) }
          | v:STRING2                       { $$ = newASTNode(compiler.vm, NODE_STRING, v, 0, 0, compiler.line) }
          | v:CONST                         { $$ = newASTNode(compiler.vm, NODE_CONST, v, 0, 0, compiler.line) }
          | 'nil'                           { $$ = newASTNode(compiler.vm, NODE_NIL, 0, 0, 0, compiler.line) }
          | 'true'                          { $$ = newASTNode(compiler.vm, NODE_BOOL, TR_TRUE, 0, 0, compiler.line) }
          | 'false'                         { $$ = newASTNode(compiler.vm, NODE_BOOL, TR_FALSE, 0, 0, compiler.line) }
          | 'self'                          { $$ = newASTNode(compiler.vm, NODE_SELF, 0, 0, 0, compiler.line) }
          | name:IVAR                       { $$ = newASTNode(compiler.vm, NODE_GETIVAR, name, 0, 0, compiler.line) }
          | name:CVAR                       { $$ = newASTNode(compiler.vm, NODE_GETCVAR, name, 0, 0, compiler.line) }
          | name:GLOBAL                     { $$ = newASTNode(compiler.vm, NODE_GETGLOBAL, name, 0, 0, compiler.line) } # TODO
          | '[' - ']'                       { $$ = newASTNode(compiler.vm, NODE_ARRAY, compiler.vm.newArray2(0), 0, 0, compiler.line) }
          | '[' - items:AryItems - ']'      { $$ = newASTNode(compiler.vm, NODE_ARRAY, items, 0, 0, compiler.line) }
          | '{' - '}'                       { $$ = newASTNode(compiler.vm, NODE_HASH, compiler.vm.newArray2(0), 0, 0, compiler.line) }
          | '{' - items:HashItems - '}'     { $$ = newASTNode(compiler.vm, NODE_HASH, items, 0, 0, compiler.line) }
          | '(' - Expr - ')'

AryItems  = - head:Expr -                   { head = compiler.vm.newArray2(1, head) }
            ( ',' - tail:Expr -             { head.Push(tail) }
            )*                              { $$ = head }

HashItems = head:Expr - '=>' - val:Expr     { head = compiler.vm.newArray2(2, head, val) }
            ( - ',' - key:Expr -            { head.Push(key) }
                '=>' - val:Expr             { head.Push(val) }
            )*                              { $$ = head }

KEYWORD   = 'while' | 'until' | 'do' | 'end' |
            'if' | 'unless' | 'else' |
            'true' | 'false' | 'nil' | 'self' |
            'class' | 'module' | 'def' |
            'yield' | 'return' | 'break'

NAME      = [a-zA-Z0-9_]+
ID        = !'self'                         # self is special, can never be a method name
            < KEYWORD > &('.' | '(' | '[')  { $$ = TrSymbol_new(yyvm, yytext) } # hm, there's probably a better way
          | < KEYWORD NAME >                { $$ = TrSymbol_new(yyvm, yytext) }
          | !KEYWORD
            < [a-z_] NAME?
              ( '=' &'(' | '!'| '?' )? >    { $$ = TrSymbol_new(yyvm, yytext) }
CONST     = < [A-Z] NAME? >                 { $$ = TrSymbol_new(yyvm, yytext) }
BINOP     = < ( '**' | '^'  | '&'  | '|'  | '~'  |
                '+'  | '-'  | '*'  | '/'  | '%'  | '<=>' |
                '<<' | '>>' | '==' | '=~' | '!=' | '===' |
                '<'  | '>'  | '<=' | '>='
              ) >                           { $$ = TrSymbol_new(yyvm, yytext) }
UNOP      = < ( '-@' | '!' ) >              { $$ = TrSymbol_new(yyvm, yytext) }
METHOD    = ID | UNOP | BINOP
ASSIGN    = < '=' > &(!'=')                 { $$ = TrSymbol_new(yyvm, yytext) }
IVAR      = < '@' NAME >                    { $$ = TrSymbol_new(yyvm, yytext) }
CVAR      = < '@@' NAME >                   { $$ = TrSymbol_new(yyvm, yytext) }
GLOBAL    = < '$' NAME >                    { $$ = TrSymbol_new(yyvm, yytext) }
NUMBER    = < [0-9]+ >                      { $$ = TR_INT2FIX(atoi(yytext)) }
SYMBOL    = ':' < (NAME | KEYWORD) >        { $$ = TrSymbol_new(yyvm, yytext) }

STRING1   = '\''                            { STRING_START }
            (
              '\\\''                        { assert(nbuf + 1 < 4096); memcpy(sbuf + nbuf, "'", sizeof(char) * 1); nbuf += 1; }
            | < [^\'] >                     { assert(nbuf + yyleng < 4096); memcpy(sbuf + nbuf, yytext, sizeof(char) * yyleng); nbuf += yyleng; }
            )* '\''                         { $$ = TrString_new2(yyvm, sbuf) }

ESC_CHAR  = '\\n'                           { assert(nbuf + 1 < 4096); memcpy(sbuf + nbuf, "\n", sizeof(char) * 1); nbuf += 1; }
          | '\\b'                           { assert(nbuf + 1 < 4096); memcpy(sbuf + nbuf, "\b", sizeof(char) * 1); nbuf += 1; }
          | '\\f'                           { assert(nbuf + 1 < 4096); memcpy(sbuf + nbuf, "\f", sizeof(char) * 1); nbuf += 1; }
          | '\\r'                           { assert(nbuf + 1 < 4096); memcpy(sbuf + nbuf, "\r", sizeof(char) * 1); nbuf += 1; }
          | '\\t'                           { assert(nbuf + 1 < 4096); memcpy(sbuf + nbuf, "\t", sizeof(char) * 1); nbuf += 1; }
          | '\\\"'                          { assert(nbuf + 1 < 4096); memcpy(sbuf + nbuf, "\"", sizeof(char) * 1); nbuf += 1; }
          | '\\\\'                          { assert(nbuf + 1 < 4096); memcpy(sbuf + nbuf, "\\", sizeof(char) * 1); nbuf += 1; }

STRING2   = '"'                             { STRING_START }
            (
              ESC_CHAR
            | < [^\"] >                     { assert(nbuf + yyleng < 4096); memcpy(sbuf + nbuf, yytext, sizeof(char) * yyleng); nbuf += yyleng; }  #" for higlighting
            )*
            '"'                             { $$ = TrString_new2(yyvm, sbuf) }

REGEXP    = '/'                             { STRING_START }
            (
              ESC_CHAR
            | < [^/] >                      { assert(nbuf + yyleng < 4096); memcpy(sbuf + nbuf, yytext, sizeof(char) * yyleng); nbuf += yyleng; }
            )*
            '/'                             { $$ = TrRegexp_new(yyvm, sbuf, 0) }

-         = [ \t]*
SPACE     = [ ]+
EOL       = ( '\n' | '\r\n' | '\r' )        { compiler.line++ }
EOF       = !.
SEP       = ( - Comment? (EOL | ';') )+

%%

/* Raise a syntax error. */
RubyObject yyerror() {
	vm := RubyVM *(yyvm);
	if !compiler.filename.(String) && !compiler.filename.(Symbol) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + compiler.filename));
		return TR_UNDEF;
	}
	msg := tr_sprintf(vm, "SyntaxError in %s at line %d", compiler.filename.ptr, compiler.line);
 	// Stupid ugly code, just to build a string... I suck...
	if yytext[0] { TrString_push(vm, msg, tr_sprintf(vm, " near token '%s'", yytext)); }
  	if yypos < yylimit {
		yybuf[yylimit]= '\0';
		TrString_push(vm, msg, tr_sprintf(vm, " before text \""));
		while (yypos < yylimit) {
			if '\n' == yybuf[yypos] || '\r' == yybuf[yypos] { break; }
			char c[2] = { yybuf[yypos++], '\0' };
			TrString_push(vm, msg, tr_sprintf(vm, c));
		}
		TrString_push(vm, msg, tr_sprintf(vm, "\""));
	}
 	// TODO msg should not be a String object
	if !msg.(String) && !msg.(Symbol) {
		vm.throw_reason = TR_THROW_EXCEPTION;
		vm.throw_value = TrException_new(vm, vm.cTypeError, TrString_new2(vm, "Expected " + msg));
		return TR_UNDEF;
	}
	vm.throw_reason = TR_THROW_EXCEPTION;
	vm.throw_value = TrException_new(vm, vm.cSyntaxError, tr_sprintf(vm, msg.ptr));
	return TR_UNDEF;
}

/* Compiles code to a Block.
   Returns NULL on error, error is stored in TR_EXCEPTION. */
func Block_compile(vm *RubyVM, code *string, fn *string, lineno size_t) Block * {
	assert(!compiler && "parser not reentrant");
	charbuf = code;
	compiler = newCompiler(vm, fn);
	compiler.line += lineno;
	compiler.filename = TrString_new2(vm, fn);
	Block *b = NULL;

	if yyparse() {
		compiler.compile
		b = compiler.block;		
	} else {}
		yyerror();
	}
	charbuf = 0;
	compiler = 0;
	return b;
}