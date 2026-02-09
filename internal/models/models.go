package models

import "time"

// Derive represents a derive record
type Derive struct {
	ID           int    `json:"id"`
	Number       int    `json:"number"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	ImageUrl     string `json:"image_url"`
	ImageLqip    string `json:"image_lqip"`
	ContribCount int    `json:"contrib_count"`
	// Points assigned to the derive (used for badges and overlay selection)
	Points int `json:"points"`
	// PointsTier maps points to 1..3 for styling purposes
	PointsTier int `json:"points_tier"`
}

// Contribution represents a user contribution
type Contribution struct {
	ImageUrl    string
	ImageLqip   string
	UserName    string
	UserCity    string
	UserComment string
	CreatedAt   time.Time
}

// FooterStats holds database statistics for the footer
type FooterStats struct {
	TotalDeriven       int
	TotalContributions int
	ActiveUsers        int
	TotalCities        int
	LastActivity       time.Time
}

// TokenInfo holds information about an upload token
type TokenInfo struct {
	ID                int       `json:"id"`
	Token             string    `json:"token"`
	BagName           string    `json:"bag_name"`
	CurrentPlayer     string    `json:"current_player"`
	CurrentPlayerCity string    `json:"current_player_city"`
	IsActive          bool      `json:"is_active"`
	MaxUploads        int       `json:"max_uploads"`
	TotalUploads      int       `json:"total_uploads"`
	TotalSessions     int       `json:"total_sessions"`
	SessionStartedAt  time.Time `json:"session_started_at"`
	CreatedAt         time.Time `json:"created_at"`
	Remaining         int       `json:"remaining"`
}

// RecentContrib represents a recent contribution for the admin dashboard
type RecentContrib struct {
	ID           int
	ImageUrl     string
	PlayerName   string
	DeriveNumber int
}

// BagRequest represents a bag request
type BagRequest struct {
	ID        int
	Email     string
	CreatedAt time.Time
	Handled   bool
}

// PageNumber represents pagination information
type PageNumber struct {
	Number    int
	IsCurrent bool
	IsDots    bool
}
