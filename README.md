## Preconditions

This utility makes a script to reproduce manual actions done on a Mongo DB.

It makes internally a simple diff (only new items are being detected, updates are ignored currently).

As a result, one should get BASH, JS and JSON files that alltogether work to reproduce actions when it's needed.
 
To see what options are available, please run application with `--help` parameter