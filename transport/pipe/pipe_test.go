package pipe_test

import (
	"context"
	"io"
	"testing"
	"time"

	"v2ray.com/core/common/buf"
	"v2ray.com/core/common/signal"
	. "v2ray.com/core/transport/pipe"
	. "v2ray.com/ext/assert"
)

func TestPipeReadWrite(t *testing.T) {
	assert := With(t)

	pReader, pWriter := New()
	payload := []byte{'a', 'b', 'c', 'd'}
	b := buf.New()
	b.Write(payload)
	assert(pWriter.WriteMultiBuffer(buf.NewMultiBufferValue(b)), IsNil)

	rb, err := pReader.ReadMultiBuffer()
	assert(err, IsNil)
	assert(rb.String(), Equals, b.String())
}

func TestPipeCloseError(t *testing.T) {
	assert := With(t)

	pReader, pWriter := New()
	payload := []byte{'a', 'b', 'c', 'd'}
	b := buf.New()
	b.Write(payload)
	assert(pWriter.WriteMultiBuffer(buf.NewMultiBufferValue(b)), IsNil)
	pWriter.CloseError()

	rb, err := pReader.ReadMultiBuffer()
	assert(err, Equals, io.ErrClosedPipe)
	assert(rb.IsEmpty(), IsTrue)
}

func TestPipeClose(t *testing.T) {
	assert := With(t)

	pReader, pWriter := New()
	payload := []byte{'a', 'b', 'c', 'd'}
	b := buf.New()
	b.Write(payload)
	assert(pWriter.WriteMultiBuffer(buf.NewMultiBufferValue(b)), IsNil)
	assert(pWriter.Close(), IsNil)

	rb, err := pReader.ReadMultiBuffer()
	assert(err, IsNil)
	assert(rb.String(), Equals, b.String())

	rb, err = pReader.ReadMultiBuffer()
	assert(err, Equals, io.EOF)
	assert(rb.IsEmpty(), IsTrue)
}

func TestPipeLimitZero(t *testing.T) {
	assert := With(t)

	pReader, pWriter := New(WithSizeLimit(0))
	bb := buf.New()
	bb.Write([]byte{'a', 'b'})
	assert(pWriter.WriteMultiBuffer(buf.NewMultiBufferValue(bb)), IsNil)

	err := signal.ExecuteParallel(context.Background(), func() error {
		b := buf.New()
		b.Write([]byte{'c', 'd'})
		return pWriter.WriteMultiBuffer(buf.NewMultiBufferValue(b))
	}, func() error {
		time.Sleep(time.Second)

		rb, err := pReader.ReadMultiBuffer()
		if err != nil {
			return err
		}
		assert(rb.String(), Equals, "ab")

		rb, err = pReader.ReadMultiBuffer()
		if err != nil {
			return err
		}
		assert(rb.String(), Equals, "cd")
		return nil
	})

	assert(err, IsNil)
}
