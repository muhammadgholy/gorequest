package gorequest

import (
	"compress/gzip"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func (GoRequestContext *GoRequestContext) Init() {

	// create a custom error to know if a redirect happened.
	var RedirectAttemptedError = errors.New("redirect");

	// return the error, so client won't attempt redirects.
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	GoRequestContext.HTTPContext = &http.Client{}
	GoRequestContext.HTTPContext.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) > 10 {
			return errors.New("too many redirects");
			
		}
		lastUrlQuery := req.URL.RequestURI();
		GoRequestContext.URLStack = append(GoRequestContext.URLStack, lastUrlQuery);
		GoRequestContext.URLLast = lastUrlQuery;

		return RedirectAttemptedError
	}
}

func (GoRequestContext *GoRequestContext) GetHeaders(uri string) map[string]string {
	u, err := url.Parse(uri)
    if err != nil {
        panic(err)
    }
	
	headers := make(map[string]string);

	if (GoRequestContext.AditionalHeader) {
		headers["User-Agent"] = GoRequestContext.UserAgent;
		headers["Accept"] = GoRequestContext.Accept;
		headers["Accept-Language"] = "en-US,en;q=0.9,mt;q=0.8";
		headers["Accept-Encoding"] = "gzip, deflate";

	}

	if (GoRequestContext.RequestData.Status) {
		headers["Content-Type"] = GoRequestContext.RequestData.Type;
		headers["Content-Length"] = strconv.Itoa(int(GoRequestContext.RequestData.Length));
		if (GoRequestContext.AditionalHeader) {
			headers["origin"] = u.Scheme + "://" + u.Host;
		
		}
	}

	if (GoRequestContext.CookiesEnable) {
		// Set Headers Cookies
		currentCookies := GoRequestContext.CookiesFetch(u.Host, "/");
		if (len(currentCookies) > 0) {
			var tmp1 []string;
			for key, value := range currentCookies {
				tmp1 = append(tmp1, key + "=" + strings.Replace(value, " ", "+", 0));
		
			}
			headers["Cookie"] = strings.Join(tmp1, "; ") + ";";
		}
	}

	if (GoRequestContext.Referer != "") {
		headers["Referer"] = GoRequestContext.Referer;
		
	}

	if (GoRequestContext.AditionalHeader) {
		headers["Upgrade-Insecure-Requests"] = "1";
		headers["Cache-Control"] = "max-age=0";
		headers["TE"] = "Trailers";

	}

	for _, hData := range GoRequestContext.Header {
		if (strings.ToLower(hData.Name) == "content-type") {
			headers["Content-Type"] = hData.Value;

		} else if (strings.ToLower(hData.Name) == "accept") {
			headers["Accept"] = hData.Value;

		} else if (strings.ToLower(hData.Name) == "accept-encoding") {
			headers["Accept-Encoding"] = hData.Value;
	
		} else {
			headers[hData.Name] = hData.Value;
	
		}
	}

	return headers;
}

func (GoRequestContext *GoRequestContext) GetPage(uri string,) (string, string) {
	var respondHeader string;
	var respondBody string;
	var requestHeader = make(map[string]string);

	// Fix Uri
	u, err := url.Parse(uri);
    if (err != nil) {
        panic(err);

    }
	if (strings.ToUpper(u.Scheme) != "HTTP" && strings.ToUpper(u.Scheme) != "HTTPS") {
		panic("Invalid URL Schema");

	}

	// Do It
	var req *http.Request;
	if (GoRequestContext.RequestData.Status) {
		var postdata io.Reader;

		//
		if (len(GoRequestContext.RequestData.FormData) > 0) {
			form := url.Values{}
			for k,v := range GoRequestContext.RequestData.FormData {
				form.Add(k, v);
	
			}
			postdata = strings.NewReader(form.Encode());

		} else {
			postdata = strings.NewReader(GoRequestContext.RequestData.Data);

		}

		//
		req, _ = http.NewRequest(GoRequestContext.Method, u.String(), postdata);	

	} else {
		req, _ = http.NewRequest(GoRequestContext.Method, u.String(), nil);
	
	}

	// Header
	requestHeader = GoRequestContext.GetHeaders(uri);
	for hName, hValue := range requestHeader {
		req.Header.Add(hName, hValue);

	}
	request, _ := GoRequestContext.HTTPContext.Do(req);

	//
	var newRedirectLink string = "";
	tmp1 := make([]string, 0);
	for key, value := range request.Header {
		for _, value2 := range value {
			tmp1 = append(tmp1, key + ": " + value2);

			// 
			headerName := strings.ToLower(key);
			headerValue := value2;
			if (headerName == "set-cookie") {
				GoRequestContext.CookiesAdd(headerValue, request.Request.Host);
				
			} else if (headerName == "location") {
				newRedirectLink = strings.TrimSpace(headerValue);

			}
		}
	}
	if (newRedirectLink != "") {
		GoRequestContext.Method = "GET";
		GoRequestContext.RequestData = RequestData{};

		if (strings.ToUpper(newRedirectLink[:4]) != "HTTP") {
			if (strings.Contains(newRedirectLink, "?")) {
				tmp2 := strings.SplitN(newRedirectLink, "?", 2);
				u.Path = tmp2[0];
				if (strings.Contains(tmp2[1], "#")) {
					tmp3 := strings.SplitN(tmp2[1], "#", 2);
					u.RawQuery = tmp3[0];
					u.Fragment = tmp3[1];

				} else {
					u.RawQuery = tmp2[1];
					u.Fragment = "";

				}

			} else {
				u.Path = newRedirectLink;
				u.Fragment = "";
				u.RawQuery = "";

			}
			newRedirectLink = u.String();
		}

		GoRequestContext.Referer = u.String();
		return GoRequestContext.GetPage(newRedirectLink);
	}
	respondHeader = strings.Join(tmp1, "\n");

	// Check that the server actual sent compressed data
	var reader io.ReadCloser
	switch request.Header.Get("Content-Encoding") {
		case "gzip":
			reader, err = gzip.NewReader(request.Body)
			if err != nil {
				log.Fatal(err)
			}
			defer reader.Close();

		default:
			reader = request.Body;
	}
	
	//
	tmpbody, err := ioutil.ReadAll(reader)
	if (err != nil) {
	   log.Fatalln(err)
	}

	//
	respondBody = string(tmpbody);
	
	//
	return respondHeader, respondBody;
}

func (GoRequestContext *GoRequestContext) POST(uri string, postdata string) (string, string) {
	defer func() {
		errorMessage := recover();
		if (errorMessage != nil) {
			fmt.Println("Something went wrong while GET! " + errorMessage.(string));

		}
	}();

	GoRequestContext.Method = "POST";
	requestData := RequestData{
		Status: true,
		Type: "application/x-www-form-urlencoded",
		Data: postdata,
		Length: int32(len(postdata)),
	};
	GoRequestContext.RequestData = requestData;

	return GoRequestContext.GetPage(uri);	
}

func (GoRequestContext *GoRequestContext) GET(uri string) (string, string) {
	defer func() {
		errorMessage := recover();
		if (errorMessage != nil) {
			fmt.Println("Something went wrong while GET! " + errorMessage.(string));

		}
	}();

	return GoRequestContext.GetPage(uri);	
}