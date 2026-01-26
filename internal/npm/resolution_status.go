package npm

type ResolutionStatus string

const (
	StatusResolved ResolutionStatus = "resolved"
	StatusCyclic   ResolutionStatus = "cyclic"
	StatusFailed   ResolutionStatus = "failed"
)
