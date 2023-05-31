package related_resp

import "errors"

var ERR_FAILED_TO_PARSE_INPUT = errors.New("cannot parse to target")
var ERR_TRANSOWL_FLAG_WRONG = errors.New("not a valid transowl flag")
