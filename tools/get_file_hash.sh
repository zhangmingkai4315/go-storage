#! /bin/bash

if [ "$#" -ne 1 ] || ! [ -f "$1" ]; then
  echo "usage : $0 file_path" 
  exit 1
fi

HASH=`openssl dgst -sha256 -binary $1 | base64`
echo "Digest: SHA-256=$HASH"