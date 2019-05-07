package httpserver

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/http/httpguts"

	"github.com/chenlx0/echoes/internal/util/balance"

	"github.com/chenlx0/echoes/internal/version"

	"github.com/chenlx0/echoes/internal/config"
	"github.com/chenlx0/echoes/internal/lg"
)

// HTTP connection constants
const (
	KeepAliveTimeout = 30 * time.Second
	MaxIdleConns     = 1000
	IdleConnTimeout  = 90 * time.Second
	TLSTimeout       = 10 * time.Second
	BufferSize       = 1024
)

var hopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te",      // canonicalized version of "TE"
	"Trailer", // not Trailers per URL above; https://www.rfc-editor.org/errata_search.php?eid=4522
	"Transfer-Encoding",
	"Upgrade",
}

var additionHeaders = map[string]string{
	"server": "echoes/" + version.Binary,
}

// PhaseFilter filt an EchoHandler
// There are 3 phases in an EchoHandler
// rewritePhase: run before proxy pass excute, modify url, upstream... etc here
// respFilter: run after proxy pass excuted, modify response headers, body here
// logPhase: run after the whole http request end, it would be called in a goroutine.
type PhaseFilter func(*EchoHandler) error

// EchoHandler implements http handler interface
type EchoHandler struct {
	globalConf *config.GlobalConfig
	HostConf   *config.VHost
	Host       string
	Logf       lg.AppLogFunc

	// Correspond upstream
	Upstream *config.Upstream
	URL      *url.URL

	ReqIn *http.Request
	ctx   context.Context

	writer     http.ResponseWriter
	RespIn     *http.Response
	HeadersOut http.Header

	rewritePhase PhaseFilter
	respFilter   PhaseFilter
	logPhase     PhaseFilter

	respFinished bool
}

func (e EchoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// initialize basic vars
	// search corresponding config
	e.Host = r.Host
	e.Logf(lg.INFO, "Host name: ", e.Host)
	for _, v := range e.globalConf.VHosts {
		for _, vv := range v.ServerName {
			if e.Host == vv {
				e.HostConf = &v
			}
		}
	}

	// context config
	e.ctx = r.Context()
	if cn, ok := w.(http.CloseNotifier); ok {
		var cancel context.CancelFunc
		e.ctx, cancel = context.WithCancel(e.ctx)
		defer cancel()
		notifyChan := cn.CloseNotify()
		go func() {
			select {
			case <-notifyChan:
				e.Logf(lg.WARN, "Response cancel")
				cancel()
			case <-e.ctx.Done():

			}
		}()
	}

	e.ReqIn = r
	e.writer = w

	// rewrite phase
	if e.rewritePhase != nil {
		err := e.rewritePhase(&e)
		if err != nil {
			e.Logf(lg.ERROR, "Rewrite Phase: ", err)
			e.Abort(500, "Internal Server Error")
		}
	}
	e.exec()
	defer e.RespIn.Body.Close()

	// response filter
	if e.respFilter != nil {
		err := e.respFilter(&e)
		if err != nil {
			e.Logf(lg.ERROR, "Response Filter: ", err)
			e.Abort(500, "Internal Server Error")
		}
	}

	modifyHeaders(e.RespIn.Header)
	copyHeader(w.Header(), e.RespIn.Header)
	w.WriteHeader(e.RespIn.StatusCode)
	err := copyBody(w, e.RespIn.Body)
	if err != nil {
		e.Abort(500, err)
	}

	if e.logPhase != nil {
		go e.logPhase(&e)
	}
}

// Abort http request
func (e *EchoHandler) Abort(status int, msg interface{}) {
	content := fmt.Sprintf(`{"status":"%d","msg":"%s"}`, status, msg)
	e.RespIn = &http.Response{
		StatusCode: status,
		Body:       ioutil.NopCloser(bytes.NewBuffer([]byte(content))),
	}
}

func (e *EchoHandler) retrieveUpstream() {
	balanceFunc := balance.GetLoadBalanceFunc(e.HostConf.LoadBlance)
	e.Upstream = balanceFunc(e.HostConf.Upstreams)
}

func (e *EchoHandler) retrieveURL() {
	if e.Upstream == nil {
		e.retrieveUpstream()
	}
	originURL := e.ReqIn.URL
	host := fmt.Sprintf("%s:%d", e.Upstream.Host, e.Upstream.Port)
	e.URL = &url.URL{
		Scheme:   e.Upstream.Scheme,
		Path:     originURL.Path,
		RawPath:  originURL.RawPath,
		RawQuery: originURL.RawQuery,
		Host:     host,
		Fragment: originURL.Fragment,
	}
}

func (e *EchoHandler) exec() {
	transport := http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   time.Duration(e.HostConf.MaxTimeout) * time.Second,
			KeepAlive: KeepAliveTimeout,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          MaxIdleConns,
		IdleConnTimeout:       IdleConnTimeout,
		TLSHandshakeTimeout:   TLSTimeout,
		ExpectContinueTimeout: time.Second,
	}
	e.retrieveURL()

	// init and execute proxy pass request
	reqOut := e.ReqIn.WithContext(e.ctx)
	reqOut.URL = e.URL
	reqOut.Host = e.Upstream.Host
	var err error
	e.RespIn, err = transport.RoundTrip(reqOut)
	if err != nil {
		e.Logf(lg.ERROR, "Proxy pass response: ", err)
		e.Abort(500, err)
	}
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func copyBody(dst io.Writer, src io.Reader) error {
	buf := make([]byte, BufferSize)
	_, err := io.CopyBuffer(dst, src, buf)
	return err
}

func modifyHeaders(headers http.Header) {
	// Remove hopHeaders
	for _, h := range hopHeaders {
		hv := headers.Get(h)
		if hv == "" {
			continue
		}
		headers.Del(hv)
	}

	// Add addtional heders
	for k, h := range additionHeaders {
		headers.Set(k, h)
	}

	// After stripping hop-by-hop headers, add back neccesary headers
	// for protocol upgrades, such as websockets
}

func upgradeType(h http.Header) string {
	if !httpguts.HeaderValuesContainsToken(h["Connection"], "Upgrade") {
		return ""
	}
	return strings.ToLower(h.Get("Upgrade"))
}
