#!/bin/bash

if [ "$GITHUB_TOKEN" = "" ]
then
    echo "GITHUB_TOKEN is not set!"
    exit 1
fi

if [ "$1" = "" ]
then
    echo "Which tag? No argument given to the application"
    exit 1
fi

git tag $1
git push --tags

github-release release -u milanaleksic -r mongodiff --tag $1 --name "v$1"

GOOS=windows
go build
upx mongodiff.exe
github-release upload -u milanaleksic -r mongodiff --tag $1 --name "mongodiff-$1-windows-amd64.exe" -f mongodiff.exe
rm mongodiff.exe

GOOS=linux
go build
goupx mongodiff
github-release upload -u milanaleksic -r mongodiff --tag $1 --name "mongodiff-$1-linux-amd64" -f mongodiff
rm mongodiff

