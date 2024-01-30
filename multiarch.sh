#!/bin/bash

ARCH=`uname -m`
if [ "$ARCH" = "unknown" ]; then
  # Debian 9 gives unknown on this call
  ARCH=`uname -m`
fi
APP=`basename $0`
DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
echo ${DIR}

exec "${DIR}/${ARCH}-linux-gnu/${APP}" "$@"

