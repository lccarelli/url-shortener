package model

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Short string `json:"short"`
	Url   string `json:"url"`
}
