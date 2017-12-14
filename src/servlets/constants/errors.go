// this file defines the http return code

package constants

// HTTP return code constants
const (
	RC_OK           = 0     // ok
	RC_PROTOCOL_ERR = 10001 // protocol error
)

const (
	ERR_INT_OK           = 0 //internal errors
	ERR_INT_TK_DB        = -1
	ERR_INT_TK_DUPLICATE = -2
	ERR_INT_TK_NOTEXISTS = -3
)

const (
	ERR_EXT_OK = 0 //external errors
)
