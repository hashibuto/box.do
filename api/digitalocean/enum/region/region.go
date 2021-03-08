package region

// DigitalOcean region slugs
const (
	NYC1 = "NYC1"
	NYC2 = "NYC2"
	NYC3 = "NYC3"
	AMS2 = "AMS2"
	AMS3 = "AMS3"
	SFO1 = "SFO1"
	SFO2 = "SFO2"
	SFO3 = "SFO3"
	SGP1 = "SGP1"
	LON1 = "LON1"
	FRA1 = "FRA1"
	TOR1 = "TOR1"
	BLR1 = "BLR1"
)

// Mapping between slug and city/content name
var prefixToName map[string]string = map[string]string{
	"NYC": "New York City, North America",
	"AMS": "Amsterdam, Europe",
	"SFO": "San Francisco, North America",
	"SGP": "Singapore, Asia",
	"LON": "London, Europe",
	"FRA": "Frankfurt, Europe",
	"TOR": "Toronto, North America",
	"BLR": "Bangalore, Asia",
}

var regionToName map[string]string

// GetName returns the full name associated with the region slug
func GetName(regionSlug string) string {
	return regionToName[regionSlug]
}

// IsValid returns true if the regionSlug is valid
func IsValid(regionSlug string) bool {
	_, ok := regionToName[regionSlug]
	return ok
}

// Values is an array of all the region slugs
var Values []string = []string{
	NYC1,
	NYC2,
	NYC3,
	AMS2,
	AMS3,
	SFO1,
	SFO2,
	SFO3,
	SGP1,
	LON1,
	FRA1,
	TOR1,
	BLR1,
}

// Initializes the module
func init() {
	regionToName = make(map[string]string)
	for _, region := range Values {
		prefix := region[:3]
		regionToName[region] = prefixToName[prefix]
	}
}
