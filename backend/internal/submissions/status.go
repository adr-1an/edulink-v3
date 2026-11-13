// Package submissions contains constants for assignment submission status
package submissions

type SubmissionStatus string

const (
	SubmissionStatusPending   SubmissionStatus = "pending"
	SubmissionStatusSubmitted SubmissionStatus = "submitted"
	SubmissionStatusReturned  SubmissionStatus = "returned"
)
