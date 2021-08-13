package gorequest

type GoRequest interface {
	GET(string) (int, string, string, string)
	POST(string, string) (int, string, string, string)
	Init()
}
