#define TR_ALLOC(T)          (T *)TR_MALLOC(sizeof(T))
#define TR_ALLOC_N(T,N)      (T *)TR_MALLOC(sizeof(T)*(N))

#define TR_MEMZERO(X,T)      memset((X), 0, sizeof(T))
#define TR_MEMZERO_N(X,T,N)  memset((X), 0, sizeof(T)*(N))
#define TR_MEMCPY(X,Y,T)     memcpy((X), (Y), sizeof(T))
#define TR_MEMCPY_N(X,Y,T,N) memcpy((X), (Y), sizeof(T)*(N))

/* ast building macros */
#define NODE(T,A)            newASTNode(compiler.vm, NODE_##T, (A), 0, 0, compiler.line)
#define NODE2(T,A,B)         newASTNode(compiler.vm, NODE_##T, (A), (B), 0, compiler.line)
#define NODE3(T,A,B,C)       newASTNode(compiler.vm, NODE_##T, (A), (B), (C), compiler.line)
#define NODES(I)             newArray2(compiler.vm, 1, (I))
#define NODES_N(N,...)       newArray2(compiler.vm, (N), ##__VA_ARGS__)
#define PUSH_NODE(A,N)       (A).kv.Push(N)
#define SYMCAT(A,B)          tr_intern(strcat(((TrString*)(A)).ptr, ((TrString*)(B)).ptr))

/* This provides the compiler about branch hints, so it
   keeps the normal case fast. Stolen from Rubinius. */
#ifdef __GNUC__
#define likely(x)       __builtin_expect((long int)(x),1)
#define unlikely(x)     __builtin_expect((long int)(x),0)
#else
#define likely(x) x
#define unlikely(x) x
#endif

/* types of nodes in the AST built by the parser */
const (
	NODE_ROOT = iota;
	NODE_BLOCK;
	NODE_VALUE;
	NODE_STRING;
	NODE_ASSIGN;
	NODE_ARG;
	NODE_SEND;
	NODE_MSG;
	NODE_IF;
	NODE_UNLESS;
	NODE_AND;
	NODE_OR;
	NODE_WHILE;
	NODE_UNTIL;
	NODE_BOOL;
	NODE_NIL;
	NODE_SELF;
	NODE_LEAVE;
	NODE_RETURN;
	NODE_BREAK;
	NODE_YIELD;
	NODE_DEF;
	NODE_METHOD;
	NODE_PARAM;
	NODE_CLASS;
	NODE_MODULE;
	NODE_CONST;
	NODE_SETCONST;
	NODE_ARRAY;
	NODE_HASH;
	NODE_RANGE;
	NODE_GETIVAR;
	NODE_SETIVAR;
	NODE_GETCVAR;
	NODE_SETCVAR;
	NODE_GETGLOBAL;
	NODE_SETGLOBAL;
	NODE_ADD;
	NODE_SUB;
	NODE_LT;
	NODE_NEG;
	NODE_NOT;
)