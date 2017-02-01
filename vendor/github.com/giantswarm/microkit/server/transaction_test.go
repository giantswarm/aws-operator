package server

import (
	"testing"
)

func Test_Transaction_IDFormat(t *testing.T) {
	testCases := []struct {
		TransactionID string
		Valid         bool
	}{
		{
			TransactionID: "",
			Valid:         false,
		},
		{
			TransactionID: "foo",
			Valid:         false,
		},
		{
			TransactionID: "d.99ab4af-ddc7-4c7b-8e2b-1cdef5b129c7",
			Valid:         false,
		},
		{
			TransactionID: "-99ab4af-ddc7-4c7b-8e2b-1cdef5b129c7",
			Valid:         false,
		},
		{
			TransactionID: "d99ab4af-ddc7-4c7b-8e2b-1cdef5b129c-",
			Valid:         false,
		},
		{
			TransactionID: "d99ab4af-ddc7-4c7b-8e2b-1cdef5b129c7",
			Valid:         true,
		},
		{
			TransactionID: "a1e0d43b-fea2-4240-84a7-7abdffca1999",
			Valid:         true,
		},
		{
			TransactionID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
			Valid:         true,
		},
		{
			TransactionID: "00000000-0000-0000-0000-000000000000",
			Valid:         true,
		},
	}

	for _, testCase := range testCases {
		isValid := IsValidTransactionID(testCase.TransactionID)
		if isValid != testCase.Valid {
			t.Fatal("expected", testCase.Valid, "got", isValid)
		}
	}
}
