package v_2_0_2

type FakeExtension struct{}

func (f *FakeExtension) Files() ([]FileAsset, error) {
	return nil, nil
}

func (f *FakeExtension) Units() ([]UnitAsset, error) {
	return nil, nil
}

func (f *FakeExtension) VerbatimSections() []VerbatimSection {
	return nil
}
