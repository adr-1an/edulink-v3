package submissions

type SubmissionStatus string

const (
	SubmissionStatusPending   SubmissionStatus = "pending"
	SubmissionStatusSubmitted SubmissionStatus = "submitted"
	SubmissionStatusReturned  SubmissionStatus = "returned"
)
