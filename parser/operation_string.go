// Code generated by "stringer -type Operation"; DO NOT EDIT.

package parser

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[OperationGET-0]
	_ = x[OperationPOST-1]
	_ = x[OperationPATCH-2]
	_ = x[OperationPUT-3]
	_ = x[OperationDELETE-4]
	_ = x[OperationHEAD-5]
}

const _Operation_name = "OperationGETOperationPOSTOperationPATCHOperationPUTOperationDELETEOperationHEAD"

var _Operation_index = [...]uint8{0, 12, 25, 39, 51, 66, 79}

func (i Operation) String() string {
	if i < 0 || i >= Operation(len(_Operation_index)-1) {
		return "Operation(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Operation_name[_Operation_index[i]:_Operation_index[i+1]]
}
