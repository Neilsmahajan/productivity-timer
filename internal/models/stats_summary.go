package models

// StatsSummary represents aggregated stats for a time period
type StatsSummary struct {
	TotalDuration  int64      `json:"totalDuration"`  // Total seconds across all tags
	TotalSessions  int        `json:"totalSessions"`  // Total number of sessions
	AverageSession int64      `json:"averageSession"` // Average session duration in seconds
	MostUsedTag    string     `json:"mostUsedTag"`    // Tag with most time spent
	TagBreakdown   []TagStats `json:"tagBreakdown"`   // Per-tag breakdown
}

// TagStats represents stats for a single tag within a time period
type TagStats struct {
	Tag               string  `bson:"_id" json:"tag"`
	TotalDuration     int64   `bson:"total_duration" json:"totalDuration"` // Total seconds for this tag
	SessionCount      int     `bson:"session_count" json:"sessionCount"`   // Number of sessions
	AverageSession    int64   `json:"averageSession"`                      // Average session duration
	PercentageOfTotal float64 `json:"percentageOfTotal"`                   // Percentage of total time
}
