#!/bin/bash
set -euo pipefail
source ./scripts/gitlab-ci/ssh_add.sh

ssh_add
scp pkg/alicloud-sd root@172.19.24.74:/usr/local/bin/
