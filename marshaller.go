package rpc
import (
	"io"
	"encoding/json"
	"encoding/binary"
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


