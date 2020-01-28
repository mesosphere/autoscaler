package status

// Status represents the operational status of an addon
type Status string

const (
	// Empty status indicates the addon has not tried to install yet
	Empty Status = ""

	// Outdated status indicates that the addon needs to update
	Outdated Status = "outdated"

	// Deploying status indicates that the addon is still undergoing deployment
	Deploying Status = "deploying"

	// Deployed status indicates the addon is fully deployed and no operations are underway
	Deployed Status = "deployed"

	// Cleaning status indicates that there are cleanup operations underway
	Cleaning Status = "cleaning"

	// Absent status indicates that the addons components are not present on the cluster
	Absent Status = "absent"

	// Failed status indicates that an operation on the addon failed to complete
	Failed Status = "failed"

	// Unknown status indicates that the addon status can not currently be determined
	Unknown Status = "unknown"
)

// FromStr produces a Status given a string. If that string does not match any known status
// the status "Unknown" will be provided.
func FromStr(s string) Status {
	for _, status := range []Status{Deploying, Deployed, Cleaning, Absent, Failed} {
		if s == string(status) {
			return status
		}
	}
	return Unknown
}
