package key

func IsDeleted(getter DeletionTimestampGetter) bool {
	return getter.GetDeletionTimestamp() != nil
}
