package services

import "github.com/ThalysSilva/unicast-backend/pkg/utils"

var customError = utils.CustomError{}
var makeError = customError.MakeError
var trace = utils.TraceError