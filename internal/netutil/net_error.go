package netutil

import "errors"

var (
	ERR_FAILED_TO_READ_INTERFACE  = errors.New("cannot read internet interface detail")
	ERR_FAILED_TO_READ_UDP        = errors.New("failed to read udp flow")
	ERR_FAILED_TO_SEND_UDP_PACKET = errors.New("cannot send a udp packet")
	ERROR_UDP_TIME_OUT            = errors.New("udp time out")
	ERR_FAILED_TO_COMM_UDP        = errors.New("Cannot establish udp conn!")
	ERR_NO_INTERFACE_MATCH        = errors.New("interface not found")
)
