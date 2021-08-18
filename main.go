package main

import (
	"fmt"
	"os"

	"github.com/muhammadgholy/gorequest/gorequest"
)

func main() {
	var CookiesContext *gorequest.CookiesContext = &gorequest.CookiesContext{};
	var DebuggerContext *gorequest.DebuggerContext = &gorequest.DebuggerContext{};
	var NewRequest *gorequest.NewRequest = &gorequest.NewRequest {
		DebuggerContext: DebuggerContext,
		CookiesContext: CookiesContext,
		CookiesEnable: true,
		AdditionalHeader: true,
	};

	var GoRequest gorequest.GoRequest = &gorequest.GoRequestContext{
		Timeout: 15,
		// EnableDebug: true,
		MaxRedirect: 10,
		
	};
	GoRequest.Init();
	if (len(os.Args) == 2) { 
		statuscode, header, body, err := GoRequest.GET(NewRequest, os.Args[1]);
		fmt.Println(header, "\n", "\n", body, "\n\nStatus Code: ", statuscode, "\nError: ", err);

	} else if (len(os.Args) == 3) { 
		statuscode, header, body, err := GoRequest.POST(NewRequest, os.Args[1], os.Args[2]);
		fmt.Println(header, "\n", "\n", body, "\n\nStatus Code: ", statuscode, "\nError: ", err);

	}
}