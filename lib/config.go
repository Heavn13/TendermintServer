package lib

/*
Server Configuration
*/
var (
	ConfigIsLocal   = true
	ConfigIsDevelop = true
	ConfigIsProduct = false

	ConfigIsDebug   = true || ConfigIsLocal
	ConfigIsRelease = !ConfigIsDebug
)

/*
MongoDB Configuration
*/
var (
	ConfigIsDB  = true
	DbConnMongo = "mongodb://localhost:27017"
)
