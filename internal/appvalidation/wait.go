package appvalidation

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/splunk/terraform-provider-scp/appinspect"
	"github.com/splunk/terraform-provider-scp/internal/wait"
)

func WaitAppValidationRead(ctx context.Context, appInspectClient appinspect.ClientInterface, requestID string) (*appinspect.WaitPrivateAppValidationRead, error) {
	waitAppValidationRead := wait.GenerateReadStateChangeConf(wait.PendingStatusCRUD, []string{"success"}, AppStatusValidationRead(ctx, appInspectClient, requestID))
	output, err := waitAppValidationRead.WaitForStateContext(ctx)

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error reading App validation request ID %s: %s", requestID, err))
		return nil, err
	}

	validation := output.(*appinspect.WaitPrivateAppValidationRead)
	return validation, nil
}
