#!/bin/sh

{
set -e

REPO=stelligent/mu
LATEST_VERSION=`curl -s https://raw.githubusercontent.com/$REPO/master/VERSION`
INSTALL_VERSION=${INSTALL_VERSION:-$LATEST_VERSION}
INSTALL_DIR=${INSTALL_DIR:-/usr/local/bin}
INSTALL_PATH="${INSTALL_DIR%/}/mu"
touch "$INSTALL_PATH" || { echo "ERROR: Cannot write to $INSTALL_DIR set INSTALL_DIR elsewhere or use sudo"; exit 1; }

arch=""
if [ "$(uname -m| tr '[:upper:]' '[:lower:]')" = "x86_64" ]; then
    arch="amd64"
else
    arch="386"
fi

os=`uname -s| tr '[:upper:]' '[:lower:]'`


url="https://github.com/$REPO/releases/download/v$INSTALL_VERSION/mu-$os-$arch"

echo "Downloading $url"
curl -sL "$url" -o "$INSTALL_PATH"
chmod +rx "$INSTALL_PATH"
echo "mu $INSTALL_VERSION has been installed to $INSTALL_PATH"
}