package types

type DiffOutput struct {
	PRID     int      `json:"prId,omitempty"`
	RepoURL  string   `json:"repoUrl,omitempty"`
	Remote   string   `json:"remote,omitempty"`
	Provider string   `json:"provider,omitempty"`
	BareDir  string   `json:"bareDir,omitempty"`
	MergeRef string   `json:"mergeRef,omitempty"`
	WorkDir  string   `json:"workDir,omitempty"`
	Range    string   `json:"range"`
	Files    []string `json:"files"`
	Stat     string   `json:"stat"`
	Patch    string   `json:"patch"`
}
