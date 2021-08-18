package gorequest

import "net/http"

type GoRequestContext struct {
	HTTPContext *http.Client

	MaxRedirect int

	Proxy string
	ProxyType string

	Timeout int
}

type DebuggerContext struct {
	Enable bool
	Data []string
}

type NewRequest struct {
	CookiesContext *CookiesContext
	DebuggerContext *DebuggerContext
	Header []HeaderData

	FollowLocation bool
	AdditionalHeader bool
	CookiesEnable bool
	
	Referer string
	Accept string
	RequestRAW string
	
	RequestLastUrl string

	Method string
	Body RequestBody
}
type RequestBody struct {
	Status bool
	Type string
	FormData map[string]string
	Data string
	Length int32
}
type CookiesContext struct {
	Cookies []CookiesData
}

type HeaderData struct {
	Name string
	Value string
}

type CookiesData struct {
	Domain string
	Name string
	Value string
	Path string
}
