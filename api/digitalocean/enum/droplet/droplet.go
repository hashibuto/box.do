package droplet

// DigitalOcean droplet slugs
const (
	S1VCPU1GB    = "s-1vcpu-1gb"
	S1VCPU2GB    = "s-1vcpu-2gb"
	S2VCPU3GB    = "s-1vcpu-3gb"
	S2VCPU2GB    = "s-2vcpu-2gb"
	S2VCPU4GB    = "s-2vcpu-4gb"
	S4VCPU4GB    = "s-4vcpu-8gb"
	S6VCPU16GB   = "s-6vcpu-16gb"
	S8VCPU32GB   = "s-8vcpu-32gb"
	S12VCPU48GB  = "s-12vcpu-48gb"
	S16VCPU64GB  = "s-16vcpu-64gb"
	S20VCPU96GB  = "s-20vcpu-96gb"
	S24VCPU128GB = "s-24vcpu-128gb"
	S32VCPU192GB = "s-32vcpu-192gb"
)

var validSlugs map[string]bool

// IsValid returns true if the droplet slug is valid
func IsValid(dropletSlug string) bool {
	_, ok := validSlugs[dropletSlug]
	return ok
}

// Values is an array of all the droplet slugs
var Values []string = []string{
	S1VCPU1GB,
	S1VCPU2GB,
	S2VCPU3GB,
	S2VCPU2GB,
	S2VCPU4GB,
	S4VCPU4GB,
	S6VCPU16GB,
	S8VCPU32GB,
	S12VCPU48GB,
	S16VCPU64GB,
	S20VCPU96GB,
	S24VCPU128GB,
	S32VCPU192GB,
}

// Initializes the module
func init() {
	validSlugs = make(map[string]bool)
	for _, droplet := range Values {
		validSlugs[droplet] = true
	}
}
