package http

import (
	http "github.com/pkopriv2/golang-sdk/http/client"
	"github.com/pkopriv2/golang-sdk/http/headers"
	"github.com/pkopriv2/golang-sdk/lang/enc"
	"github.com/pkopriv2/services-catalog/core"
)

type Client struct {
	Raw http.Client
	Enc enc.Encoder
}

func NewClient(raw http.Client, enc enc.Encoder) core.Transport {
	return &Client{raw, enc}
}

func (c *Client) SaveService(svc core.Service) (ret core.Service, err error) {
	err = c.Raw.Call(
		http.BuildRequest(
			http.Put("/v1/services"),
			http.WithHeader(headers.Accept, c.Enc.Mime()),
			http.WithStruct(c.Enc, svc)),
		http.ExpectAll(
			http.ExpectCode(200),
			http.ExpectStruct(enc.DefaultRegistry, &ret)))
	return
}

func (c *Client) SaveVersion(ver core.Version) (ret core.Version, err error) {
	err = c.Raw.Call(
		http.BuildRequest(
			http.Put("/v1/versions"),
			http.WithHeader(headers.Accept, c.Enc.Mime()),
			http.WithStruct(c.Enc, ver)),
		http.ExpectAll(
			http.ExpectCode(200),
			http.ExpectStruct(enc.DefaultRegistry, &ret)))
	return
}

func (c *Client) ListServices(filter core.Filter, page core.Page) (ret core.Catalog, err error) {
	err = c.Raw.Call(
		http.BuildRequest(
			http.Get("/v1/services"),
			http.WithHeader(headers.Accept, c.Enc.Mime()),
			http.WithQueryParam("name", filter.NameContains),
			http.WithQueryParam("desc", filter.DescContains),
			http.WithQueryParam("id", filter.ServiceId),
			http.WithQueryParam("offset", page.Offset),
			http.WithQueryParam("limit", page.Limit),
			http.WithQueryParam("order", page.OrderBy)),
		http.ExpectAll(
			http.ExpectCode(200),
			http.ExpectStruct(enc.DefaultRegistry, &ret)))
	return
}
