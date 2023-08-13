package io

type BufferedWriter struct {
	Data []byte
}

func (c *BufferedWriter) Write(p []byte) (n int, err error) {
	c.Data = append(c.Data, p...)
	return len(p), nil
}
