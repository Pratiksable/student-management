package models

type StudentRisk struct {
	StudentID            int      `json:"student_id"`
	RiskScore            int      `json:"risk_score"`
	RiskLevel            string   `json:"risk_level"`
	AttendancePercentage float64  `json:"attendance_percentage"`
	AverageMarks         float64  `json:"average_marks"`
	MissingAssignments   int      `json:"missing_assignments"`
	Reasons              []string `json:"reasons"`
	PerformanceTrend     string   `json:"performance_trend"`
}
