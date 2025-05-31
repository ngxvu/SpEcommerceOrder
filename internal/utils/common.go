package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/validation"
	"github.com/ericlagergren/decimal"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/volatiletech/sqlboiler/v4/types"
	"golang.org/x/crypto/bcrypt"
)

func ToPointer[T any](in T) *T {
	return &in
}

func Float64ToNullDecimal(value *float64) types.NullDecimal {
	if value == nil {
		return types.NullDecimal{}
	}
	// Convert float64 to *decimal.Big
	d := new(decimal.Big).SetFloat64(*value)
	return types.NewNullDecimal(d)
}

func CheckRequireValid(ob interface{}) error {
	validator := validation.Validation{RequiredFirst: true}
	passed, err := validator.Valid(ob)
	if err != nil {
		return err
	}
	if !passed {
		var err string
		for _, e := range validator.Errors {
			err += fmt.Sprintf("[%s: %s] ", e.Field, e.Message)
		}
		return fmt.Errorf(err)
	}
	return nil
}

func HashWithSHA256(input string) string {
	hashed := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hashed[:])
}

// CheckPasswordHash compares a bcrypt hashed password with its possible plaintext equivalent.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// HashPassword hashes a plaintext password using bcrypt.
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func UUIDtoString(uuid uuid.UUID) string {
	return uuid.String()
}

func ContainsString(value string, allowedValues []string) bool {
	for _, allowed := range allowedValues {
		if value == allowed {
			return true
		}
	}
	return false
}

func TransferDataToJsonB(data []*string) (*postgres.Jsonb, error) {

	jsonb, err := toJsonb(data)
	if err != nil {

		return nil, err
	}

	return jsonb, nil
}

func toJsonb(data interface{}) (*postgres.Jsonb, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	var jsonbData postgres.Jsonb
	if err := jsonbData.UnmarshalJSON(jsonData); err != nil {
		return nil, err
	}
	return &jsonbData, nil
}

func ValidOperators() []string {
	return []string{
		"contains", "not_contains", "equals", "not_equals",
		"starts_with", "ends_with", "is_empty", "is_not_empty",
		"is_any_of", "greater_than", "less_than",
		"greater_than_or_equal", "less_than_or_equal",
	}
}
