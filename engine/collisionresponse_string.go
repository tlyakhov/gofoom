// Code generated by "stringer -type=CollisionResponse"; DO NOT EDIT.

package engine

import "strconv"

const _CollisionResponse_name = "SlideBounceStopRemoveCallback"

var _CollisionResponse_index = [...]uint8{0, 5, 11, 15, 21, 29}

func (i CollisionResponse) String() string {
	if i < 0 || i >= CollisionResponse(len(_CollisionResponse_index)-1) {
		return "CollisionResponse(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _CollisionResponse_name[_CollisionResponse_index[i]:_CollisionResponse_index[i+1]]
}
