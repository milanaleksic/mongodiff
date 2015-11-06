#!/bin/bash

rm bindata_release.go 2>/dev/null
go-bindata --debug -o=bindata_debug.go data/...
