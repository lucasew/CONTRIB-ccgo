// exprUintptr assignment: (union_var = fn()).ptr_member
//
// Before the fix, ccgo failed with:
//   TODO exprUintptr (... assignmentExpression:)
//
// The expression pattern (result = fn()).ptr where result is a tagged union
// (Subtree = union of inline data and pointer) requires exprUintptr support
// in assignmentExpression's default branch.

typedef union Subtree {
	struct {
		unsigned kind, flag;
	} data;
	const void *ptr;
} Subtree;

static Subtree make_ptr(const int *p) {
	Subtree s;
	s.ptr = p;
	return s;
}

int main(void) {
	static const int g = 42;
	Subtree result;
	const int *src = &g;
	while ((result = make_ptr(src)).ptr) {
		src = 0;
	}
	return 0;
}
