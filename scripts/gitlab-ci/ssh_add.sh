#!/bin/bash
set -euo pipefail

ssh_add(){
  # Install ssh-agent if not already installed, it is required by Docker.
  # (change apt-get to yum if you use a CentOS-based image)
  #which ssh-agent || ( apt-get update -y && apt-get install -y openssh-client)
  # Run ssh-agent (inside the build environment)
  eval $(ssh-agent -s)
  # Add the SSH key stored in SSH_PRIVATE_KEY_CC variable to the agent store
  ssh-add <(echo "$SSH_PRIVATE_KEY_BETA")
  # For Docker builds disable host key checking. Be aware that by adding that
  # you are suspectible to man-in-the-middle attacks.
  # WARNING: Use this only with the Docker executor, if you use it with shell
  # you will overwrite your user's SSH config.
  mkdir -p ~/.ssh
  [[ -f /.dockerenv ]] && echo -e "Host *\n\tStrictHostKeyChecking no\n\tServerAliveInterval 60\n\n" > ~/.ssh/config
}
