package response_test

import (
	"encoding/json"
	"testing"
	"time"

	"code-base-golang/internal/pkg/response"
)

func TestAPIResponse_Serialization(t *testing.T) {
	now := time.Now()
	res := response.APIResponse[string]{
		Status:    "success",
		Code:      "200",
		Message:   "OK",
		Data:      "test-payload",
		Timestamp: now,
	}

	bytes, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("failed to marshal APIResponse: %v", err)
	}

	var parsed response.APIResponse[string]
	if err := json.Unmarshal(bytes, &parsed); err != nil {
		t.Fatalf("failed to unmarshal APIResponse: %v", err)
	}

	if parsed.Status != "success" || parsed.Data != "test-payload" {
		t.Errorf("unexpected parsed response: %+v", parsed)
	}
}

func TestBaseResponse_Serialization(t *testing.T) {
	now := time.Now()
	res := response.BaseResponse{
		Status:    "error",
		Code:      "500",
		Message:   "internal error",
		Timestamp: now,
	}

	bytes, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("failed to marshal BaseResponse: %v", err)
	}

	var parsed response.BaseResponse
	if err := json.Unmarshal(bytes, &parsed); err != nil {
		t.Fatalf("failed to unmarshal BaseResponse: %v", err)
	}

	if parsed.Status != "error" || parsed.Code != "500" {
		t.Errorf("unexpected parsed base response: %+v", parsed)
	}
}
