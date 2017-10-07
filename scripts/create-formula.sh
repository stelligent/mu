#!/bin/bash
function usage {
  echo Usage $0 version branch
}
if [ $# -lt 2 ]; then
  usage
  exit 1
fi
VERSION=$1
BRANCH=$2
if [ "x${BRANCH}" != "xmaster" ] && [ "x${BRANCH}" != "xdevelop" ]; then
  usage
  exit 1
fi
FILE=$(cat mu-cli.rb)
# Download binaries and get a hash
MAC_URL="https://github.com/stelligent/mu/releases/download/${VERSION}/mu-darwin-amd64"
LINUX_URL="https://github.com/stelligent/mu/releases/download/${VERSION}/mu-linux-amd64"
wget -O mu-darwin-amd64-homebrew $MAC_URL
wget -O mu-linux-amd64-homebrew $LINUX_URL
MAC_SHA256=$(shasum -a 256 mu-darwin-amd64-homebrew | cut -d' ' -f1)
LINUX_SHA256=$(shasum -a 256 mu-linux-amd64-homebrew | cut -d' ' -f1)
# Update formula in mu-cli.rb
sed -i".bak" 's|.*\( # The MacOS '$BRANCH' url\)|    url "'$MAC_URL'"\1|g ;'\
's|.*\( # The MacOS '$BRANCH' sha256sum\)|    sha256 "'$MAC_SHA256'"\1|g;'\
's|.*\( # The Linux '$BRANCH' url\)|    url "'$LINUX_URL'"\1|g;'\
's|.*\( # The Linux '$BRANCH' sha256sum\)|    sha256 "'$LINUX_SHA256'"\1|g;'\
's|\(\s*version\).*\( # The '$BRANCH' version\)|\1 "'$VERSION'"\2|g'\
 homebrew/mu-cli.rb
