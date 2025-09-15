# https://github.com/terraform-linters/tflint/blob/master/docs/user-guide/config.md

tflint { required_version = ">= 0.55" }

config {
  format     = "compact"
  plugin_dir = "~/.tflint.d/plugins"
}

plugin "terraform" {
  enabled = true
  preset  = "recommended"
}

# disabled since mainly are examples in the TF code
rule "terraform_required_version" { enabled = false }
rule "terraform_required_providers" { enabled = false }
rule "terraform_unused_declarations" { enabled = false }
