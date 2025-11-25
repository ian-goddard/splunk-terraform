package appvalidation

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/splunk/terraform-provider-scp/appinspect"
	"github.com/stretchr/testify/assert"
)

func TestAppStatusValidationRead_Success(t *testing.T) {
	mockResp := `{"request_id":"abc","status":"success","links":[]}`
	client := &appinspect.MockClient{ResponseBody: mockResp, StatusCode: http.StatusOK}
	result, state, err := AppStatusValidationRead(context.Background(), client, "abc")()

	assert.NoError(t, err)
	assert.Equal(t, "success", state)
	validationResult, ok := result.(*appinspect.WaitPrivateAppValidationRead)
	assert.True(t, ok)
	assert.Equal(t, "success", validationResult.Status)
}

func TestAppStatusValidationRead_Processing(t *testing.T) {
	mockResp := `{"request_id":"abc","status":"processing","links":[]}`
	client := &appinspect.MockClient{ResponseBody: mockResp, StatusCode: http.StatusOK}
	result, state, err := AppStatusValidationRead(context.Background(), client, "abc")()

	assert.NoError(t, err)
	assert.Equal(t, "processing", state)
	validationResult, ok := result.(*appinspect.WaitPrivateAppValidationRead)
	assert.True(t, ok)
	assert.Equal(t, "processing", validationResult.Status)
}

func TestAppStatusValidationRead_Error(t *testing.T) {
	mockResp := `{"request_id":"abc","status":"error","links":[]}`
	client := &appinspect.MockClient{ResponseBody: mockResp, StatusCode: http.StatusOK}
	result, state, err := AppStatusValidationRead(context.Background(), client, "abc")()

	assert.Error(t, err)
	assert.Equal(t, "error", state)
	assert.Nil(t, result)
}

func TestAppStatusValidationRead_UnknownStatus(t *testing.T) {
	mockResp := `{"request_id":"abc","status":"foobar","links":[]}`
	client := &appinspect.MockClient{ResponseBody: mockResp, StatusCode: http.StatusOK}
	result, state, err := AppStatusValidationRead(context.Background(), client, "abc")()

	assert.Error(t, err)
	assert.Equal(t, "foobar", state)
	assert.Nil(t, result)
}

func TestAppStatusValidationRead_Non200Status(t *testing.T) {
	client := &appinspect.MockClient{Err: fmt.Errorf("check validation status failed with status %d: %s", http.StatusBadRequest, "bad request")}
	result, state, err := AppStatusValidationRead(context.Background(), client, "abc")()

	assert.Error(t, err)
	assert.Equal(t, "", state)
	assert.Nil(t, result)
}
