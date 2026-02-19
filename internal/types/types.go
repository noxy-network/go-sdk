package types

// NoxyIdentityAddress is an EVM wallet address in 0x format.
type NoxyIdentityAddress = string

// NoxyPushDeliveryStatus represents the delivery status of a push notification.
type NoxyPushDeliveryStatus int32

const (
	NoxyPushDeliveryStatusDelivered  NoxyPushDeliveryStatus = 0
	NoxyPushDeliveryStatusQueued    NoxyPushDeliveryStatus = 1
	NoxyPushDeliveryStatusNoDevices NoxyPushDeliveryStatus = 2
	NoxyPushDeliveryStatusRejected  NoxyPushDeliveryStatus = 3
	NoxyPushDeliveryStatusError     NoxyPushDeliveryStatus = 4
)

// NoxyPushResponse is the response for a push notification send.
type NoxyPushResponse struct {
	Status    NoxyPushDeliveryStatus
	RequestID string
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
	RequestID       string
	AppName         string
	QuotaTotal      uint64
	QuotaRemaining  uint64
	Status          NoxyQuotaStatus
}

// NoxyIdentityDevice represents a device with keys for encryption.
type NoxyIdentityDevice struct {
	DeviceID    string
	PublicKey   []byte
	PQPublicKey []byte
}
