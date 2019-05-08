package httpserver

import (
	"crypto/tls"
	"net/http"
	"os"

	"github.com/chenlx0/echoes/internal/config"

	"github.com/chenlx0/echoes/internal/lg"

	"github.com/chenlx0/echoes/internal/ssl"
)

func newTLSConfig(certsDir string) *tls.Config {
	// Initialize certificates
	domains := []string{"a.icug.net.cn"}
	certs, err := ssl.GetAllCertificates(domains, certsDir)
	if err != nil {
		panic(err)
	}
	c := &tls.Config{
		Certificates: certs,
		MaxVersion:   tls.VersionTLS13,
		NextProtos:   []string{"h2"},
	}
	return c
}

// Run config initialization and echoes http server
func Run(conf *config.GlobalConfig) {
	// init tls config and start listening
	tlsConfig := newTLSConfig(conf.CertsDir)
	ln, err := tls.Listen("tcp", ":443", tlsConfig)
	if err != nil {
		lg.LogFatal("init tls config: ", err)
		return
	}
	defer ln.Close()

	// log
	logFile, err := os.OpenFile("access.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		lg.LogFatal("open log file: ", err)
	}
	defer logFile.Close()
	logf := lg.GetAppLogFunc(logFile, lg.INFO)

	// init log handler and log conifg
	handler := EchoHandler{
		Logf:       logf,
		globalConf: conf,
	}

	// start http server
	if err = http.Serve(ln, handler); err != nil {
		lg.LogFatal("Serve http: ", err)
	}
}
