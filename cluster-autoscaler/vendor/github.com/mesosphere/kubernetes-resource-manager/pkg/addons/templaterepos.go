package addons

// -----------------------------------------------------------------------------
// TemplateRepos - Public Types
// -----------------------------------------------------------------------------

// TemplateRepo defines a git repo from which to assemble and Addon
type TemplateRepo struct {
	// URL is the git URL for the repo that contains a /template directory containing Addon definitions
	URL string

	// Priority is the order in which to apply the Addon templates, in decending order: 1 is applied after 99.
	Priority int

	// Tag is the tag version that should be used to retrieve Addon definitions
	Tag string
}

// TemplateRepos provides a list of TemplateRepo repos
type TemplateRepos []TemplateRepo

// -----------------------------------------------------------------------------
// TemplateRepos - Sort Implementation
// -----------------------------------------------------------------------------

func (t TemplateRepos) Len() int {
	return len(t)
}

func (t TemplateRepos) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t TemplateRepos) Less(i, j int) bool {
	return t[i].Priority < t[j].Priority
}
