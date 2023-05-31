package cfg

import "time"

const (
	// Deprecated
	UDP_PORT = 6438

	UDP_PORT_INWARD  = 6438
	UDP_PORT_OUTWARD = 6439

	STATUS_INWARD  = 0
	STATUS_OUTWARD = 1
)
const (
	ASK_FOR_AVAILABLE_DEVICES = 1000
	ACK_BEING_DISCOVERED      = 1001
	READY_TO_SEND_FILE        = 1002
	READT_TO_RECV_FILE
)

const (
	CACHED_UDP_READ_CHANNEL_MAX_BUFFER = 5

	MAX_DEVICE_DISCOVER_TIMEOUT = 3 * time.Second

	// Deprecated since v0.0.2
	MAX_UDP_LISTENING_RETRY_CHANCES = 3

	// Deprecated since v0.0.2
	MAX_LISTENING_TIMEOUT = 500

	MAX_FILE_SEND_ACK_WAIT_TIME = 1
)

const (
	TRANSOWL_FLAG = "0x3d"

	LAN = 2000
	WAN = 2001

	NETTYPE_SAMESEGMENT = 3000
	NETTYPE_WHOLENET    = 3001
)

const (
	STATUS_OK        = 4001
	STATUS_RECV_FILE = 4002
	STATUS_SEND_FILE = 4003
)