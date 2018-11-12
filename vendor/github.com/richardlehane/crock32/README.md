Implementation of Douglas Crockford's Base32 encoding.

Example usage:

    i, _ := crock32.Decode("a1j3")
    s := crock32.Encode(i)
    fmt.Println(s)

Crock32 is useful for "expressing numbers in a form that can be conveniently and accurately transmitted between humans and computer systems".

See [http://www.crockford.com/wrmg/base32.html](http://www.crockford.com/wrmg/base32.html) for details.

Note: crock32 differs from Crockford in its use of lower-case letters when encoding (decode works for both cases). To change, use: `crock32.SetDigits("0123456789ABCDEFGHJKMNPQRSTVWXYZ")`

Install with `go get github.com/richardlehane/crock32`
