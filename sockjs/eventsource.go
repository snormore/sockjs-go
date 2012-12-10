package sockjs

import (
	"bytes"
	"code.google.com/p/gorilla/mux"
	"fmt"
	"io"
	"net/http"
)

type eventSourceProtocol struct{}

func (this *context) EventSourceHandler(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	sessid := vars["sessionid"]

	httpTx := &httpTransaction{
		protocolHelper: eventSourceProtocol{},
		req:            req,
		rw:             rw,
		sessionId:      sessid,
		done:           make(chan bool),
	}
	this.baseHandler(httpTx)
}

func (eventSourceProtocol) isStreaming() bool   { return true }
func (eventSourceProtocol) contentType() string { return "text/event-stream; charset=UTF-8" }

func (eventSourceProtocol) writeOpenFrame(w io.Writer) (int, error) {
	return fmt.Fprintf(w, "data: o\r\n\r\n")
}
func (eventSourceProtocol) writeHeartbeat(w io.Writer) (int, error) {
	return fmt.Fprintln(w, "data: h\r\n\r\n")
}
func (eventSourceProtocol) writePrelude(w io.Writer) (int, error) {
	return fmt.Fprintf(w, "\r\n")
}
func (eventSourceProtocol) writeClose(w io.Writer, code int, msg string) (int, error) {
	return fmt.Fprintf(w, "data: c[%d,\"%s\"]\r\n\r\n", code, msg)
}
func (eventSourceProtocol) writeData(w io.Writer, frames ...[]byte) (int, error) {
	b := &bytes.Buffer{}
	fmt.Fprintf(b, "data: a[")
	for n, frame := range frames {
		if n > 0 {
			b.Write([]byte(","))
		}
		sesc := re.ReplaceAllFunc(frame, func(s []byte) []byte {
			return []byte(fmt.Sprintf(`\u%04x`, []rune(string(s))[0]))
		})
		b.Write(sesc)
	}
	fmt.Fprintf(b, "]\r\n\r\n")
	n, err := b.WriteTo(w)
	return int(n), err
}
