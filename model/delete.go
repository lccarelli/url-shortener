package model

type DeleteRequest struct {
	Keys []string `json:"keys" example:"[\"abc123\", \"def456\"]"`
}

type DeleteResponse struct {
	Deleted  []string `json:"deleted"`
	NotFound []string `json:"not_found,omitempty"`
}
