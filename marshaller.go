package rpc

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
)

type Marshaller struct {
}

func (m *Marshaller) Marshal(writer io.Writer, rpcType uint8, args interface{}) error {
	// write message type
	if err := binary.Write(writer, binary.LittleEndian, byte(0)); err != nil {
		return err
	}

	// marshal args
	bytes, err := json.Marshal(args)

	if err != nil {
		return err
	}

	// write args len as int32
	if err := binary.Write(writer, binary.LittleEndian, int32(len(bytes))); err != nil {
		return err
	}

	// write marshalled args
	if _, err = writer.Write(bytes); err != nil {
		return err
	}
	return nil

}

func (m *Marshaller) UnMarshalRequest(reader io.Reader) (interface{}, error) {
	var rpcType byte
	if err := binary.Read(reader, binary.LittleEndian, &rpcType); err != nil {
		return nil, err
	}

	var size int32
	if err := binary.Read(reader, binary.LittleEndian, &size); err != nil {
		return nil, err
	}

	buf := make([]byte, size)

	if _, err := reader.Read(buf); err != nil {
		return nil, err
	}
	req := &EchoRequest{}
	if err := json.Unmarshal(buf, &req); err != nil {
		return nil, err
	}
	return req, nil
}

func (m *Marshaller) UnMarshalResponse(reader io.Reader, resp interface{}) error {
	var isError byte
	if err := binary.Read(reader, binary.LittleEndian, &isError); err != nil {
		return err
	}

	var size int32
	if err := binary.Read(reader, binary.LittleEndian, &size); err != nil {
		return err
	}

	buf := make([]byte, size)

	if _, err := reader.Read(buf); err != nil {
		return err
	}

	if isError == 1 {
		var errorMessage string
		if err := json.Unmarshal(buf, &errorMessage); err != nil {
			return err
		}
		return errors.New(errorMessage)
	} else {
		if err := json.Unmarshal(buf, resp); err != nil {
			return err
		}
	}
	return nil
}
