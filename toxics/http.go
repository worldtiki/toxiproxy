package toxics

import (
	"bufio"
	"bytes"
	"io"
	"net/http"

	"github.com/Shopify/toxiproxy/stream"
)

type HttpToxic struct {
	// Times in milliseconds
	Host string `json:"host"`
}

func (t *HttpToxic) ModifyRequest(resp *http.Request) {
	resp.Host = t.Host
}

func (t *HttpToxic) Pipe(stub *ToxicStub) {
	buffer := bytes.NewBuffer(make([]byte, 0, 32*1024))
	writer := stream.NewChanWriter(stub.Output)
	reader := stream.NewChanReader(stub.Input)
	reader.SetInterrupt(stub.Interrupt)
	for {
		tee := io.TeeReader(reader, buffer)
		req, err := http.ReadRequest(bufio.NewReader(tee))
		if err == stream.ErrInterrupted {
			buffer.WriteTo(writer)
			return
		} else if err == io.EOF {
			stub.Close()
			return
		}
		if err != nil {
			buffer.WriteTo(writer)
		} else {
			t.ModifyRequest(req)
			req.Write(writer)
		}
		buffer.Reset()
	}
}

func init() {
	Register("http", new(HttpToxic))
}
