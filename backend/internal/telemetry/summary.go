package telemetry

import "time"

type SummaryBucket struct {
	BucketStart     time.Time
	Average         *float64
	Minimum         *float64
	Maximum         *float64
	ValidCount      int64
	OutOfRangeCount int64
}
