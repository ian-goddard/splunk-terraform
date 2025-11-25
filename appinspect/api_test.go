package appinspect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCheckValidationStatusRequest(t *testing.T) {
	client := GetAppInspectClient("dummy token")
	req, err := client.CheckValidationStatus("12345")
	assert.NoError(t, err)
	assert.NotNil(t, req)
	assert.Equal(t, "GET", req.Request.Method)
	assert.Equal(t, "https://appinspect.splunk.com/v1/app/validate/status/12345", req.Request.URL.String())
	assert.Equal(t, "no-cache", req.Request.Header.Get("Cache-Control"))
	assert.Equal(t, "bearer dummy token", req.Request.Header.Get("Authorization"))
}

func TestNewCheckValidationStatusRequest_EmptyID(t *testing.T) {
	client := GetAppInspectClient("dummy token")
	req, err := client.CheckValidationStatus("")
	assert.Error(t, err)
	assert.Nil(t, req)
	assert.Equal(t, "validation ID cannot be empty", err.Error())
}

func TestNewCheckValidationStatusRequest_WhitespaceID(t *testing.T) {
	client := GetAppInspectClient("dummy token")
	req, err := client.CheckValidationStatus("  	 ")
	assert.Error(t, err)
	assert.Nil(t, req)
	assert.Equal(t, "validation ID cannot be empty", err.Error())
}

func TestNewCheckValidationStatusRequest_EmptyToken(t *testing.T) {
	client := GetAppInspectClient("")
	req, err := client.CheckValidationStatus("12345")
	assert.Error(t, err)
	assert.Nil(t, req)
	assert.Equal(t, "authorization token cannot be empty", err.Error())
}

func TestNewCheckValidationStatusRequest_WhitespaceToken(t *testing.T) {
	client := GetAppInspectClient("   ")
	req, err := client.CheckValidationStatus("12345")
	assert.Error(t, err)
	assert.Nil(t, req)
	assert.Equal(t, "authorization token cannot be empty", err.Error())
}
