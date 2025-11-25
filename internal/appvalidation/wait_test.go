package appvalidation

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/splunk/terraform-provider-scp/appinspect"
	"github.com/stretchr/testify/assert"
)

var (
	unexpectedStatusCodes = []int{400, 401, 403, 404, 409, 501}
)

func Test_WaitAppInspectRead(t *testing.T) {
	client := &appinspect.MockClient{}
	mockRequestID := "mockRequestID"

	t.Run("with some client interface error", func(_ *testing.T) {
		client.Err = errors.New("some client error")

		validation, err := WaitAppValidationRead(context.TODO(), client, mockRequestID)
		assert.Error(t, err)
		assert.Nil(t, validation)
	})

	t.Run("with http 200 response", func(t *testing.T) {
		client.Err = nil
		client.StatusCode = 200
		client.ResponseBody = `{"request_id": "mockRequestID","status": "success"}`
		validation, err := WaitAppValidationRead(context.TODO(), client, mockRequestID)
		assert.NoError(t, err)
		assert.NotNil(t, validation)
		assert.NotNil(t, validation.Status)
		assert.Equal(t, "success", validation.Status)
	})

	t.Run("with unexpected response", func(t *testing.T) {
		for _, code := range unexpectedStatusCodes {
			t.Run(fmt.Sprintf("with unexpected response %v", code), func(t *testing.T) {
				client.Err = nil
				client.StatusCode = code
				client.ResponseBody = `{"message": "You are not authorized to access request mockRequestID or the request was not found", "request_id": "mockRequestID"}`
				validation, err := WaitAppValidationRead(context.TODO(), client, mockRequestID)
				assert.Error(t, err)
				assert.Nil(t, validation)
			})
		}
	})
}
