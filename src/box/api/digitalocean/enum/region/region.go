package region

// DigitalOcean region slugs
const (
	NYC1 = "nyc1"
	NYC2 = "nyc2"
	NYC3 = "nyc3"
	AMS2 = "ams2"
	AMS3 = "ams3"
	SFO1 = "sfo1"
	SFO2 = "sfo2"
	SFO3 = "sfo3"
	SGP1 = "sgp1"
	LON1 = "lon1"
	FRA1 = "fra1"
	TOR1 = "tor1"
	BLR1 = "blr1"
)

// Mapping between slug and city/content name
var prefixToName map[string]string = map[string]string{
	"nyc": "New York City, North America",
	"ams": "Amsterdam, Europe",
	"sfo": "San Francisco, North America",
	"sgp": "Singapore, Asia",
	"lon": "London, Europe",
	"fra": "Frankfurt, Europe",
	"tor": "Toronto, North America",
	"blr": "Bangalore, Asia",
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
