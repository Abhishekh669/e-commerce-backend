package main

import (
	"fmt"
	"os"

	"e-commerce.com/internal/utils"
)

func main() {
	// Test data from eSewa documentation
	testCases := []struct {
		totalAmount     string
		transactionUUID string
		productCode     string
		secretKey       string
		expectedResult  string
	}{
		{
			totalAmount:     "100",
			transactionUUID: "241028",
			productCode:     "EPAYTEST",
			secretKey:       "8gBm/:&EnhH.1/q(",
			expectedResult:  "i94zsd3oXF6ZsSr/kGqT4sSzYQzjj1W/waxjWyRwaME=",
		},
		{
			totalAmount:     "110",
			transactionUUID: "241028",
			productCode:     "EPAYTEST",
			secretKey:       "8gBm/:&EnhH.1/q(",
			expectedResult:  "",
		},
	}

	fmt.Println("Testing eSewa Signature Generation")
	fmt.Println("==================================")

	for i, testCase := range testCases {
		fmt.Printf("\nTest Case %d:\n", i+1)
		fmt.Printf("Total Amount: %s\n", testCase.totalAmount)
		fmt.Printf("Transaction UUID: %s\n", testCase.transactionUUID)
		fmt.Printf("Product Code: %s\n", testCase.productCode)
		fmt.Printf("Secret Key: %s\n", testCase.secretKey)

		// Test main signature function
		signature1 := utils.GenerateEsewaSignature(
			testCase.totalAmount,
			testCase.transactionUUID,
			testCase.productCode,
			testCase.secretKey,
		)
		fmt.Printf("Signature 1 (concatenated): %s\n", signature1)

		// Test alternative signature function
		signature2 := utils.GenerateEsewaSignatureAlternative(
			testCase.totalAmount,
			testCase.transactionUUID,
			testCase.productCode,
			testCase.secretKey,
		)
		fmt.Printf("Signature 2 (comma-separated): %s\n", signature2)

		// Test with our new transaction UUID generator
		newUUID := utils.GenerateEsewaTransactionUUID()
		fmt.Printf("Generated eSewa UUID: %s\n", newUUID)

		// Test signature with new UUID
		signature3 := utils.GenerateEsewaSignature(
			testCase.totalAmount,
			newUUID,
			testCase.productCode,
			testCase.secretKey,
		)
		fmt.Printf("Signature 3 (with new UUID): %s\n", signature3)
	}

	// Test environment variables
	fmt.Println("\nEnvironment Variables:")
	fmt.Printf("ESEWA_MERCHANT_CODE: %s\n", os.Getenv("ESEWA_MERCHANT_CODE"))
	fmt.Printf("ESEWA_SECRET_KEY: %s\n", os.Getenv("ESEWA_SECRET_KEY"))
	fmt.Printf("ESEWA_PAYMENT_URL: %s\n", os.Getenv("ESEWA_PAYMENT_URL"))
}
