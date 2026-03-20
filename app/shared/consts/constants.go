package consts

const (
	StatusActive   = 1
	StatusDisabled = 0
)

const (
	VideoStatusDraft     int8 = 0
	VideoStatusPublished int8 = 1
	VideoStatusOffline   int8 = 2
)

const (
	TranscodePending    int8 = 0
	TranscodeProcessing int8 = 1
	TranscodeSuccess    int8 = 2
	TranscodeFailed     int8 = 3
)
