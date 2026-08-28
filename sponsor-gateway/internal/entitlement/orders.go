package entitlement

type Order struct {
	ActualPaidFen int64
	Status        string
}

// NetPaidFen totals orders in their final provider state. A refunded, revoked,
// or cancelled order contributes zero; it is not a separate negative ledger row.
func NetPaidFen(orders []Order) int64 {
	var total int64
	for _, order := range orders {
		switch order.Status {
		case "paid", "success":
			total += order.ActualPaidFen
		}
	}
	if total < 0 {
		return 0
	}
	return total
}
