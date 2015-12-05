# Record & replay changes to a MongoDB

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

## Why?

Weekly Scrum Demos. This tool makes it a breeze for most cases which might otherwise take too much preparation.


## Building, tagging and artifact deployment

This is `#golang` project. I used Go 1.5. 

`go get github.com/milanaleksic/mongodiff` should be enough to get the code and build. 

I also use these utilities for various stages of post-compilation development:

1. `go get github.com/aktau/github-release`
2. `go get github.com/jteeuwen/go-bindata`
3. `go get github.com/pwaller/goupx` (goes around bug of upx-ing linux golang-generated binaries)
4. upx 3.91

