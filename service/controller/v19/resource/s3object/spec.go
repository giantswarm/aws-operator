package s3object

const (
	prefixWorker = "worker"
)

type BucketObjectState struct {
	Bucket string
	Body   string
	Key    string
}
