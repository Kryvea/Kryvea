package mongo

import (
	"github.com/google/uuid"
)

func IsNullCustomer(customer *Customer) bool {
	if customer == nil {
		return true
	}

	if customer.ID == uuid.Nil {
		return true
	}

	return false
}
