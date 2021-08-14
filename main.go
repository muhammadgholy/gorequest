package main

import (
	"fmt"
	"os"

	"github.com/muhammadgholy/gorequest/gorequest"
)

func main() {
	var GoRequestCookiesContext gorequest.GoRequestCookiesContext = gorequest.GoRequestCookiesContext{};
	var GoRequest gorequest.GoRequest = &gorequest.GoRequestContext{
		CookiesEnable: true,
		AdditionalHeader: true,
		HTTPVersion: "1",
		MaxRedirect: 10,
		// EnableDebug: true,
		CookiesContext: &GoRequestCookiesContext,
		Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:72.0) Gecko/20100101 Firefox/72.0",
	};

	GoRequest.Init();
	if (len(os.Args) == 2) { 
		statuscode, header, _, err := GoRequest.GET(os.Args[1]);
		fmt.Println(header, "\n", "\n", "\r\nStatus Code: ", statuscode, "\nError: ", err);

	} else if (len(os.Args) == 3) { 
		statuscode, header, body, err := GoRequest.POST(os.Args[1], os.Args[2]);
		fmt.Println(header, "\n", "\n", body, "\r\nStatus Code: ", statuscode, "\nError: ", err);

	}
}