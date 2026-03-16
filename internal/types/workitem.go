package types

// WorkItem represents a linked work item, issue, or ticket associated with the PR.
type WorkItem struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	State string `json:"state,omitempty"`
	URL   string `json:"url,omitempty"`
}
