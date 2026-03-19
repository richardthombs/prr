package types

type RelatedIssue struct {
	ID       string            `json:"id"`
	Type     string            `json:"type,omitempty"`
	Provider string            `json:"provider,omitempty"`
	URL      string            `json:"url"`
	Title    string            `json:"title"`
	Body     string            `json:"body,omitempty"`
	State    string            `json:"state,omitempty"`
	Labels   []string          `json:"labels,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}
