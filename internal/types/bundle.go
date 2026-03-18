package types

// WorkItem represents an ADO work item linked to a pull request.
type WorkItem struct {
	ID          int    `json:"id"`
	Type        string `json:"type,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	State       string `json:"state,omitempty"`
}

type BundleV1 struct {
	Version      string     `json:"version"`
	PRID         int        `json:"prId,omitempty"`
	RepoURL      string     `json:"repoUrl,omitempty"`
	Remote       string     `json:"remote,omitempty"`
	Provider     string     `json:"provider,omitempty"`
	MergeRef     string     `json:"mergeRef,omitempty"`
	Range        string     `json:"range"`
	Files        []string   `json:"files"`
	Stat         string     `json:"stat"`
	Patch        string     `json:"patch"`
	ChangedFiles int        `json:"changedFiles"`
	PatchBytes   int        `json:"patchBytes"`
	WorkItems    []WorkItem `json:"workItems,omitempty"`
}
