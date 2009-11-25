#define kv_pushp(type, v) (((v).n == (v).m)?							\
						   ((v).m = ((v).m? (v).m<<1 : 2),				\
							(v).a = (type*)TR_REALLOC((v).a, sizeof(type) * (v).m), 0)	\
						   : 0), ((v).a + ((v).n++))

