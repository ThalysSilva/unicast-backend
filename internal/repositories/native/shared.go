package native

import "github.com/ThalysSilva/unicast-backend/pkg/utils"

var customError = &utils.CustomError{}
var MakeError = customError.MakeError
var trace = utils.TraceError