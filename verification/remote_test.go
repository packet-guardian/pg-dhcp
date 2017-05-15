package verification

import (
	"bytes"
	"net"
	"testing"
	"time"
)

type remoteTester struct {
	nextRequest  []byte
	nextResponse []byte
	t            *testing.T
}

func (r *remoteTester) Read(b []byte) (n int, err error) {
	copy(b, r.nextResponse)
	if len(b) > len(r.nextResponse) {
		return len(r.nextResponse), nil
	}
	return len(b), nil
}

func (r *remoteTester) Write(b []byte) (n int, err error) {
	if !bytes.Equal(b, r.nextRequest) {
		r.t.Errorf("expected %#v, got %#v", r.nextRequest, b)
	}
	return len(b), nil
}

func (r *remoteTester) Close() error                       { return nil }
func (r *remoteTester) LocalAddr() net.Addr                { return nil }
func (r *remoteTester) RemoteAddr() net.Addr               { return nil }
func (r *remoteTester) SetDeadline(t time.Time) error      { return nil }
func (r *remoteTester) SetReadDeadline(t time.Time) error  { return nil }
func (r *remoteTester) SetWriteDeadline(t time.Time) error { return nil }
func (r *remoteTester) setNextRequest(req []byte)          { r.nextRequest = req }
func (r *remoteTester) setNextResponse(resp []byte)        { r.nextResponse = resp }

func TestRemoteVerifier(t *testing.T) {
	mock := &remoteTester{t: t}
	mock.setNextRequest([]byte{1, 'V', 0x12, 0x34, 0x56, 0xab, 0xcd, 0xef})
	mock.setNextResponse([]byte{1, 'R', 'R'})

	mac, _ := net.ParseMAC("12:34:56:ab:cd:ef")

	r := &RemoteVerifier{c: mock}
	resp, err := r.VerifyClient(mac)
	if err != nil {
		t.Errorf("got error: %s", err.Error())
	}

	if resp != ClientRegistered {
		t.Errorf("expected %d, got %d", ClientRegistered, resp)
	}

	mock.setNextRequest([]byte{1, 'V', 0x12, 0x34, 0x56, 0xab, 0xcd, 0xef})
	mock.setNextResponse([]byte{1, 'R', 'U'})

	r = &RemoteVerifier{c: mock}
	resp, err = r.VerifyClient(mac)
	if err != nil {
		t.Errorf("got error: %s", err.Error())
	}

	if resp != ClientUnregistered {
		t.Errorf("expected %d, got %d", ClientUnregistered, resp)
	}
}
