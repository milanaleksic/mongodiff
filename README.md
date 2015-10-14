## Preconditions

Expectations:

    export GO15VENDOREXPERIMENT=1
    go get github.com/kardianos/govendor
    
Package in vendor an existing 3rd party library dependency

    govendor list
    govendor add <import>

### IntelliJ

Until IntelliJ starts supporting officially the 1.5 vendor experiment, expectations is that
each library is _still_ manually fetched via `go get` although it shouldn't really be any need for that

More info: [Issue 1820 for Intellij plugin for GoLang](https://github.com/go-lang-plugin-org/go-lang-idea-plugin/issues/1820)
