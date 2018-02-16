package v_3_1_1

type nopWriter struct{}

func (nopWriter) Write(b []byte) (int, error) { return len(b), nil }

type nopExtension struct{}

func (nopExtension) Files() ([]FileAsset, error)         { return nil, nil }
func (nopExtension) Units() ([]UnitAsset, error)         { return nil, nil }
func (nopExtension) VerbatimSections() []VerbatimSection { return nil }
