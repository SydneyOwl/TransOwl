package netutil

import "errors"

var (
	ERR_FAILED_TO_READ_INTERFACE  = errors.New("cannot read internet interface detail")
	ERR_FAILED_TO_READ_UDP        = errors.New("failed to read udp flow")
	ERR_FAILED_TO_SEND_UDP_PACKET = errors.New("cannot send a udp packet")
	ERR_NO_INTERFACE_MATCH        = errors.New("interface not found")
	ERR_TYPE_NOT_DEFINED          = errors.New("undefined type")
)
