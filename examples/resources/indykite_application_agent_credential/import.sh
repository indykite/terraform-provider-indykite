#!/bin/bash
# pipefail - BASH only, not supported in POSIX Shell
set -o errexit -o nounset -o pipefail

terraform import indykite_application_agent_credential.id gid:AAABBBCCC_000111222333
