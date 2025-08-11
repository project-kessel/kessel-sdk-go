package common

// ApiVersion defines an API version
type ApiVersion string

var V1 ApiVersion = "v1"
var V1Beta2 ApiVersion = "v1beta2"

// ValidApiVersions defines all the current API versions supported
var ValidApiVersions = map[ApiVersion]bool{
	V1:      true,
	V1Beta2: true,
}

// ApiOperation defines an operation the API can perform via an API endpoint
type ApiOperation string

var GetReadyZ ApiOperation = "GetReadyZ"
var GetLiveZ ApiOperation = "GetLiveZ"
var Check ApiOperation = "Check"
var CheckForUpdate ApiOperation = "CheckForUpdate"
var ReportResource ApiOperation = "ReportResource"
var DeleteResource ApiOperation = "DeleteResource"

// ValidApiOperations maps an API version to all supported operations of that API for validation purposes
var ValidApiOperations = map[ApiVersion]map[ApiOperation]bool{
	V1:      {GetReadyZ: true, GetLiveZ: true},
	V1Beta2: {Check: true, CheckForUpdate: true, ReportResource: true, DeleteResource: true},
}
