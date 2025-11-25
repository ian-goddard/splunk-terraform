# scp_splunkbase_apps (Resource)
## THIS IS BETA RELEASE
### IT IS ONLY APPLICABLE FOR SPLUNK VICTORIA CLOUD EXPERIENCE 
Manages Splunkbase apps in Splunk Cloud Platform. For detailed attribute requirements, refer to the [Splunk Cloud documentation](https://docs.splunk.com/Documentation/SplunkCloud/9.3.2408/Config/ManageSplunkbaseApps) and the ACS API.
Default parallelism for terraform operations is 10. Due to the sequential nature of app installation, it is recommended to use the --parallelism=1 flag when applying Terraform changes with this resource (or at least some number < 10).

Due to the nature of app installation, it is recommended to use the --parallelism=1 flag when applying Terraform changes with this resource.
## Example Usage

```terraform
resource "scp_splunkbase_apps" "example" {
  name            = "Splunk_TA_nix"
  version         = "8.6.0"
  splunkbase_id   = "833"
  acs_licensing_ack = "https://www.splunk.com/en_us/legal/splunk-general-terms.html"
}
```


## Schema

### Required

- `name` (`String`): The name of the Splunkbase app.
- `version` (`String`): The version of the Splunkbase app.
- `splunkbase_id` (`String`): The ID of the Splunkbase app.


## Timeouts
Defaults are currently set to:
- `create` -  20m
- `read` -  20m
- `update` -  20m
- `delete` -  40m
