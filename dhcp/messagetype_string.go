// Code generated by "stringer -type=MessageType"; DO NOT EDIT.

package dhcp4

import "strconv"

const _MessageType_name = "DiscoverOfferRequestDeclineACKNAKReleaseInform"

var _MessageType_index = [...]uint8{0, 8, 13, 20, 27, 30, 33, 40, 46}

func (i MessageType) String() string {
	i -= 1
	if i >= MessageType(len(_MessageType_index)-1) {
		return "MessageType(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _MessageType_name[_MessageType_index[i]:_MessageType_index[i+1]]
}
