#define TR_VERSION "0.0"

#include <limits.h>

/* Non portable optimizations */
#ifndef TR_COMPAT_MODE

/* Force the interpreter to store the stack and instruction
   pointer in machine registers. Works only on x86 machines. */
#if __GNUC__
#define TR_USE_MACHINE_REGS 1
#endif

#endif

/* Various limits */
#define TR_MAX_FRAMES 255
#define MAX_INT       (INT_MAX-2)  /* maximum value of an int (-2 for safety) */