package client

import (
	"encoding/base64"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"strconv"
)

type Writer struct {
	writer io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer: w,
	}
}

func (w *Writer) WriteArray(values []interface{}) error {
	if values == nil {
		return w.write(typeArray, []byte("-1"), separator)
	}

	valuesSize := int64(len(values))

	if err := w.write(typeArray, []byte(strconv.FormatInt(valuesSize, 10)), separator); err != nil {
		return err
	}

	for _, v := range values {
		switch t := v.(type) {

		case string:
			if err := w.WriteBulkString([]byte(t)); err != nil {
				return err
			}
		case []byte:
			if err := w.WriteBulkString(t); err != nil {
				return err
			}
		case []interface{}:
			if err := w.WriteArray(t); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported type: the value [%#v] is not supported by this client, supported types are int8 to int64, strings, []byte, nil, and []interface{} of these same types", v)
		}
	}

	return nil
}

func (w *Writer) WriteBulkString(value []byte) error {
	return w.write(
		typeBulkString,
		[]byte(strconv.FormatInt(int64(len(value)), 10)),
		separator,
		value,
		separator,
	)
}

func (w *Writer) write(messageType byte, contents ...[]byte) error {
	if _, err := w.writer.Write([]byte{messageType}); err != nil {
		return errors.Wrapf(err, "failed to write message type: %v", messageType)
	}

	for _, b := range contents {
		if _, err := w.writer.Write(b); err != nil {
			return errors.Wrapf(err, "failed to write bytes, content in base64: [%v]", base64.RawStdEncoding.EncodeToString(b))
		}
	}

	return nil
}
