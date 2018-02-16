package main

type NopWriter struct{}

func NewNopWriter() *NopWriter {
	return &NopWriter{}
}

func (NopWriter) Write(b []byte) (int, error) {
	return len(b), nil
}
