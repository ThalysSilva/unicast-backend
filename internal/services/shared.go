package services

import "unicast-api/pkg/utils"

var customError = utils.CustomError{}
var makeError = customError.MakeError
var trace = utils.TraceError