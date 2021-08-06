package gorequest

type GoRequest interface {
	GET(string) (string, string)
	POST(string, string) (string, string)
	Init()
}
