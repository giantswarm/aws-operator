package microstorage

import (
	"fmt"
)

func ExampleNewK() {
	firstKey, _ := NewK("/a/b/c")
	fmt.Println(firstKey.Key())

	secondKey, _ := NewK("a/b/c")
	fmt.Println(secondKey.Key())

	thirdKey, _ := NewK("a/b/c/")
	fmt.Println(thirdKey.Key())
	// Output: /a/b/c
	// /a/b/c
	// /a/b/c
}

func ExampleK_Key() {
	key, _ := NewK("/a/b/c")
	fmt.Println(key.Key())
	// Output: /a/b/c
}

func ExampleK_KeyNoLeadingSlash() {
	key, _ := NewK("/a/b/c")
	fmt.Println(key.KeyNoLeadingSlash())
	// Output: a/b/c
}

func ExampleNewKV() {
	firstKeyValue, _ := NewKV("/a/b/c", "foo")
	fmt.Println(firstKeyValue.Key())
	fmt.Println(firstKeyValue.Val())

	secondKeyValue, _ := NewKV("a/b/c", "bar")
	fmt.Println(secondKeyValue.Key())
	fmt.Println(secondKeyValue.Val())

	thirdKeyValue, _ := NewKV("a/b/c/", "baz")
	fmt.Println(thirdKeyValue.Key())
	fmt.Println(thirdKeyValue.Val())
	// Output: /a/b/c
	// foo
	// /a/b/c
	// bar
	// /a/b/c
	// baz
}

func ExampleKV_K() {
	kv, _ := NewKV("/a/b/c", "foo")
	fmt.Println(kv.K().Key())
	// Output: /a/b/c
}

func ExampleKV_Key() {
	kv, _ := NewKV("/a/b/c", "foo")
	fmt.Println(kv.Key())
	// Output: /a/b/c
}

func ExampleKV_KeyNoLeadingSlash() {
	kv, _ := NewKV("/a/b/c", "foo")
	fmt.Println(kv.KeyNoLeadingSlash())
	// Output: a/b/c
}

func ExampleKV_Val() {
	kv, _ := NewKV("/a/b/c", "foo")
	fmt.Println(kv.Val())
	// Output: foo
}
