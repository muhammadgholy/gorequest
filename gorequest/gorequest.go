package gorequest

import (
	"compress/gzip"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func (GoRequestContext *GoRequestContext) Debug(message string) {
	if (GoRequestContext.EnableDebug) {
		fmt.Println(message);
		
	}
}
func (GoRequestContext *GoRequestContext) Init() {
	if (GoRequestContext.Timeout <= 0) {
		GoRequestContext.Timeout = 15;

	}

	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.MaxIdleConns = 100000;
	customTransport.MaxIdleConnsPerHost = 100000;
	customTransport.MaxConnsPerHost = 100000;
	customTransport.IdleConnTimeout = 90 * time.Second;
	customTransport.TLSHandshakeTimeout = time.Duration(GoRequestContext.Timeout) * time.Second;
	customTransport.DialContext =  (&net.Dialer{
		Timeout:   time.Duration(GoRequestContext.Timeout) * time.Second,
		KeepAlive: 30 * time.Minute,
	}).DialContext;

	// Proxy
	if (GoRequestContext.Proxy != "") {
		proxy, _ := url.Parse(GoRequestContext.ProxyType + "://" + GoRequestContext.Proxy);
		customTransport.Proxy = http.ProxyURL(proxy);

	}
	
	GoRequestContext.HTTPContext = &http.Client{
		Transport: customTransport,
		Timeout: time.Second * time.Duration(GoRequestContext.Timeout),
	}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
		
	// create a custom error to know if a redirect happened.
	var RedirectAttemptedError = errors.New("GOREQUEST_DO_REDIRECT");

	// return the error, so client won't attempt redirects.
	GoRequestContext.HTTPContext.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if (len(via) > GoRequestContext.MaxRedirect) {
			return errors.New("too many redirects");
			
		}
		lastUrlQuery := req.URL.RequestURI();
		GoRequestContext.URLStack = append(GoRequestContext.URLStack, lastUrlQuery);
		GoRequestContext.URLLast = lastUrlQuery;

		return RedirectAttemptedError
	}
}

func (GoRequestContext *GoRequestContext) GetHeaders(uri string) map[string]string {
	headers := make(map[string]string);

	u, err := url.Parse(uri)
    if (err != nil) {
		return headers;
		
    }
	

	if (GoRequestContext.AdditionalHeader) {
		headers["User-Agent"] = GoRequestContext.UserAgent;
		headers["Accept"] = GoRequestContext.Accept;
		headers["Accept-Language"] = "en-US,en;q=0.9,mt;q=0.8";
		headers["Accept-Encoding"] = "gzip, deflate";
		// headers["Connection"] = "Keep-Alive";

	}

	if (GoRequestContext.RequestData != nil) {
		headers["Content-Type"] = GoRequestContext.RequestData.Type;
		headers["Content-Length"] = strconv.Itoa(int(GoRequestContext.RequestData.Length));
		if (GoRequestContext.AdditionalHeader) {
			headers["origin"] = u.Scheme + "://" + u.Host;
		
		}
	}

	if (GoRequestContext.CookiesEnable) {
		// Set Headers Cookies
		currentCookies := GoRequestContext.CookiesFetch(u.Host, "/");
		if (len(currentCookies) > 0) {
			var tmp1 []string;
			for key, value := range currentCookies {
				tmp1 = append(tmp1, key + "=" + strings.Replace(value, " ", "+", -1));
		
			}
			headers["Cookie"] = strings.Join(tmp1, "; ") + ";";
		}
	}

	if (GoRequestContext.Referer != "") {
		headers["Referer"] = GoRequestContext.Referer;
		
	}

	if (GoRequestContext.AdditionalHeader) {
		headers["Upgrade-Insecure-Requests"] = "1";
		headers["Cache-Control"] = "max-age=0";
		
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

func (GoRequestContext *GoRequestContext) GetPage(uri string,) (int, string, string, string) {
	defer func() {
		errorMessage := recover();
		if (errorMessage != nil) {
			fmt.Println("Something went wrong while do request! ");

		}
	}();

	GoRequestContext.Debug("Preparing Request for \"" + uri + "\"");

	var respondHeader string;
	var respondBody string;
	var requestHeader = make(map[string]string);

	// Fix Uri
	u, err := url.Parse(uri);
    if (err != nil) {
		return 0, "", "", err.Error();

    }
	if (strings.ToUpper(u.Scheme) != "HTTP" && strings.ToUpper(u.Scheme) != "HTTPS") {
		return 0, "", "", err.Error();

	}

	// Do It
	var req *http.Request;
	if (GoRequestContext.RequestData != nil) {	
		GoRequestContext.Debug(" > With PostData!");
		var postdata io.Reader;

		//
		if (len(GoRequestContext.RequestData.FormData) > 0) {
			form := url.Values{}
			for k,v := range GoRequestContext.RequestData.FormData {
				GoRequestContext.Debug(" > Added Data \"" + k + "\" with value \"" + v + "\"");
				form.Add(k, v);
	
			}
			postdata = strings.NewReader(form.Encode());

		} else {
			GoRequestContext.Debug(" > Data \"" + GoRequestContext.RequestData.Data + "\"");
			postdata = strings.NewReader(GoRequestContext.RequestData.Data);

		}

		//
		req, _ = http.NewRequest(GoRequestContext.Method, u.String(), postdata);
		if (err != nil) {
			return 0, "", "", err.Error();
	
		}

	} else {
		GoRequestContext.Debug(" > Without PostData!");
		req, _ = http.NewRequest(GoRequestContext.Method, u.String(), nil);
	 	if (err != nil) {
			return 0, "", "", err.Error();
	
		}
	
	}

	// Headers
	requestHeader = GoRequestContext.GetHeaders(uri);
	for hName, hValue := range requestHeader {
		GoRequestContext.Debug(" > Added Header \"" + hName + "\" with value \"" + hValue + "\"");
		req.Header.Add(hName, hValue);

	}
	request, err := GoRequestContext.HTTPContext.Do(req);
	if (err != nil) {
		errMessage := err.Error();
		if (!strings.Contains(errMessage, "GOREQUEST_DO_REDIRECT")) {
			return 0, "", "", err.Error();
	
		}
	}

	var requestRaw string = request.Request.Method + " " + u.RequestURI() + " " + request.Proto + "\n";
	for key, value := range request.Request.Header {
		for _, value2 := range value {
			headerName := strings.ToLower(key);
			headerValue := value2;
			requestRaw = requestRaw + headerName + ": " + headerValue + "\n";

		}
	}
	requestRaw = requestRaw + "\n";
	if (GoRequestContext.RequestData != nil) {	
		if (len(GoRequestContext.RequestData.FormData) > 0) {
			requestRaw = requestRaw + "Form Data: \n";
			for k,v := range GoRequestContext.RequestData.FormData {
				requestRaw = requestRaw + k + ": " + v + "\n";
	
			}

		} else {
			requestRaw = requestRaw + GoRequestContext.RequestData.Data + "\n";

		}
	}
	GoRequestContext.RequestRAW = requestRaw;

	if (GoRequestContext.EnableDebug) {
		fmt.Println("Header: ", request.Request.Header);
		fmt.Println("Method: ", request.Request.Method);
		fmt.Println("Body: ", request.Request.Body);
		fmt.Println("Host: ", request.Request.Host);
		fmt.Println("ContentLength: ", request.ContentLength);
		fmt.Println("Proto: ", request.Proto);
		fmt.Println("TransferEncoding: ", request.Request.TransferEncoding);
		fmt.Println("Cookies: ", request.Request.Cookies());
		fmt.Println("Referer: ", request.Request.Referer());
		fmt.Println(GoRequestContext.RequestRAW);
		
	}

	//
	var newRedirectLink string = "";
	tmp1 := make([]string, 0);
	for key, value := range request.Header {
		for _, value2 := range value {
			tmp1 = append(tmp1, key + ": " + value2);

			// 
			headerName := strings.ToLower(key);
			headerValue := value2;
			GoRequestContext.Debug(" < Get Header \"" + headerName + "\" with value \"" + headerValue + "\"");
			if (headerName == "set-cookie") {
				GoRequestContext.CookiesAdd(headerValue, request.Request.Host);
				
			} else if (headerName == "location") {
				newRedirectLink = strings.TrimSpace(headerValue);

			}
		}
	}
	if (newRedirectLink != "") {
		GoRequestContext.Method = "GET";
		GoRequestContext.RequestData = &RequestData{};

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
	return request.StatusCode, respondHeader, respondBody, "";
}

func (GoRequestContext *GoRequestContext) POST(uri string, postdata string) (int, string, string, string) {
	defer func() {
		errorMessage := recover();
		if (errorMessage != nil) {
			fmt.Println("Something went wrong while POST! ");

		}
	}();

	GoRequestContext.Method = "POST";
	requestData := &RequestData{
		Status: true,
		Type: "application/x-www-form-urlencoded",
		Data: postdata,
		Length: int32(len(postdata)),
	};
	GoRequestContext.RequestData = requestData;

	return GoRequestContext.GetPage(uri);	
}

func (GoRequestContext *GoRequestContext) GET(uri string) (int, string, string, string) {
	defer func() {
		errorMessage := recover();
		if (errorMessage != nil) {
			fmt.Println("Something went wrong while GET! ");

		}
	}();

	GoRequestContext.Method = "GET";
	GoRequestContext.RequestData = nil;
	return GoRequestContext.GetPage(uri);	
}