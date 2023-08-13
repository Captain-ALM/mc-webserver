package info

var BuildName string
var BuildDescription string
var BuildVersion string
var BuildDate string

func SetupProductInfo(buildName string, buildDescription string, buildVersion string, buildDate string) {
	BuildName = buildName
	BuildDescription = buildDescription
	BuildVersion = buildVersion
	BuildDate = buildDate
}
