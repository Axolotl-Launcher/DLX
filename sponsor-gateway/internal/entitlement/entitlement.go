package entitlement

const DefaultThresholdFen int64 = 990

func Eligible(netPaidFen, thresholdFen int64) bool { return netPaidFen >= thresholdFen }

func RemainingFen(netPaidFen, thresholdFen int64) int64 {
	if netPaidFen >= thresholdFen {
		return 0
	}
	return thresholdFen - netPaidFen
}
