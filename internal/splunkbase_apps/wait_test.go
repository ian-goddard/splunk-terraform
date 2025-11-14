package splunkbaseapps_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/acs/v2/mocks"
	splunkbaseapps "github.com/splunk/terraform-provider-scp/internal/splunkbase_apps"
	"github.com/stretchr/testify/assert"
)

const (
	mockStack = v2.Stack("mock-stack")
)

func Test_WaitAppCreate(t *testing.T) {
	splunkbase := true
	acsLicensing := "https://www.splunk.com/en_us/legal/splunk-general-terms.html"
	client := &mocks.ClientInterface{}
	params := v2.InstallAppVictoriaParams{
		Splunkbase:               &splunkbase,
		XSplunkbaseAuthorization: nil,
		XSplunkAuthorization:     nil,
		ACSLegalAck:              nil,
		ACSLicensingAck:          &acsLicensing,
	}
	mockBody := io.NopCloser(strings.NewReader("{}"))

	t.Run("with http response 202", func(t *testing.T) {
		client.On("InstallAppVictoriaWithBody", context.TODO(), mockStack, &params, "application/x-www-form-urlencoded", mockBody).Return(generateResponse(202), nil).Once()

		err := splunkbaseapps.WaitAppCreate(context.TODO(), client, mockStack, params, mockBody)
		assert.Nil(t, err)
	})

	t.Run("with http response 400", func(t *testing.T) {
		client.On("InstallAppVictoriaWithBody", context.TODO(), mockStack, &params, "application/x-www-form-urlencoded", mockBody).Return(generateResponse(400), nil).Once()

		err := splunkbaseapps.WaitAppCreate(context.TODO(), client, mockStack, params, mockBody)
		assert.Error(t, err.Err)
	})
	t.Run("with http response 409", func(t *testing.T) {
		client.On("InstallAppVictoriaWithBody", context.TODO(), mockStack, &params, "application/x-www-form-urlencoded", mockBody).Return(generateResponse(409), nil).Once()

		err := splunkbaseapps.WaitAppCreate(context.TODO(), client, mockStack, params, mockBody)
		assert.Error(t, err.Err)
	})
	t.Run("with http response 500", func(t *testing.T) {
		client.On("InstallAppVictoriaWithBody", context.TODO(), mockStack, &params, "application/x-www-form-urlencoded", mockBody).Return(generateResponse(500), nil).Once()

		err := splunkbaseapps.WaitAppCreate(context.TODO(), client, mockStack, params, mockBody)
		assert.Error(t, err.Err)
	})
	t.Run("with http response 503", func(t *testing.T) {
		client.On("InstallAppVictoriaWithBody", context.TODO(), mockStack, &params, "application/x-www-form-urlencoded", mockBody).Return(generateResponse(503), nil).Once()

		err := splunkbaseapps.WaitAppCreate(context.TODO(), client, mockStack, params, mockBody)
		assert.Error(t, err.Err)
	})

	t.Run("with http response 408", func(t *testing.T) {
		client.On("InstallAppVictoriaWithBody", context.TODO(), mockStack, &params, "application/x-www-form-urlencoded", mockBody).Return(generateResponse(408), nil).Once()

		err := splunkbaseapps.WaitAppCreate(context.TODO(), client, mockStack, params, mockBody)
		assert.Error(t, err.Err)
	})
}

func generateResponse(code int) *http.Response {
	var b []byte
	if code == http.StatusAccepted {

		victoriaApp := v2.App{
			Name:   "mock-app",
			Status: "mock-status",
		}

		b, _ = json.Marshal(&victoriaApp)
	} else {
		b, _ = json.Marshal(&v2.Error{
			Code:    http.StatusText(code),
			Message: http.StatusText(code),
		})
	}
	recorder := httptest.NewRecorder()
	recorder.Header().Add("Content-Type", "json")
	recorder.WriteHeader(code)
	if b != nil {
		_, _ = recorder.Write(b)
	}
	return recorder.Result()
}
