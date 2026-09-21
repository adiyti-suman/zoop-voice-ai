package verification

type VerificationStatus string

const (
	StatusCreated                 VerificationStatus = "CREATED"
	StatusCollecting              VerificationStatus = "COLLECTING"
	StatusProcessing              VerificationStatus = "PROCESSING"
	StatusEvaluating              VerificationStatus = "EVALUATING"
	StatusReviewRequired          VerificationStatus = "REVIEW_REQUIRED"
	StatusReviewing               VerificationStatus = "REVIEWING"
	StatusMoreInformationRequired VerificationStatus = "MORE_INFORMATION_REQUIRED"
	StatusApproved                VerificationStatus = "APPROVED"
	StatusRejected                VerificationStatus = "REJECTED"
	StatusFailed                  VerificationStatus = "FAILED"
	StatusCancelled               VerificationStatus = "CANCELLED"
	StatusExpired                 VerificationStatus = "EXPIRED"
	StatusCompleted               VerificationStatus = "COMPLETED"
)

func (s VerificationStatus) String() string {
	return string(s)
}

func IsTerminalState(status VerificationStatus) bool {
	switch status {
	case StatusCompleted, StatusFailed, StatusCancelled, StatusExpired:
		return true
	}
	return false
}

func IsValidTransition(current, next VerificationStatus) bool {
	if current == next {
		return true
	}
	
	switch current {
	case StatusCreated:
		return next == StatusCollecting || next == StatusCancelled || next == StatusExpired
	case StatusCollecting:
		return next == StatusProcessing || next == StatusCancelled || next == StatusExpired
	case StatusProcessing:
		return next == StatusEvaluating || next == StatusFailed || next == StatusCancelled
	case StatusEvaluating:
		return next == StatusApproved || next == StatusReviewRequired || next == StatusMoreInformationRequired || next == StatusRejected || next == StatusFailed
	case StatusReviewRequired:
		return next == StatusReviewing || next == StatusCancelled
	case StatusReviewing:
		return next == StatusApproved || next == StatusRejected || next == StatusMoreInformationRequired
	case StatusMoreInformationRequired:
		return next == StatusCollecting || next == StatusProcessing || next == StatusCancelled || next == StatusExpired
	}
	return false
}
