#! /bin/bash

if [ "$#" -ne 1 ]; then
  echo "usage : $0 server" 
  exit 1
fi

SERVER=$1

curl http://$SERVER/metadata -XPUT -H 'Content-Type: application/json' -d '
{
  "mappings":{
    "objects":{
      "properties":{
        "name":{
          "type":"text",
          "fielddata": true
        },
        "version":{
          "type":"integer"
        },
        "size":{
          "type":"integer"
        },
        "hash":{
          "type":"text"
        }
      }
    }
  }
}'