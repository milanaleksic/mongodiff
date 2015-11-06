#!/bin/bash

rm bindata_debug.go 2>/dev/null
go-bindata -nocompress=true -nomemcopy=true -o=bindata_release.go data/...
