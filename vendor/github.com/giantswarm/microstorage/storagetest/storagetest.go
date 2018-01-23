package storagetest

import (
	"context"
	"fmt"
	"path"
	"sort"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/giantswarm/microstorage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test is Storage conformance test.
func Test(t *testing.T, storage microstorage.Storage) {
	testBasicCRUD(t, storage)
	testPutIdempotent(t, storage)
	testDeleteNotExisting(t, storage)
	testList(t, storage)
	testListEmpty(t, storage)
	testListNested(t, storage)
	testListInvalid(t, storage)
}

func testBasicCRUD(t *testing.T, storage microstorage.Storage) {
	var (
		name = "testBasicCRUD"

		ctx = context.TODO()

		baseKey = name + "-key"
		value   = name + "-value"
	)

	for _, key := range validKeyVariations(baseKey) {
		kv := microstorage.MustKV(microstorage.NewKV(key, value))

		ok, err := storage.Exists(ctx, kv.K())
		require.NoError(t, err, "%s: kv=%#v", name, kv)
		require.False(t, ok, "%s: kv=%#v", name, kv)

		_, err = storage.Search(ctx, kv.K())
		require.NotNil(t, err, "%s: kv=%#v", name, kv)
		require.True(t, microstorage.IsNotFound(err), "%s: key=%s expected IsNotFoundError", name, kv.Key)

		err = storage.Put(ctx, kv)

		ok, err = storage.Exists(ctx, kv.K())
		require.NoError(t, err, "%s: kv=%#v", name, kv)
		require.True(t, ok, "%s: kv=%#v", name, kv)

		gotKV, err := storage.Search(ctx, kv.K())
		require.NoError(t, err, "%s: kv=%#v", name, kv)
		require.Equal(t, kv, gotKV, "%s: kv=%#v", name, kv)

		err = storage.Delete(ctx, kv.K())
		require.NoError(t, err, "%s: kv=%#v", name, kv)

		ok, err = storage.Exists(ctx, kv.K())
		require.NoError(t, err, "%s: kv=%#v", name, kv)
		require.False(t, ok, "%s: kv=%#v", name, kv)

		_, err = storage.Search(ctx, kv.K())
		require.NotNil(t, err, "%s: kv=%#v", name, kv)
		require.True(t, microstorage.IsNotFound(err), "%s: key=%s expected IsNotFoundError", name, kv.Key)
	}
}

func testPutIdempotent(t *testing.T, storage microstorage.Storage) {
	var (
		name = "testPutIdempotent"

		ctx = context.TODO()

		baseKey        = name + "-key"
		value          = name + "-value"
		overridenValue = name + "-overriden-value"
	)

	for _, key := range validKeyVariations(baseKey) {
		kv := microstorage.MustKV(microstorage.NewKV(key, value))

		ok, err := storage.Exists(ctx, kv.K())
		require.NoError(t, err, "%s: kv=%#v", name, kv)
		require.False(t, ok, "%s: kv=%#v", name, kv)

		// First Put call.

		err = storage.Put(ctx, kv)
		require.NoError(t, err, "%s: kv=%#v", name, kv)

		gotKV, err := storage.Search(ctx, kv.K())
		require.NoError(t, err, "%s: kv=%#v", name, kv)
		require.Equal(t, kv, gotKV, "%s: kv=%#v", name, kv)

		// Second Put call with the same value.

		err = storage.Put(ctx, kv)
		require.NoError(t, err, "%s: kv=%#v", name, kv)

		gotKV, err = storage.Search(ctx, kv.K())
		require.NoError(t, err, "%s: kv=%#v", name, kv)
		require.Equal(t, kv, gotKV, "%s: kv=%#v", name, kv)

		// Third Put call with overriding value.

		overridenKV := microstorage.MustKV(microstorage.NewKV(kv.Key(), overridenValue))
		err = storage.Put(ctx, overridenKV)
		require.NoError(t, err, "%s: kv=%#v", name, overridenKV)

		gotKV, err = storage.Search(ctx, kv.K())
		require.NoError(t, err, "%s: kv=%#v", name, overridenKV)
		require.Equal(t, overridenKV, gotKV, "%s: kv=%#s", name, overridenKV)
	}
}

func testDeleteNotExisting(t *testing.T, storage microstorage.Storage) {
	var (
		name = "testDeleteNotExisting"

		ctx = context.TODO()

		baseKey = name + "-key"
	)

	for _, key := range validKeyVariations(baseKey) {
		k := microstorage.MustK(microstorage.NewK(key))
		err := storage.Delete(ctx, k)
		require.NoError(t, err, name, "key=%s", k.Key())
	}
}

func testList(t *testing.T, storage microstorage.Storage) {
	var (
		name = "testList"

		ctx = context.TODO()

		baseKey = name + "-key"
		value   = name + "-value"
	)

	for _, key0 := range validKeyVariations(baseKey) {
		kv0 := microstorage.MustKV(microstorage.NewKV(key0, value))
		kv1 := microstorage.MustKV(microstorage.NewKV(path.Join(key0, "one"), value))
		kv2 := microstorage.MustKV(microstorage.NewKV(path.Join(key0, "two"), value))

		err := storage.Put(ctx, kv1)
		assert.Nil(t, err, "%s: key=%s", name, kv1.Key())

		err = storage.Put(ctx, kv2)
		assert.Nil(t, err, "%s: key=%s", name, kv2.Key())

		kvs := []microstorage.KV{
			microstorage.MustKV(microstorage.NewKV("one", value)),
			microstorage.MustKV(microstorage.NewKV("two", value)),
		}
		sort.Sort(kvSlice(kvs))

		gotKVs, err := storage.List(ctx, kv0.K())
		assert.NoError(t, err, "%s: key=%s", name, kv0.Key())
		sort.Sort(kvSlice(gotKVs))
		assert.Equal(t, kvs, gotKVs, "%s: key=%s", name, kv0.Key())
	}
}

func testListEmpty(t *testing.T, storage microstorage.Storage) {
	var (
		name = "testListEmpty"

		ctx = context.TODO()

		baseKey = name + "-key"
	)

	for _, key := range validKeyVariations(baseKey) {
		gotKVs, err := storage.List(ctx, microstorage.MustK(microstorage.NewK(key)))
		assert.NoError(t, err, "%s: key=%#v", name, microstorage.RootKey.Key())

		// Make sure empty list is returned when listing non-existing
		// key.
		assert.Empty(t, gotKVs, "%s: kvs=%#v", name, gotKVs)
	}
}

func testListNested(t *testing.T, storage microstorage.Storage) {
	var (
		name = "testListNested"

		ctx = context.TODO()

		baseKey = name + "-key"
		value   = name + "-value"
	)

	for _, key := range validKeyVariations(baseKey) {
		kvs := []microstorage.KV{
			microstorage.MustKV(microstorage.NewKV(path.Join(key, "nested/one"), value)),
			microstorage.MustKV(microstorage.NewKV(path.Join(key, "nested/two"), value)),
			microstorage.MustKV(microstorage.NewKV(path.Join(key, "extremaly/nested/two"), value)),
		}

		for _, kv := range kvs {
			err := storage.Put(ctx, kv)
			assert.Nil(t, err, "%s: kv=%#v", name, kv.Key())
		}

		gotKVs, err := storage.List(ctx, microstorage.RootKey)
		assert.NoError(t, err, "%s: key=%#v", name, microstorage.RootKey.Key())

		for _, kv := range kvs {
			assert.Contains(t, gotKVs, kv, "%s: key=%#v", name, microstorage.RootKey.Key())
		}
	}
}

func testListInvalid(t *testing.T, storage microstorage.Storage) {
	var (
		name = "testListInvalid"

		ctx = context.TODO()

		baseKey = name + "-key"
		value   = name + "-value"
	)

	for _, key0 := range validKeyVariations(baseKey) {
		k0 := microstorage.MustK(microstorage.NewK(baseKey))
		kv1 := microstorage.MustKV(microstorage.NewKV(path.Join(key0, "one"), value))
		kv2 := microstorage.MustKV(microstorage.NewKV(path.Join(key0, "two"), value))

		err := storage.Put(ctx, kv1)
		assert.Nil(t, err, "%s: kv=%#v", name, kv1)

		err = storage.Put(ctx, kv2)
		assert.Nil(t, err, "%s: kv=%#v", name, kv2)

		// baseKey is key0 prefix.
		//
		// We have keys like:
		//
		// - /testListInvalid-key-XXXX/one
		// - /testListInvalid-key-XXXX/two
		//
		// Listing /testListInvalid-key should fail.
		list, err := storage.List(ctx, k0)
		assert.NoError(t, err, "%s: key=%#v", name, k0)
		assert.Empty(t, list, "%s: key=%#v", name, k0)
	}
}

var validKeyVariationsIDGen int64

func validKeyVariations(key string) []string {
	if strings.HasPrefix(key, "/") {
		key = key[1:]
	}
	if strings.HasSuffix(key, "/") {
		key = key[:len(key)-1]
	}

	next := func() string {
		return fmt.Sprintf("%s-%04d", key, atomic.AddInt64(&validKeyVariationsIDGen, 1))
	}

	return []string{
		next(),
		"/" + next(),
		next() + "/",
		"/" + next() + "/",
	}
}

type kvSlice []microstorage.KV

func (p kvSlice) Len() int           { return len(p) }
func (p kvSlice) Less(i, j int) bool { return p[i].Key() < p[j].Key() }
func (p kvSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
