#!/bin/sh

# Refresh local-dev
echo Refreshing local-dev...
git submodule update --remote --merge
echo Done!

# Install the grpc go plugins
echo "alias apply='/workspaces/auth/local-dev/apply-proto-go.sh'" >> /home/vscode/.bashrc

$(dirname "$0")/install-proto-plugins.sh
