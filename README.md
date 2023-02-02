bytes is a small utility for working with small binary artifacts

It converts between programming-language syntax for byte arrays and actual
binary. This is primarily useful if you're trying to debug or modify something
you have the bits for using a text editor, or if you have a test fixture inline
in the tests you want to extract to binary.

To install:

```
go install github.com/ConradIrwin/bytes@latest
```

To use:

```
usage: bytes [-d|--decode|--rust|--go] <file>?

bytes formats binary input as a []byte{} array for use in go code, or a vec![] for rust.

If no file name is provided, bytes reads from stdin

If -d or --decode is passed the transformation is reversed, and formatted bytes
are output as binary. Supported input formats are valid go []bytes{} and rust
vec![]'s.  Care is taken to remove comments, spaces, semicolons, etc. so you can
paste directly from code.  As a special case bytes can also decode go fuzz fixture files
containing bytes.
  -d
  -decode
    	decode formatted bytes and output binary
  -go
    	output in go syntax (default)
  -rust
    	output in rust syntax
```

Bug reports and pull requests welcome!
