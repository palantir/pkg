package safejson

func Decode(in, out interface{}) error {
	if out == nil {
		return nil // nothing to do
	}

	b, err := Marshal(in)
	if err != nil {
		return err
	}
	return Unmarshal(b, &out)
}
