package mika

import (
	"fmt"
	"io"

	"idcproxy/utils"
)

type header struct {
	ProtocolRelated []byte
	Addr            string
}

func newHeader(rawAddr []byte) *header {
	return &header{
		ProtocolRelated: rawAddr,
	}
}

func (h *header) Bytes() (hb []byte) {
	hb = append(hb, h.ProtocolRelated...)

	return
}

func getHeader(c io.Reader) (*header, error) {
	header := new(header)

	var err error
	header.ProtocolRelated, header.Addr, err = utils.GetAddress(c)
	if err != nil {
		return nil, fmt.Errorf("error mika %s", err)
	}

	return header, nil
}
