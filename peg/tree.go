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
 * Last edited: 2007-05-15 10:32:05 by piumarta on emilia
 */

#include <stdio.h>

enum { Unknown= 0, Rule, Variable, Name, Dot, Character, String, Class, Action, Predicate, Alternate, Sequence, PeekFor, PeekNot, Query, Star, Plus };

enum {
  RuleUsed	= 1<<0,
  RuleReached	= 1<<1,
};

package pegleg

struct Rule	 {
	string		name;
	variables	Vector;
	expression	*Node;
	id			int;
//	flags		int;
	true		bool;
}

func makeRule(name string) Rule {
	ruleCount++;
	node := Rule{name: name, id: ruleCount, flags: 0}
	rules.Push(node);
	return node;
}

func findRule(name string) Rule {
	rule_name = strings.Join(strings.Split(name, '-', 0), "_");
	for rule := range rules.Iter() {
		if rule_name == rule.name { return rule; }
	}
	return makeRule(name);
}

func beginRule(rule *Rule) *Rule {
	actionCount = 0;
	return thisRule = rule;
}

func (self *Rule) setExpression(expression *Node) {
	self.expression = expression;
	if !start || self.name == "start" { start = self; }
}

func (self *Rule) fprint(stream *FILE) {
	fprintf(stream, "%s.%d =", self.name, self.id);
	if self.expression {
		Node_fprint(stream, self.expression);
	} else {
		fprintf(stream, " UNDEFINED");
	}
	fprintf(stream, " ;\n");
}

func (self *Rule) print() {
	self.fprint(stderr);
}

struct Variable	 {
	Node *next;
	char *name;
	Node *value;
	int offset;
};

func findVariable(name string) *Variable {
	for node := thisRule.variables.Iter() {
		if name == node.name { return node; }
	}
	return nil
}

func newVariable(name string) *Variable {
	thisRule.variables.Push(findVariable(name) || Variable{name: name});
	return thisRule.variables.Last();
}

struct Name	 {
	Node *next;
	Node *rule;
	Node *variable;
};

struct Dot	 {
	Node *next;
};

struct Character {
	Node *next;
	char *value;
};

struct String	 {
	Node *next;
	char *value;
};

struct Class	 {
	Node *next;
	unsigned char *value;
};

struct Action	 {
	Node *next;
	char *text;
	Node *list;
	char *name;
	Node *rule;
};

struct Predicate {
	Node *next;
	char *text;
};

struct Alternate {
	Node *next;
	Node *first;
	Node *last;
};

struct Sequence	 {
	Node *next;
	Node *first;
	Node *last;
};

struct PeekFor	 {
	Node *next;
	Node *element;
};

struct PeekNot	 {
	Node *next;
	Node *element;
};

struct Query	 {
	Node *next;
	Node *element;
};

struct Star	 {
	Node *next;
	Node *element;
};

struct Plus	 {
	Node *next;
	Node *element;
};

struct Any	 {
	Node *next;
};

actions := Vector.new();
rules := Vector.New(0);
thisRule *Rule;

Node *start= 0;
FILE *output= 0;

int actionCount= 0;
int ruleCount= 0;
int lastToken= -1;

import "strings";

func makeAction(text string) *Node {
	name := new([1024]byte);
	sprintf(name, "_%d_%s", actions.Len(), thisRule.name);
	actions.Push(Action{name: name, text: strings.Join(strings.Split(text, '$$', 0), 'yy'), list: actions, rule: thisRule})
	return node;
}

func makeAlternate(e *Node) *Node {
	if Alternate != e.type {
		node := Alternate{};
		assert(e);
		assert(!e.any.next);
		node.alternate.first = node.alternate.last = e;
		return node;
	}
	return e;
}

func (node *Alternate) *Alternate {
	a := makeAlternate(node);
}

func Alternate_append(a, e *Node) *Node {
	a := makeAlternate(a);
	assert(a.alternate.last);
	assert(e);
	a.alternate.last.any.next = e;
	a.alternate.last = e;
	return a;
}

func makeSequence(e *Node) *Node {
	if Sequence != e.type {
		node := Sequence{};
		assert(e);
		assert(!e.any.next);
		node.sequence.first = node.sequence.last = e;
		return node;
	}
	return e;
}

func Sequence_append(a, e *Node) *Node {
	assert(a);
	a := makeSequence(a);
	assert(a.sequence.last);
	assert(e);
	a.sequence.last.any.next = e;
	a.sequence.last = e;
	return a;
}

stack := Vector.New(1024);

func showStack() {
	index := 0;
	for node := range stack.Iter() {
		fprintf(stderr, "### %d\t", index);
		Node_print(node);
		fprintf(stderr, "\n");
		index++;
	}
}

func push(node *Node) *Node {
	assert(node);
	return *++stackPointer= node;
}

func top() *Node {
	assert(stackPointer > stack);
	return *stackPointer;
}

func pop() *Node {
	assert(stackPointer > stack);
	return *stackPointer--;
}


func Node_fprint(stream *FILE, node *Node) {
	assert(node);
	switch node.type {
		case Rule:		fprintf(stream, " %s", node->rule.name);
		case Name:		fprintf(stream, " %s", node->name.rule->rule.name);
		case Dot:		fprintf(stream, " .");
		case Character:	fprintf(stream, " '%s'", node->character.value);
		case String:	fprintf(stream, " \"%s\"", node->string.value);
		case Class:		fprintf(stream, " [%s]", node->cclass.value);
		case Action:	fprintf(stream, " { %s }", node->action.text);
		case Predicate:	fprintf(stream, " ?{ %s }", node->action.text);
		case Alternate:	node= node->alternate.first;
			fprintf(stream, " (");
			Node_fprint(stream, node);
			while ((node= node->any.next)) {
				fprintf(stream, " |");
				Node_fprint(stream, node);
			}
			fprintf(stream, " )");
		case Sequence:	node= node->sequence.first;
			fprintf(stream, " (");
			Node_fprint(stream, node);
			while node = node->any.next {
				Node_fprint(stream, node);
			}
			fprintf(stream, " )");
			break;

		case PeekFor:	fprintf(stream, "&");  Node_fprint(stream, node->query.element);
		case PeekNot:	fprintf(stream, "!");  Node_fprint(stream, node->query.element);
		case Query:		Node_fprint(stream, node->query.element);  fprintf(stream, "?");
		case Star:		Node_fprint(stream, node->query.element);  fprintf(stream, "*");
		case Plus:		Node_fprint(stream, node->query.element);  fprintf(stream, "+");
		default:
			fprintf(stream, "\nunknown node type %d\n", node->type);
			exit(1);
	}
}

func Node_print(Node *node)	{
	Node_fprint(stderr, node);
}