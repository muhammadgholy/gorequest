package gorequest

type GoRequest interface {
	GET(*NewRequest, string) (int, string, string, string)
	POST(*NewRequest, string, string) (int, string, string, string)
	Init()
}
