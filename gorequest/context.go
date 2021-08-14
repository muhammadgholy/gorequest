package gorequest

import "net/http"

type GoRequestContext struct {
	HTTPContext *http.Client

	Request *NewRequest

	Proxy string
	ProxyType string

	Timeout int
}


type NewRequest struct {
	Cookies []CookiesData
	Header []HeaderData
	
	EnableDebug bool

	FollowLocation bool
	AdditionalHeader bool
	CookiesEnable bool
	
	Referer string
	Accept string
	UserAgent string
	RequestRAW string

	MaxRedirect int

	URLLast string
	URLStack []string

	Debugger []string

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
