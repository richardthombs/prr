package types

type BundleV1 struct {
	Version      string     `json:"version"`
	PRID         int        `json:"prId,omitempty"`
	RepoURL      string     `json:"repoUrl,omitempty"`
	Remote       string     `json:"remote,omitempty"`
	Provider     string     `json:"provider,omitempty"`
	MergeRef     string     `json:"mergeRef,omitempty"`
	PRTitle      string     `json:"prTitle,omitempty"`
	WorkItems    []WorkItem `json:"workItems,omitempty"`
	WorkItemNote string     `json:"workItemNote,omitempty"`
	Range        string     `json:"range"`
	Files        []string   `json:"files"`
	Stat         string     `json:"stat"`
	Patch        string     `json:"patch"`
	ChangedFiles int        `json:"changedFiles"`
	PatchBytes   int        `json:"patchBytes"`
}
