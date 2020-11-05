package output

// import (
// 	"bytes"
// 	"context"
// 	"fmt"
// 	"io"
// 	"io/ioutil"
// 	"net/http"
// 	"strings"
// 	"time"

// 	"next-stage.com.cn/maya/pip"
// 	"next-stage.com.cn/maya/pip/serializers"
// )

// const (
// 	defaultURL = "http://127.0.0.1:8080/telegraf"
// )

// var sampleConfig = `
//   ## URL is the address to send metrics to
//   url = "http://127.0.0.1:8080/telegraf"

//   ## Timeout for HTTP message
//   # timeout = "5s"

//   ## HTTP method, one of: "POST" or "PUT"
//   # method = "POST"

//   ## HTTP Basic Auth credentials
//   # username = "username"
//   # password = "pa$$word"

//   ## OAuth2 Client Credentials Grant
//   # client_id = "clientid"
//   # client_secret = "secret"
//   # token_url = "https://indentityprovider/oauth2/v1/token"
//   # scopes = ["urn:opc:idm:__myscopes__"]

//   ## Optional TLS Config
//   # tls_ca = "/etc/telegraf/ca.pem"
//   # tls_cert = "/etc/telegraf/cert.pem"
//   # tls_key = "/etc/telegraf/key.pem"
//   ## Use TLS but skip chain & host verification
//   # insecure_skip_verify = false

//   ## Data format to output.
//   ## Each data format has it's own unique set of configuration options, read
//   ## more about them here:
//   ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
//   # data_format = "influx"

//   ## HTTP Content-Encoding for write request body, can be set to "gzip" to
//   ## compress body or "identity" to apply no encoding.
//   # content_encoding = "identity"

//   ## Additional HTTP headers
//   # [outputs.http.headers]
//   #   # Should be set manually to "application/json" for json data_format
//   #   Content-Type = "text/plain; charset=utf-8"
// `

// const (
// 	defaultClientTimeout = 5 * time.Second
// 	defaultContentType   = "text/plain; charset=utf-8"
// 	defaultMethod        = http.MethodPost
// )

// type HTTP struct {
// 	URL             string            `toml:"url"`
// 	Method          string            `toml:"method"`
// 	Username        string            `toml:"username"`
// 	Password        string            `toml:"password"`
// 	Headers         map[string]string `toml:"headers"`
// 	ClientID        string            `toml:"client_id"`
// 	ClientSecret    string            `toml:"client_secret"`
// 	TokenURL        string            `toml:"token_url"`
// 	Scopes          []string          `toml:"scopes"`
// 	ContentEncoding string            `toml:"content_encoding"`

// 	client     *http.Client
// 	serializer serializers.Serializer
// }

// func (h *HTTP) SetSerializer(serializer serializers.Serializer) {
// 	h.serializer = serializer
// }

// func (h *HTTP) createClient(ctx context.Context) (*http.Client, error) {

// 	client := &http.Client{
// 		Transport: &http.Transport{
// 			Proxy: http.ProxyFromEnvironment,
// 		},
// 	}

// 	return client, nil
// }

// func (h *HTTP) Connect() error {
// 	if h.Method == "" {
// 		h.Method = http.MethodPost
// 	}
// 	h.Method = strings.ToUpper(h.Method)
// 	if h.Method != http.MethodPost && h.Method != http.MethodPut {
// 		return fmt.Errorf("invalid method [%s] %s", h.URL, h.Method)
// 	}

// 	ctx := context.Background()
// 	client, err := h.createClient(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	h.client = client

// 	return nil
// }

// func (h *HTTP) Close() error {
// 	return nil
// }

// func (h *HTTP) Description() string {
// 	return "A plugin that can transmit metrics over HTTP"
// }

// func (h *HTTP) SampleConfig() string {
// 	return sampleConfig
// }

// func (h *HTTP) Write(metrics []pip.Metric) error {
// 	reqBody, err := h.serializer.SerializeBatch(metrics)
// 	if err != nil {
// 		return err
// 	}

// 	if err := h.write(reqBody); err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (h *HTTP) write(reqBody []byte) error {
// 	var reqBodyBuffer io.Reader = bytes.NewBuffer(reqBody)

// 	var err error

// 	req, err := http.NewRequest(h.Method, h.URL, reqBodyBuffer)
// 	if err != nil {
// 		return err
// 	}

// 	req.Header.Set("Content-Type", defaultContentType)

// 	resp, err := h.client.Do(req)
// 	if err != nil {
// 		return err
// 	}
// 	defer resp.Body.Close()
// 	_, err = ioutil.ReadAll(resp.Body)

// 	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
// 		return fmt.Errorf("when writing to [%s] received status code: %d", h.URL, resp.StatusCode)
// 	}

// 	return nil
// }

// func init() {
// 	Add("http", func() pip.Output {
// 		return &HTTP{
// 			Method: defaultMethod,
// 			URL:    defaultURL,
// 		}
// 	})
// }
