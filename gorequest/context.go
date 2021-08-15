package gorequest

import "net/http"

type GoRequestContext struct {
	HTTPContext *http.Client

	MaxRedirect int
	EnableDebug bool
	Debugger []string

	Proxy string
	ProxyType string

	Timeout int
}


type NewRequest struct {
	CookiesContext *CookiesContext
	Header []HeaderData

	FollowLocation bool
	AdditionalHeader bool
	CookiesEnable bool
	
	Referer string
	Accept string
	UserAgent string
	RequestRAW string

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
