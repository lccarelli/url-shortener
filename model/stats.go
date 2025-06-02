package model

type GeneralStatsResponse struct {
	TotalShortened int               `json:"total_shortened"`
	TotalVisits    int               `json:"total_visits"`
	Created        int               `json:"created"`
	Deleted        int               `json:"deleted"`
	Resolved       int               `json:"resolved"`
	TopAccessed    []ShortStatsEntry `json:"top_accessed"`
	RecentAccesses []ShortStatsEntry `json:"recent_accesses"`
}

type ShortStatsEntry struct {
	ShortCode  string `json:"short_code"`
	Visits     int    `json:"visits"`
	LastAccess string `json:"last_access,omitempty"`
}
