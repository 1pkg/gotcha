# Go monkeypatching :monkey_face: :monkey:

Actual arbitrary monkeypatching for Go. Yes really.

Read this blogpost for an explanation on how it works: https://bou.ke/blog/monkey-patching-in-go/

## Notes

1. Monkey sometimes fails to patch a function if inlining is enabled. Try running your tests with inlining disabled, for example: `go test -gcflags=-l`. The same command line argument can also be used for build.
2. Monkey won't work on some security-oriented operating system that don't allow memory pages to be both write and execute at the same time. With the current approach there's not really a reliable fix for this.
3. Monkey is not threadsafe. Or any kind of safe.
4. I've tested monkey on OSX 10.10.2 and Ubuntu 14.04. It should work on any unix-based x86 or x86-64 system.
5. `printf '\x07' | dd of=binary bs=1 seek=160 count=1 conv=notrunc` to overcome newer mac restrictions https://github.com/agiledragon/gomonkey/issues/10

© Bouke van der Bijl
