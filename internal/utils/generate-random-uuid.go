package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func GenerateRandomUUID() string {
	return uuid.New().String()
}

// GenerateEsewaTransactionUUID generates a transaction UUID that complies with eSewa requirements
// eSewa supports alphanumeric characters and hyphen(-) only
// Format: YYMMDD-HHMMSS-XXXXX (where XXXXX is a random 5-character string)
func GenerateEsewaTransactionUUID() string {
	now := time.Now()
	dateStr := now.Format("060102") // YYMMDD
	timeStr := now.Format("150405") // HHMMSS

	// Generate random 5-character string (alphanumeric only)
	randomBytes := make([]byte, 3)
	rand.Read(randomBytes)
	randomStr := fmt.Sprintf("%05X", randomBytes)[:5]

	return fmt.Sprintf("%s-%s-%s", dateStr, timeStr, randomStr)
}

func GenerateEsewaSignature(
	totalAmount string,
	transactionUuid string,
	productCode string,
	secretKey string,
) string {
	// According to the blog post and eSewa documentation, the signature should be generated from:
	// "total_amount=value,transaction_uuid=value,product_code=value"
	signatureString := fmt.Sprintf("total_amount=%s,transaction_uuid=%s,product_code=%s",
		totalAmount, transactionUuid, productCode)

	if secretKey == "" {
		return "ESEWA_SECRET_KEY is required"
	}

	// Debug: Print the data being signed (remove in production)
	fmt.Printf("DEBUG: Signature string: '%s'\n", signatureString)
	fmt.Printf("DEBUG: Secret key length: %d\n", len(secretKey))

	hash := hmac.New(sha256.New, []byte(secretKey))
	hash.Write([]byte(signatureString))
	signature := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	// Debug: Print the generated signature (remove in production)
	fmt.Printf("DEBUG: Generated signature: %s\n", signature)

	return signature
}

// Alternative signature generation method - try this if the above doesn't work
func GenerateEsewaSignatureAlternative(
	totalAmount string,
	transactionUuid string,
	productCode string,
	secretKey string,
) string {
	// Some implementations use a different format
	// Try concatenating with commas but without field names
	data := totalAmount + "," + transactionUuid + "," + productCode

	if secretKey == "" {
		return "ESEWA_SECRET_KEY is required"
	}

	hash := hmac.New(sha256.New, []byte(secretKey))
	hash.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}
