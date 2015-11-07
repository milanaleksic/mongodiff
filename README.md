# Record & replay changes to a MongoDB

This utility makes a script to reproduce manual actions done on a Mongo DB.

It makes internally a simple diff (*only new items* are being detected, updates are ignored currently).

As a result, one should get `BASH`, `BAT`, `JS` and `JSON` files that alltogether work to reproduce actions when it's needed.
 
To see what options are available, please run application with `--help` parameter

## Why?

Weekly Scrum Demos. This tool makes it a breeze for most cases which might otherwise take too much preparation.


## Build/release preconditions

This is `#golang` project. I used Go 1.5. `go get github.com/milanaleksic/mongodiff` should be enough.

To deploy I used also additional `go`-driven utilities:

1. go-bindata
2. github-release
3. upx 3.91
4. goupx (goes around bug of upx-ing linux golang-generated binaries)

### Testing

I made only a trivial sanity-checking test that demands Mongo running - otherwise test will fail.
