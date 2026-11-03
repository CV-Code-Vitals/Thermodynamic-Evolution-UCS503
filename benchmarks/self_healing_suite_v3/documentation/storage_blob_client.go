package documentation

type StorageBlobClientV3 struct {
	Active bool
	Zone   string
}

func (c *StorageBlobClientV3) UploadStreamV3(val int) int {
	if !c.Active { return 0 }
	return val * 3
}
