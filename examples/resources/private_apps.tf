resource "scp_private_app" "test" {
  name        = "test_private_app"
  filename    = "../test_app.tar.gz"
  pre_vetted=true
  acs_legal_ack = "Y"
}