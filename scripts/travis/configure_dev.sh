#!/usr/bin/env bash

# keep script execution on errors
set +e

SCRIPTPATH="$( cd "$(dirname "$0")" ; pwd -P )"
OS=$("${SCRIPTPATH}/../ostype.sh")
ARCH=$("${SCRIPTPATH}/../archtype.sh")
export DEBIAN_FRONTEND="noninteractive"
if [[ "${OS}" == "linux" ]]; then
    if [[ "${ARCH}" == "arm64" ]]; then
        set -e
        sudo apt-get update -y
        sudo apt-get -y install sqlite3 python3-venv libffi-dev libssl-dev
    elif [[ "${ARCH}" == "arm" ]]; then
        sudo sh -c 'echo "CONF_SWAPSIZE=1024" > /etc/dphys-swapfile; dphys-swapfile setup; dphys-swapfile swapon'
        set -e
        sudo apt-get update -y
        sudo apt-get -y install sqlite3
    fi
elif [[ "${OS}" == "windows" ]]; then
    git config --global core.autocrlf true
    # Golang probably is not installed under MSYS2 so add the environment variable temporarily
    export GOPATH=$HOME/go
    mkdir -p $GOPATH/bin
fi

"${SCRIPTPATH}/../configure_dev.sh"
exit $?
