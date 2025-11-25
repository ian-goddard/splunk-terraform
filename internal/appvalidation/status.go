package appvalidation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/splunk/terraform-provider-scp/appinspect"
	"github.com/splunk/terraform-provider-scp/internal/wait"
)

func AppStatusValidationRead(ctx context.Context, appInspectClient appinspect.ClientInterface, requestID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := appInspectClient.CheckValidationStatus(requestID)
		if err != nil {
			return nil, "", &resource.UnexpectedStateError{LastError: err}
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				tflog.Error(ctx, fmt.Sprintf("Error closing response body: %v", err))
			}
		}(resp.Body)

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, "", &resource.UnexpectedStateError{LastError: err}
		}

		if resp.StatusCode != http.StatusOK {
			return nil, http.StatusText(resp.StatusCode), &resource.UnexpectedStateError{
				State:         http.StatusText(resp.StatusCode),
				ExpectedState: wait.TargetStatusResourceExists,
				LastError:     errors.New(string(bodyBytes)),
			}
		}

		var statusResponseMap map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &statusResponseMap); err != nil {
			return nil, "", &resource.UnexpectedStateError{LastError: err}
		}

		var validationResult appinspect.WaitPrivateAppValidationRead
		if err := json.Unmarshal(bodyBytes, &validationResult); err != nil {
			return nil, "", &resource.UnexpectedStateError{LastError: err}
		}

		status := strings.ToLower(strings.TrimSpace(validationResult.Status))

		switch status {
		case "success":
			return &validationResult, "success", nil
		case "processing":
			return &validationResult, "processing", nil
		case "error":
			return nil, "error", fmt.Errorf("app validation failed with status: %s", status)
		default:
			return nil, status, fmt.Errorf("unknown status: %s", status)
		}
	}
}
