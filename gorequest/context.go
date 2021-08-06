package gorequest

import "net/http"

type GoRequestContext struct {
	HTTPContext *http.Client
	Header []HeaderData
	Cookies []CookiesData
	URLLast string
	URLStack []string
	CookiesEnable bool
	AditionalHeader bool
	Accept string
	UserAgent  string
	FollowLocation bool
	Referer string
	Method string
	RequestData RequestData
}

type RequestData struct {
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
