# scp_private_apps (Resource)
## THIS IS BETA RELEASE 
### THIS PROVIDER IS AVAILABLE ONLY FOR SPLUNK CLOUD VICTORIA EXPERIENCE

Default parallelism for terraform operations is 10. Due to the sequential nature of app installation, it is recommended to use the --parallelism=1 flag when applying Terraform changes with this resource (or at least some number < 10).

## Example Usage

```terraform
resource "scp_private_apps" "example" {
  name         = "my-private-app"
  filename     = "/path/to/app.tgz"
  acs_legal_ack = "Y"
  pre_vetted = true 
}
```

## Schema

### Required

- `name` (`String`): The name of the private app.
- `filename` (`String`): The path to the private app file. The file must be a valid tar.gz archive.
- `acs_legal_ack` (`String`): When you install a private app, you must specify this parameter to acknowledge acceptance of the Splunk legal disclaimer for app installation. See [Set up the ACS API](https://docs.splunk.com/Documentation/SplunkCloud/latest/Config/ACSusage#Set_up_the_ACS_API).
- `pre_vetted` (`Boolean`): Whether the app has been pre-vetted by app inspect.
## Timeouts
Defaults are currently set to:
- `create` -  20m
- `read` -  20m
- `update` -  20m
- `delete` -  40m

