package splunkbaseapps

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/client"
	"github.com/splunk/terraform-provider-scp/internal/errors"
	"github.com/splunk/terraform-provider-scp/internal/locks"
	"github.com/splunk/terraform-provider-scp/internal/wait"
)

const (
	ResourceKey     = "scp_splunkbase_app"
	splunkbaseID    = "splunkbase_id"
	AcsLicensingAck = "acs_licensing_ack"
	RetryTimeout    = 30 * time.Minute
)

func splunkbaseAppSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the Splunkbase app",
		},
		splunkbaseID: {
			Type:        schema.TypeString,
			Description: "The ID of the Splunkbase app",
			Required:    true,
		},
		"version": {
			Type:        schema.TypeString,
			Description: "The version of the Splunkbase app",
			Required:    true,
		},
		AcsLicensingAck: {
			Type:        schema.TypeString,
			Description: "The app's third-party license URL. The license URL is available under 'Licensing' on the Splunkbase download page for the app.",
			Required:    true,
		},
	}
}

func ResourceSplunkbaseApp() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Splunkbase App. Please refer to https://docs.splunk.com/Documentation/SplunkCloud/9.3.2408/Config/ManageSplunkbaseApps for more latest, detailed information on attribute requirements and the ACS API.",
		CreateContext: resourceSplunkbaseAppCreate,
		UpdateContext: resourceSplunkbaseAppUpdate,
		ReadContext:   resourceSplunkbaseAppRead,
		DeleteContext: resourceSplunkbaseAppDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: splunkbaseAppSchema(),
	}
}

func resourceSplunkbaseAppCreate(ctx context.Context, resourceData *schema.ResourceData, m interface{}) diag.Diagnostics {
	// use the meta value to retrieve client and stack from the provider configure method
	acsProvider := m.(client.ACSProvider)
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack
	splunkbase := true

	// Acquire lock for app operations on this stack
	lockManager := locks.GetAppLockManager()
	unlock := lockManager.LockAppOperation(ctx, stack, "create")
	defer unlock()

	ACSLicensingAck := resourceData.Get("acs_licensing_ack").(string)
	installParams := v2.InstallAppVictoriaParams{
		Splunkbase:      &splunkbase,
		ACSLicensingAck: &ACSLicensingAck,
	}

	version, ok := resourceData.GetOk("version")
	if !ok || version.(string) == "" {
		return diag.Errorf("version must be provided")
	}
	name, ok := resourceData.GetOk("name")
	if !ok || name.(string) == "" {
		return diag.Errorf("name must be provided")
	}

	data := url.Values{}
	data.Set("name", name.(string))
	data.Set("version", version.(string))
	splunkbaseIDParam, ok := resourceData.GetOk("splunkbase_id")
	if !ok || splunkbaseIDParam.(string) == "" {
		return diag.Errorf("splunkbase_id must be provided")
	}

	data.Set("splunkbaseID", splunkbaseIDParam.(string))
	body := strings.NewReader(data.Encode())
	err := resource.RetryContext(ctx, RetryTimeout, func() *resource.RetryError {
		err := WaitAppCreate(ctx, acsClient, stack, installParams, body)
		if err != nil {
			if errors.IsConflictError(err.Err) {
				return resource.NonRetryableError(err.Err)
			}
			if strings.Contains(err.Err.Error(), "503") {
				return resource.RetryableError(fmt.Errorf("received 503 error, retrying: %v", err.Err))
			}
			if err.Retryable {
				return resource.RetryableError(fmt.Errorf("retryable error occurred: %v", err.Err))
			}
			return resource.NonRetryableError(err.Err)
		}
		return nil
	})

	if err != nil {
		if errors.IsConflictError(err) {
			tflog.Info(ctx, "App (%s) already exists, if you want to update it, change app's version")
			resourceData.SetId(resourceData.Get("name").(string))
		} else {
			return diag.Errorf("Error submitting request for app to be created. %v", err)
		}
	}

	err = resource.RetryContext(ctx, RetryTimeout, func() *resource.RetryError {
		tflog.Info(ctx, "[BETA] Splunkbase Apps: This feature is in beta release.")
		err := WaitAppPoll(ctx, acsClient, stack, resourceData.Get("name").(string), wait.TargetStatusResourceExists, wait.PendingStatusVerifyCreated)
		if err != nil {
			if strings.Contains(err.Error(), "503") {
				return resource.RetryableError(fmt.Errorf("received 503 error, retrying: %w", err))
			}
			if strings.Contains(err.Error(), "404") {
				return resource.RetryableError(fmt.Errorf("received 404 error, retrying: %w", err))
			}
			return resource.NonRetryableError(fmt.Errorf("error waiting for app (%s) to be created: %s", resourceData.Get("name").(string), err))
		}
		return nil
	})
	if err != nil {
		return diag.Errorf("Error waiting for app (%s) to be created: %s", resourceData.Get("name").(string), err)
	}
	resourceData.SetId(resourceData.Get("name").(string))
	return nil
}

func resourceSplunkbaseAppRead(ctx context.Context, resourceData *schema.ResourceData, m interface{}) diag.Diagnostics {
	acsProvider := m.(client.ACSProvider)
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack

	appName := resourceData.Id()

	err := resource.RetryContext(ctx, RetryTimeout, func() *resource.RetryError {
		_, err := WaitAppRead(ctx, acsClient, stack, appName)
		if err != nil {
			if stateErr, ok := err.(*resource.UnexpectedStateError); ok && strings.Contains(stateErr.LastError.Error(), "404-app-not-found") {
				return resource.NonRetryableError(err)
			}
			if strings.Contains(err.Error(), "503") {
				return resource.RetryableError(fmt.Errorf("received 503 error, retrying: %w", err))
			}
			if strings.Contains(err.Error(), "400") {
				return resource.RetryableError(fmt.Errorf("received 400 error, retrying: %w", err))
			}
			if strings.Contains(err.Error(), "404") {
				return resource.RetryableError(fmt.Errorf("received 404 error, retrying: %w", err))
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		if stateErr, ok := err.(*resource.UnexpectedStateError); ok && strings.Contains(stateErr.LastError.Error(), "404-app-not-found") {
			return nil
		}
		return diag.Errorf("Error reading app (%s): %s", appName, err)
	}

	return nil
}

func resourceSplunkbaseAppDelete(ctx context.Context, resourceData *schema.ResourceData, m interface{}) diag.Diagnostics {
	acsProvider := m.(client.ACSProvider)
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack

	// Acquire lock for app operations on this stack
	lockManager := locks.GetAppLockManager()
	unlock := lockManager.LockAppOperation(ctx, stack, "delete")
	defer unlock()

	appName := resourceData.Id()

	retryErr := resource.RetryContext(ctx, 2*RetryTimeout, func() *resource.RetryError {
		err := WaitAppDelete(ctx, acsClient, stack, appName)
		if err != nil {
			if strings.Contains(err.Error(), "503") {
				return resource.RetryableError(fmt.Errorf("received 503 error, retrying: %w", err))
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if retryErr != nil {
		return diag.Errorf("Error deleting app (%s): %s", appName, retryErr)
	}

	retryErr = resource.RetryContext(ctx, RetryTimeout, func() *resource.RetryError {
		tflog.Info(ctx, "[BETA] Splunkbase Apps: This feature is in beta release.")
		err := WaitAppPoll(ctx, acsClient, stack, appName, wait.TargetStatusResourceDeleted, wait.PendingStatusVerifyDeleted)
		if err != nil {
			if strings.Contains(err.Error(), "503") {
				return resource.RetryableError(fmt.Errorf("received 503 error, retrying: %w", err))
			}
			return resource.NonRetryableError(fmt.Errorf("error waiting for app (%s) to be deleted: %s", appName, err))
		}
		return nil
	})
	if retryErr != nil {
		return diag.Errorf("Error waiting for app (%s) to be deleted: %s", appName, retryErr)
	}
	return nil
}

func resourceSplunkbaseAppUpdate(ctx context.Context, resourceData *schema.ResourceData, m interface{}) diag.Diagnostics {
	tflog.Info(ctx, "[BETA] Splunkbase Apps: This feature is in beta release.")
	// use the meta value to retrieve client and stack from the provider configure method
	acsProvider := m.(client.ACSProvider)
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack

	// Acquire lock for app operations on this stack
	lockManager := locks.GetAppLockManager()
	unlock := lockManager.LockAppOperation(ctx, stack, "update")
	defer unlock()

	appName := resourceData.Id()

	ACSLicensingAck := resourceData.Get("acs_licensing_ack").(string)
	installParams := v2.PatchAppVictoriaParams{
		ACSLicensingAck: ACSLicensingAck,
	}

	data := url.Values{}
	data.Set("name", resourceData.Get("name").(string))
	data.Set("version", resourceData.Get("version").(string))
	data.Set("splunkbaseID", resourceData.Get("splunkbase_id").(string))

	body := strings.NewReader(data.Encode())

	retryErr := resource.RetryContext(ctx, RetryTimeout, func() *resource.RetryError {
		err := WaitAppUpdate(ctx, acsClient, stack, appName, installParams, body)
		if err != nil {
			if strings.Contains(err.Error(), "503") {
				return resource.RetryableError(fmt.Errorf("received 503 error, retrying: %w", err))
			}
			if strings.Contains(err.Error(), "404") {
				return resource.RetryableError(fmt.Errorf("received 404 error, retrying: %w", err))
			}
			return resource.NonRetryableError(fmt.Errorf("error updating app (%s): %s", appName, err))
		}
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("error waiting for app (%s) to be deleted: %s", appName, err))
		}
		return nil
	})
	if retryErr != nil {
		return diag.Errorf("Error updating app (%s): %s", appName, retryErr)
	}

	retryErr = resource.RetryContext(ctx, RetryTimeout, func() *resource.RetryError {
		app, err := WaitAppRead(ctx, acsClient, stack, appName)
		if err != nil {
			if strings.Contains(err.Error(), "503") {
				return resource.RetryableError(fmt.Errorf("received 503 error, retrying: %w", err))
			}
			return resource.NonRetryableError(fmt.Errorf("error waiting for app (%s) to be deleted: %s", appName, err))
		}
		if *app.Version != resourceData.Get("version").(string) {
			return resource.RetryableError(fmt.Errorf("app version (%s) does not match the expected version (%s), retrying", *app.Version, resourceData.Get("version").(string)))
		}
		return nil
	})
	if retryErr != nil {
		tflog.Error(ctx, "Error updating app.")
	}
	return nil
}
