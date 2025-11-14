terraform {
  required_providers {
    scp = {
      source = "registry.terraform.io/splunk/scp"
    }
  }
}

provider "scp" {
  stack = "example-stack"
  server = "https://admin.splunk.com"
  username = "admin"
  password = var.password
  splunk_username = var.splunk_username
  splunk_password = var.splunk_password
}

resource "scp_indexes" "index-1" {
  name = "index-1"
}

resource "scp_indexes" "index-2" {
  name = "index-2"
  searchable_days = 90
}

resource "scp_indexes" "index-3" {
  name             = "index-3"
  searchable_days  = 90
  max_data_size_mb = 512
}

resource "scp_indexes" "index-4" {
  name             = "index-4"
  searchable_days  = 90
  max_data_size_mb = 512
}

data "scp_indexes" "main" {
  name = "main"
}

data "scp_indexes" "history" {
  name = "history"
}

resource "scp_splunkbase_app" "chargeback_app_splunk_cloud" {
  name             = "chargeback_app_splunk_cloud"
  acs_licensing_ack = "https://www.splunk.com/en_us/legal/splunk-general-terms.html"
  version        = "2.0.53"
  splunkbase_id   = "5688"
}

resource "scp_splunkbase_app" "broken_hosts" {
  name             = "broken_hosts"
  acs_licensing_ack = "https://opensource.org/licenses/MIT"
  version        = "5.0.4"
  splunkbase_id   = "3247"
}

resource "scp_splunkbase_app" "DomainTools-App-for-Splunk" {
  name             = "DomainTools-App-for-Splunk"
  acs_licensing_ack = "https://cdn.splunkbase.splunk.com/static/misc/eula.html"
  version        = "5.4.0"
  splunkbase_id   = "5226"
}

