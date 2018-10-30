package s3object

const (
	prefixMaster = "master"
	prefixWorker = "worker"
)

type BucketObjectState struct {
	Bucket string
	Body   string
	Key    string
}
