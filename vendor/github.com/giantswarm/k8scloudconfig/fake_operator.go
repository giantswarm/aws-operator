package cloudconfig

type FakeOperatorExtension struct{}

func (f *FakeOperatorExtension) Files() ([]FileAsset, error) {
	return nil, nil
}

func (f *FakeOperatorExtension) Units() ([]UnitAsset, error) {
	return nil, nil
}
