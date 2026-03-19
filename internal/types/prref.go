package types

type PRRef struct {
	PRID       int    `json:"prId"`
	RepoURL    string `json:"repoUrl"`
	Remote     string `json:"remote"`
	Provider   string `json:"provider"`
	BaseBranch string `json:"baseBranch,omitempty"`
	BaseSHA    string `json:"baseSha,omitempty"`
	PRURL      string `json:"prUrl,omitempty"`
}