package s3object

type BucketObjectState struct {
	Bucket string
	Body   string
	Hash   string
	Key    string
}
