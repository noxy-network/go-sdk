package types

// NoxyIdentityAddress is an EVM wallet address in 0x format.
type NoxyIdentityAddress = string

// NoxyDeliveryStatus is relay-side delivery status after RouteDecision (matches proto DeliveryStatus).
type NoxyDeliveryStatus int32

const (
	NoxyDeliveryStatusDelivered NoxyDeliveryStatus = 0
	NoxyDeliveryStatusQueued    NoxyDeliveryStatus = 1
	NoxyDeliveryStatusNoDevices NoxyDeliveryStatus = 2
	NoxyDeliveryStatusRejected  NoxyDeliveryStatus = 3
	NoxyDeliveryStatusError     NoxyDeliveryStatus = 4
)

// NoxyDeliveryOutcome is the result of RouteDecision for one device.
type NoxyDeliveryOutcome struct {
	Status     NoxyDeliveryStatus
	RequestID  string
	DecisionID string
}

// NoxyHumanDecisionOutcome is human-in-the-loop resolution (matches proto DecisionOutcome).
type NoxyHumanDecisionOutcome int32

const (
	NoxyHumanDecisionOutcomePending  NoxyHumanDecisionOutcome = 0
	NoxyHumanDecisionOutcomeApproved NoxyHumanDecisionOutcome = 1
	NoxyHumanDecisionOutcomeRejected NoxyHumanDecisionOutcome = 2
	NoxyHumanDecisionOutcomeExpired  NoxyHumanDecisionOutcome = 3
)

// NoxyGetDecisionOutcomeResponse is a single poll of GetDecisionOutcome.
type NoxyGetDecisionOutcomeResponse struct {
	RequestID string
	Pending   bool
	Outcome   NoxyHumanDecisionOutcome
}

// NoxyQuotaStatus represents the quota status for the application.
type NoxyQuotaStatus int32

const (
	NoxyQuotaStatusActive    NoxyQuotaStatus = 0
	NoxyQuotaStatusSuspended NoxyQuotaStatus = 1
	NoxyQuotaStatusDeleted   NoxyQuotaStatus = 2
)

// NoxyGetQuotaResponse is the response for a quota query.
type NoxyGetQuotaResponse struct {
	RequestID      string
	AppName        string
	QuotaTotal     uint64
	QuotaRemaining uint64
	Status         NoxyQuotaStatus
}

// NoxyIdentityDevice represents a device with keys for encryption.
type NoxyIdentityDevice struct {
	DeviceID    string
	PublicKey   []byte
	PQPublicKey []byte
}
