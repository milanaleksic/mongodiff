# Record & replay changes to a MongoDB

[![Build Status](https://semaphoreci.com/api/v1/milanaleksic/mongodiff/branches/master/badge.svg)](https://semaphoreci.com/milanaleksic/mongodiff)

This utility makes Windows and Linux scripts to reproduce manual actions done on a Mongo DB (or actions that are result of some script / acceptance test).

It makes internally a simple diff (*only new items* are being detected, updates are ignored currently).

As a result, one should get `BASH`, `BAT`, `JS` and `JSON` files that altogether work to reproduce actions when it's needed.

To see what options are available, please run application with `--help` parameter

## How do generated scripts know on which server they need to execute insertions?

They don't, you should either:

1. give the server as script parameter; or 
2. set the env variable `MONGO_SERVER` to some IP/host name before you run the shell/batch file.

For example:

    clean.bat localhost
    setup.bat 192.168.1.101:27117
    MONGO_SERVER=127.0.0.1 setup.sh

## Why?

Weekly Scrum Demos. This tool makes it a breeze for most cases which might otherwise take too much preparation.

## (For developers) Building, tagging and artifact deployment

This is `#golang` project. I used Go 1.6. 

`go get github.com/milanaleksic/mongodiff` should be enough to get the code and build. 

To build project you can execute (this will get from internet all 3rd party utilites needed for deployment: upx, go-upx, github-release):

    make prepare

You can start building project using `make`, even `deploy` to Github (if you have privileges to do that of course).
