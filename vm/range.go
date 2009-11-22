#include "tr.h"
#include "internal.h"

OBJ TrRange_new(vm *struct TrVM, OBJ first, OBJ last, int exclusive) {
  TrRange *r = TR_INIT_CORE_OBJECT(Range);
  r.first = first;
  r.last = last;
  r.exclusive = exclusive;
  return (OBJ)r;
}

static OBJ TrRange_first(vm *struct TrVM, OBJ self) { return TR_CRANGE(self).first; }
static OBJ TrRange_last(vm *struct TrVM, OBJ self) { return TR_CRANGE(self).last; }
static OBJ TrRange_exclude_end(vm *struct TrVM, OBJ self) { return TR_BOOL(TR_CRANGE(self).exclusive); }

void TrRange_init(vm *struct TrVM) {
  OBJ c = TR_INIT_CORE_CLASS(Range, Object);
  tr_def(c, "first", TrRange_first, 0);
  tr_def(c, "last", TrRange_last, 0);
  tr_def(c, "exclude_end?", TrRange_exclude_end, 0);
}