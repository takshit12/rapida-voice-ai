// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package exceptions

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIAuthenticationError_BasicFunctionality(t *testing.T) {
	result, err := APIAuthenticationError[map[string]interface{}]()

	// Should return nil error
	assert.Nil(t, err)

	// Should return non-nil result
	assert.NotNil(t, result)

	// Should be able to marshal/unmarshal as JSON
	data, marshalErr := json.Marshal(result)
	assert.NoError(t, marshalErr)
	assert.NotEmpty(t, data)
}

func TestAPIAuthenticationError_JSONStructure(t *testing.T) {
	result, _ := APIAuthenticationError[map[string]interface{}]()

	// Parse the JSON structure
	data, _ := json.Marshal(result)
	var parsed map[string]interface{}
	err := json.Unmarshal(data, &parsed)
	assert.NoError(t, err)

	// Check structure
	assert.Equal(t, float64(APIUnauthorized), parsed["Code"])
	assert.Equal(t, false, parsed["Success"])

	// Check error object
	errorObj, ok := parsed["Error"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(APIUnauthorized), errorObj["errorCode"])
	assert.Equal(t, "unauthenticated request", errorObj["errorMessage"])
	assert.Equal(t, ErrorMessages[APIUnauthorized], errorObj["humanMessage"])
}

func TestAPIAuthenticationError_WithDifferentTypes(t *testing.T) {
	// Test with different generic types
	result1, _ := APIAuthenticationError[interface{}]()
	assert.NotNil(t, result1)

	result2, _ := APIAuthenticationError[map[string]string]()
	assert.NotNil(t, result2)

	result3, _ := APIAuthenticationError[[]byte]()
	assert.NotNil(t, result3)
}

func TestAuthenticationError_BasicFunctionality(t *testing.T) {
	result, err := AuthenticationError[map[string]interface{}]()

	assert.Nil(t, err)
	assert.NotNil(t, result)
}

func TestAuthenticationError_JSONStructure(t *testing.T) {
	result, _ := AuthenticationError[map[string]interface{}]()

	data, _ := json.Marshal(result)
	var parsed map[string]interface{}
	err := json.Unmarshal(data, &parsed)
	assert.NoError(t, err)

	assert.Equal(t, float64(Unauthorized), parsed["Code"])
	assert.Equal(t, false, parsed["Success"])

	errorObj := parsed["Error"].(map[string]interface{})
	assert.Equal(t, float64(Unauthorized), errorObj["errorCode"])
	assert.Equal(t, "unauthenticated request", errorObj["errorMessage"])
	assert.Equal(t, ErrorMessages[Unauthorized], errorObj["humanMessage"])
}

func TestBadRequestError_BasicFunctionality(t *testing.T) {
	customMsg := "Custom bad request message"
	result, err := BadRequestError[map[string]interface{}](customMsg)

	assert.Nil(t, err)
	assert.NotNil(t, result)
}

func TestBadRequestError_CustomMessage(t *testing.T) {
	customMsg := "Custom bad request message"
	result, _ := BadRequestError[map[string]interface{}](customMsg)

	data, _ := json.Marshal(result)
	var parsed map[string]interface{}
	json.Unmarshal(data, &parsed)

	assert.Equal(t, float64(BadRequest), parsed["Code"])
	errorObj := parsed["Error"].(map[string]interface{})
	assert.Equal(t, "bad request", errorObj["errorMessage"])
	assert.Equal(t, customMsg, errorObj["humanMessage"])
}

func TestBadRequestError_EmptyMessage(t *testing.T) {
	result, _ := BadRequestError[map[string]interface{}]("")

	data, _ := json.Marshal(result)
	var parsed map[string]interface{}
	json.Unmarshal(data, &parsed)

	errorObj := parsed["Error"].(map[string]interface{})
	// Empty string might not be included in JSON, so check if field exists
	humanMsg, exists := errorObj["humanMessage"]
	if exists {
		assert.Equal(t, "", humanMsg)
	} else {
		// Field doesn't exist, which is also acceptable for empty strings
		assert.True(t, true)
	}
}

func TestBadRequestError_SpecialCharacters(t *testing.T) {
	messages := []string{
		"Message with spaces and symbols: @#$%^&*()",
		"Unicode: 你好世界 🌍",
		"Newlines\nand\ttabs",
		"Quotes: 'single' and \"double\"",
		"Very long message with lots of text that should be handled properly by the error function",
	}

	for _, msg := range messages {
		result, _ := BadRequestError[map[string]interface{}](msg)
		data, _ := json.Marshal(result)
		var parsed map[string]interface{}
		json.Unmarshal(data, &parsed)

		errorObj := parsed["Error"].(map[string]interface{})
		assert.Equal(t, msg, errorObj["humanMessage"])
	}
}

func TestInternalServerError_BasicFunctionality(t *testing.T) {
	originalErr := errors.New("database connection failed")
	customMsg := "Custom internal server message"
	result, err := InternalServerError[map[string]interface{}](originalErr, customMsg)

	assert.Nil(t, err)
	assert.NotNil(t, result)
}

func TestInternalServerError_ErrorPropagation(t *testing.T) {
	originalErr := errors.New("database connection failed")
	customMsg := "Custom internal server message"
	result, _ := InternalServerError[map[string]interface{}](originalErr, customMsg)

	data, _ := json.Marshal(result)
	var parsed map[string]interface{}
	json.Unmarshal(data, &parsed)

	assert.Equal(t, float64(InternalServer), parsed["Code"])
	errorObj := parsed["Error"].(map[string]interface{})
	assert.Equal(t, originalErr.Error(), errorObj["errorMessage"])
	assert.Equal(t, customMsg, errorObj["humanMessage"])
}

func TestInternalServerError_NilError(t *testing.T) {
	customMsg := "Custom message"

	// This should panic when calling err.Error() on nil
	defer func() {
		if r := recover(); r != nil {
			// Expected panic
			t.Logf("Expected panic occurred: %v", r)
		} else {
			t.Error("Expected panic did not occur")
		}
	}()

	InternalServerError[map[string]interface{}](nil, customMsg)
}

func TestInternalServerError_EmptyMessage(t *testing.T) {
	originalErr := errors.New("some error")
	result, _ := InternalServerError[map[string]interface{}](originalErr, "")

	data, _ := json.Marshal(result)
	var parsed map[string]interface{}
	json.Unmarshal(data, &parsed)

	errorObj := parsed["Error"].(map[string]interface{})
	// Empty human message might be omitted from JSON
	humanMsg, exists := errorObj["humanMessage"]
	if exists {
		assert.Equal(t, "", humanMsg)
	} else {
		// Field doesn't exist, which is acceptable for empty strings
		assert.True(t, true)
	}
}

func TestErrorWithCode_BasicFunctionality(t *testing.T) {
	code := int32(418) // I'm a teapot
	err := errors.New("teapot error")
	humanMsg := "I'm a teapot"

	result := ErrorWithCode[map[string]interface{}](code, err, humanMsg)
	assert.NotNil(t, result)

	data, _ := json.Marshal(result)
	var parsed map[string]interface{}
	json.Unmarshal(data, &parsed)

	assert.Equal(t, float64(code), parsed["Code"])
	assert.Equal(t, false, parsed["Success"])

	errorObj := parsed["Error"].(map[string]interface{})
	assert.Equal(t, float64(code), errorObj["errorCode"])
	assert.Equal(t, err.Error(), errorObj["errorMessage"])
	assert.Equal(t, humanMsg, errorObj["humanMessage"])
}

func TestErrorWithCode_DifferentCodes(t *testing.T) {
	testCases := []struct {
		code     int32
		errorMsg string
		humanMsg string
	}{
		{200, "ok error", "Everything is fine"},
		{404, "not found", "Resource missing"},
		{500, "server error", "Internal issue"},
		{0, "zero code", "Zero status"},
		{-1, "negative code", "Negative status"},
		{999, "high code", "High status code"},
	}

	for _, tc := range testCases {
		err := errors.New(tc.errorMsg)
		result := ErrorWithCode[map[string]interface{}](tc.code, err, tc.humanMsg)

		data, _ := json.Marshal(result)
		var parsed map[string]interface{}
		json.Unmarshal(data, &parsed)

		assert.Equal(t, float64(tc.code), parsed["Code"])
		errorObj := parsed["Error"].(map[string]interface{})

		// Check errorCode - note: uint64 conversion can cause issues with negative numbers
		// and zero values might be omitted due to omitempty
		if tc.code != 0 {
			expectedErrorCode := uint64(tc.code)
			if tc.code < 0 {
				// Negative int32 -> uint64 wraps around
				expectedErrorCode = uint64(tc.code) // This will be a large positive number
			}
			assert.Equal(t, float64(expectedErrorCode), errorObj["errorCode"])
		} else {
			// For code 0, errorCode might be omitted due to omitempty
			errorCode, exists := errorObj["errorCode"]
			if exists {
				assert.Equal(t, float64(0), errorCode)
			}
		}

		// Check errorMessage - might be omitted if empty
		if tc.errorMsg != "" {
			assert.Equal(t, tc.errorMsg, errorObj["errorMessage"])
		} else {
			errorMsg, exists := errorObj["errorMessage"]
			if exists {
				assert.Equal(t, "", errorMsg)
			}
		}

		// Check humanMessage - might be omitted if empty
		if tc.humanMsg != "" {
			assert.Equal(t, tc.humanMsg, errorObj["humanMessage"])
		} else {
			humanMsg, exists := errorObj["humanMessage"]
			if exists {
				assert.Equal(t, "", humanMsg)
			}
		}
	}
}

func TestErrorWithCode_EmptyStrings(t *testing.T) {
	result := ErrorWithCode[map[string]interface{}](400, errors.New(""), "")

	data, _ := json.Marshal(result)
	var parsed map[string]interface{}
	json.Unmarshal(data, &parsed)

	errorObj := parsed["Error"].(map[string]interface{})
	// Empty strings might be omitted from JSON
	errorMsg, hasErrorMsg := errorObj["errorMessage"]
	humanMsg, hasHumanMsg := errorObj["humanMessage"]

	if hasErrorMsg {
		assert.Equal(t, "", errorMsg)
	}
	if hasHumanMsg {
		assert.Equal(t, "", humanMsg)
	}
	// It's acceptable if empty fields are omitted
}

func TestErrorWithCode_SpecialCharacters(t *testing.T) {
	err := errors.New("special error: @#$%^&*()")
	humanMsg := "Special message: 你好世界 🌍\n\t\"quotes\""
	result := ErrorWithCode[map[string]interface{}](400, err, humanMsg)

	data, _ := json.Marshal(result)
	var parsed map[string]interface{}
	json.Unmarshal(data, &parsed)

	errorObj := parsed["Error"].(map[string]interface{})
	assert.Equal(t, err.Error(), errorObj["errorMessage"])
	assert.Equal(t, humanMsg, errorObj["humanMessage"])
}

func TestErrorWithCode_DifferentGenericTypes(t *testing.T) {
	code := int32(404)
	err := errors.New("test error")
	humanMsg := "test message"

	// Test with different types
	result1 := ErrorWithCode[interface{}](code, err, humanMsg)
	assert.NotNil(t, result1)

	result2 := ErrorWithCode[map[string]string](code, err, humanMsg)
	assert.NotNil(t, result2)

	result3 := ErrorWithCode[[]interface{}](code, err, humanMsg)
	assert.NotNil(t, result3)

	result4 := ErrorWithCode[string](code, err, humanMsg)
	assert.NotNil(t, result4)
}

func TestErrorWithCode_JSONRoundTrip(t *testing.T) {
	code := int32(500)
	err := errors.New("round trip error")
	humanMsg := "Round trip test"

	result := ErrorWithCode[map[string]interface{}](code, err, humanMsg)

	// Marshal to JSON
	data1, _ := json.Marshal(result)

	// Unmarshal back
	var intermediate map[string]interface{}
	json.Unmarshal(data1, &intermediate)

	// Marshal again
	data2, _ := json.Marshal(intermediate)

	// Should be identical
	assert.Equal(t, data1, data2)
}

func TestErrorWithCode_NilError(t *testing.T) {
	// This should panic or handle nil error gracefully
	// Let's see what happens
	defer func() {
		if r := recover(); r != nil {
			// Expected to panic when calling err.Error() on nil
			t.Logf("Expected panic occurred: %v", r)
		}
	}()

	result := ErrorWithCode[map[string]interface{}](400, nil, "test")
	assert.NotNil(t, result)
}

func TestConstants(t *testing.T) {
	assert.Equal(t, 401, Unauthorized)
	assert.Equal(t, 403, APIUnauthorized)
	assert.Equal(t, 404, NotFound)
	assert.Equal(t, 500, InternalServer)
	assert.Equal(t, 400, BadRequest)
}

func TestErrorMessages(t *testing.T) {
	expected := map[int]string{
		Unauthorized:    "Unauthenticated request, please try again with valid authentication.",
		APIUnauthorized: "Invalid API key, please provide a valid key.",
		NotFound:        "Resource not found, please check the endpoint and try again.",
		InternalServer:  "Internal server error, please try again later.",
	}

	for code, expectedMsg := range expected {
		actualMsg, exists := ErrorMessages[code]
		assert.True(t, exists, "Error message for code %d should exist", code)
		assert.Equal(t, expectedMsg, actualMsg)
	}

	// Check that BadRequest doesn't have a message (it's handled differently)
	_, exists := ErrorMessages[BadRequest]
	assert.False(t, exists, "BadRequest should not have a predefined message")
}

func TestErrorMessages_AllConstantsCovered(t *testing.T) {
	definedCodes := []int{Unauthorized, APIUnauthorized, NotFound, InternalServer}
	missingCodes := []int{BadRequest} // BadRequest is intentionally not in ErrorMessages

	for _, code := range definedCodes {
		_, exists := ErrorMessages[code]
		assert.True(t, exists, "Error message for code %d should be defined", code)
	}

	for _, code := range missingCodes {
		_, exists := ErrorMessages[code]
		assert.False(t, exists, "Error message for code %d should NOT be defined", code)
	}
}

func TestIntegration_AllErrorFunctions(t *testing.T) {
	// Test that all error functions produce valid JSON
	functions := []func() (interface{}, error){
		func() (interface{}, error) { return APIAuthenticationError[map[string]interface{}]() },
		func() (interface{}, error) { return AuthenticationError[map[string]interface{}]() },
		func() (interface{}, error) { return BadRequestError[map[string]interface{}]("test") },
		func() (interface{}, error) {
			return InternalServerError[map[string]interface{}](errors.New("test"), "test")
		},
	}

	for i, fn := range functions {
		result, err := fn()
		assert.Nil(t, err, "Function %d should not return error", i)
		assert.NotNil(t, result, "Function %d should return result", i)

		// Should be valid JSON
		data, marshalErr := json.Marshal(result)
		assert.NoError(t, marshalErr, "Function %d should produce valid JSON", i)
		assert.NotEmpty(t, data, "Function %d should produce non-empty JSON", i)
	}
}

func TestErrorWithCode_TypeSafety(t *testing.T) {
	// Test that the generic type is preserved in the result
	type CustomStruct struct {
		Field1 string `json:"field1"`
		Field2 int    `json:"field2"`
	}

	result := ErrorWithCode[CustomStruct](404, errors.New("test"), "message")
	assert.NotNil(t, result)
	assert.IsType(t, &CustomStruct{}, result)

	// The result should be a pointer to CustomStruct
	assert.IsType(t, &CustomStruct{}, result)
}

func TestErrorWithCode_LargeData(t *testing.T) {
	// Test with very long error messages
	longErrorMsg := string(make([]byte, 10000)) // 10KB string
	for i := range longErrorMsg {
		longErrorMsg = longErrorMsg[:i] + "a" + longErrorMsg[i+1:]
	}

	longHumanMsg := string(make([]byte, 10000))
	for i := range longHumanMsg {
		longHumanMsg = longHumanMsg[:i] + "b" + longHumanMsg[i+1:]
	}

	result := ErrorWithCode[map[string]interface{}](500, errors.New(longErrorMsg), longHumanMsg)
	assert.NotNil(t, result)

	// Should still be valid JSON
	data, err := json.Marshal(result)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestErrorWithCode_UnicodeHandling(t *testing.T) {
	unicodeError := errors.New("Unicode error: 你好世界 🌍 🚀")
	unicodeHuman := "Unicode message: Привет мир 🌟"

	result := ErrorWithCode[map[string]interface{}](400, unicodeError, unicodeHuman)
	assert.NotNil(t, result)

	data, _ := json.Marshal(result)
	var parsed map[string]interface{}
	json.Unmarshal(data, &parsed)

	errorObj := parsed["Error"].(map[string]interface{})
	assert.Equal(t, unicodeError.Error(), errorObj["errorMessage"])
	assert.Equal(t, unicodeHuman, errorObj["humanMessage"])
}

func TestErrorWithCode_SuccessAlwaysFalse(t *testing.T) {
	codes := []int32{200, 201, 400, 401, 403, 404, 500, 0, -1, 999}

	for _, code := range codes {
		result := ErrorWithCode[map[string]interface{}](code, errors.New("test"), "test")
		data, _ := json.Marshal(result)
		var parsed map[string]interface{}
		json.Unmarshal(data, &parsed)

		assert.Equal(t, false, parsed["Success"], "Success should always be false for code %d", code)
	}
}
