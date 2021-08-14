package main

import (
	"fmt"
	"os"

	"github.com/muhammadgholy/gorequest/gorequest"
)

func main() {
	var NewRequest gorequest.NewRequest = gorequest.NewRequest {
		MaxRedirect: 10,
		CookiesEnable: true,
		AdditionalHeader: true,
		Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36",
		EnableDebug: true,
		Header: []gorequest.HeaderData{
			gorequest.HeaderData{
				Name: "upgrade-insecure-requests",
				Value: "1",
			},
		},
	};

	var GoRequest gorequest.GoRequest = &gorequest.GoRequestContext{
		Request: &NewRequest,
		Timeout: 15,
	};

	GoRequest.Init();
	if (len(os.Args) == 2) { 
		statuscode, header, body, err := GoRequest.GET(os.Args[1]);
		fmt.Println(header, "\n", "\n", body, "\n\nStatus Code: ", statuscode, "\nError: ", err);

	} else if (len(os.Args) == 3) { 
		statuscode, header, body, err := GoRequest.POST(os.Args[1], os.Args[2]);
		fmt.Println(header, "\n", "\n", body, "\n\nStatus Code: ", statuscode, "\nError: ", err);

	}
}