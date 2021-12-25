package gproxy

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/csby/tcpproxy"
	"io"
	"log"
	"net"
	"time"
)

type TargetProxy struct {
	// Addr is the TCP address to proxy to.
	//Addr    string
	Address TargetAddress

	// KeepAlivePeriod sets the period between TCP keep alives.
	// If zero, a default is used. To disable, use a negative number.
	// The keep-alive is used for both the client connection and
	KeepAlivePeriod time.Duration

	// DialTimeout optionally specifies a dial timeout.
	// If zero, a default is used.
	// If negative, the timeout is disabled.
	DialTimeout time.Duration

	// DialContext optionally specifies an alternate dial function
	// for TCP targets. If nil, the standard
	// net.Dialer.DialContext method is used.
	DialContext func(ctx context.Context, network, address string) (net.Conn, error)

	// OnDialError optionally specifies an alternate way to handle errors dialing Addr.
	// If nil, the error is logged and src is closed.
	// If non-nil, src is not closed automatically.
	OnDialError func(src net.Conn, dstAddr string, dstDialErr error)

	// ProxyProtocolVersion optionally specifies the version of
	// HAProxy's PROXY protocol to use. The PROXY protocol provides
	// connection metadata to the DialProxy target, via a header
	// inserted ahead of the client's traffic. The DialProxy target
	// must explicitly support and expect the PROXY header; there is
	// no graceful downgrade.
	// If zero, no PROXY header is sent. Currently, version 1 is supported.
	ProxyProtocolVersion int

	OnConnected    func(id, addr, domain, source, target string)
	OnDisconnected func(id, addr, domain, source, target string)
}

// HandleConn implements the Target interface.
func (dp *TargetProxy) HandleConn(src net.Conn, listenAddress, hostName string) {
	ctx := context.Background()
	var cancel context.CancelFunc
	if dp.DialTimeout >= 0 {
		ctx, cancel = context.WithTimeout(ctx, dp.dialTimeout())
	}
	addr := dp.Address.GetAddress()
	addr.IncreaseCount()
	defer addr.DecreaseCount()

	dst, err := dp.dialContext()(ctx, "tcp", addr.Addr)
	if cancel != nil {
		cancel()
	}
	if err != nil {
		dp.onDialError()(src, addr.Addr, err)
		return
	}
	defer goCloseConn(dst)

	if err = dp.sendProxyHeader(dst, src); err != nil {
		dp.onDialError()(src, addr.Addr, err)
		return
	}
	defer goCloseConn(src)

	if ka := dp.keepAlivePeriod(); ka > 0 {
		if c, ok := UnderlyingConn(src).(*net.TCPConn); ok {
			c.SetKeepAlive(true)
			c.SetKeepAlivePeriod(ka)
		}
		if c, ok := dst.(*net.TCPConn); ok {
			c.SetKeepAlive(true)
			c.SetKeepAlivePeriod(ka)
		}
	}

	id := newGuid()
	srcAddr := src.RemoteAddr().String()
	dstAddr := dst.RemoteAddr().String()
	if dp.OnConnected != nil {
		go dp.OnConnected(id, listenAddress, hostName, srcAddr, dstAddr)
	}

	ec := make(chan error, 1)
	go proxyCopy(ec, src, dst)
	go proxyCopy(ec, dst, src)
	<-ec

	if dp.OnDisconnected != nil {
		go dp.OnDisconnected(id, listenAddress, hostName, srcAddr, dstAddr)
	}
}

func (dp *TargetProxy) dialTimeout() time.Duration {
	if dp.DialTimeout > 0 {
		return dp.DialTimeout
	}
	return 10 * time.Second
}

var defaultDialer = new(net.Dialer)

func (dp *TargetProxy) dialContext() func(ctx context.Context, network, address string) (net.Conn, error) {
	if dp.DialContext != nil {
		return dp.DialContext
	}
	return defaultDialer.DialContext
}

func (dp *TargetProxy) onDialError() func(src net.Conn, dstAddr string, dstDialErr error) {
	if dp.OnDialError != nil {
		return dp.OnDialError
	}
	return func(src net.Conn, dstAddr string, dstDialErr error) {
		log.Printf("tcpproxy: for incoming conn %v, error dialing %q: %v", src.RemoteAddr().String(), dstAddr, dstDialErr)
		src.Close()
	}
}

func goCloseConn(c net.Conn) { go c.Close() }

func (dp *TargetProxy) sendProxyHeader(w io.Writer, src net.Conn) error {
	switch dp.ProxyProtocolVersion {
	case 0:
		return nil
	case 1:
		var srcAddr, dstAddr *net.TCPAddr
		if a, ok := src.RemoteAddr().(*net.TCPAddr); ok {
			srcAddr = a
		}
		if a, ok := src.LocalAddr().(*net.TCPAddr); ok {
			dstAddr = a
		}

		if srcAddr == nil || dstAddr == nil {
			_, err := io.WriteString(w, "PROXY UNKNOWN\r\n")
			return err
		}

		family := "TCP4"
		if srcAddr.IP.To4() == nil {
			family = "TCP6"
		}
		_, err := fmt.Fprintf(w, "PROXY %s %s %d %s %d\r\n", family, srcAddr.IP, srcAddr.Port, dstAddr.IP, dstAddr.Port)
		return err
	default:
		return fmt.Errorf("PROXY protocol version %d not supported", dp.ProxyProtocolVersion)
	}
}

func (dp *TargetProxy) keepAlivePeriod() time.Duration {
	if dp.KeepAlivePeriod != 0 {
		return dp.KeepAlivePeriod
	}
	return time.Minute
}

// UnderlyingConn returns c.Conn if c of type *Conn,
// otherwise it returns c.
func UnderlyingConn(c net.Conn) net.Conn {
	if wrap, ok := c.(*tcpproxy.Conn); ok {
		return wrap.Conn
	}
	return c
}

func newGuid() string {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return ""
	}

	uuid[8] = uuid[8]&^0xc0 | 0x80
	uuid[6] = uuid[6]&^0xf0 | 0x40

	return fmt.Sprintf("%x%x%x%x%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}

// proxyCopy is the function that copies bytes around.
// It's a named function instead of a func literal so users get
// named goroutines in debug goroutine stack dumps.
func proxyCopy(ec chan<- error, dst, src net.Conn) {
	// Before we unwrap src and/or dst, copy any buffered data.
	if wc, ok := src.(*tcpproxy.Conn); ok && len(wc.Peeked) > 0 {
		if _, err := dst.Write(wc.Peeked); err != nil {
			ec <- err
			return
		}
		wc.Peeked = nil
	}

	// Unwrap the src and dst from *Conn to *net.TCPConn so Go
	// 1.11's splice optimization kicks in.
	src = UnderlyingConn(src)
	dst = UnderlyingConn(dst)

	_, err := io.Copy(dst, src)
	ec <- err
}
