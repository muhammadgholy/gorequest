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
	if (GoRequestContext.Request.EnableDebug) {
		GoRequestContext.Request.Debugger = append(GoRequestContext.Request.Debugger, message);
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
		if (len(via) > GoRequestContext.Request.MaxRedirect) {
			return errors.New("too many redirects");
			
		}
		lastUrlQuery := req.URL.RequestURI();
		GoRequestContext.Request.URLStack = append(GoRequestContext.Request.URLStack, lastUrlQuery);
		GoRequestContext.Request.URLLast = lastUrlQuery;

		return RedirectAttemptedError
	}
}

func (GoRequestContext *GoRequestContext) GetHeaders(uri string) map[string]string {
	headers := make(map[string]string);

	u, err := url.Parse(uri)
    if (err != nil) {
		return headers;
		
    }
	

	if (GoRequestContext.Request.AdditionalHeader) {
		headers["User-Agent"] = GoRequestContext.Request.UserAgent;
		headers["Accept"] = GoRequestContext.Request.Accept;
		headers["Accept-Language"] = "en-US,en;q=0.9,mt;q=0.8";
		headers["Accept-Encoding"] = "gzip, deflate";
		headers["upgrade-insecure-requests"] = "1";
		headers["sec-fetch-user"] = "?1";
		headers["sec-fetch-site"] = "none";
		headers["sec-fetch-mode"] = "navigate";
		headers["sec-fetch-dest"] = "document";
		headers["sec-ch-ua-mobile"] = "?0";
		headers["sec-ch-ua"] = "\"Chromium\";v=\"92\", \" Not A;Brand\";v=\"99\", \"Google Chrome\";v=\"92\"";
		headers["upgrade-insecure-requests"] = "1";
		headers["cache-control"] = "no-cache";
		headers["Pragma"] = "no-cache"; // Disable Web Caching
		// headers["Connection"] = "Keep-Alive"; // Not Important

	}

	if (GoRequestContext.Request.Body.Status) {
		headers["Content-Type"] = GoRequestContext.Request.Body.Type;
		headers["Content-Length"] = strconv.Itoa(int(GoRequestContext.Request.Body.Length));
		if (GoRequestContext.Request.AdditionalHeader) {
			headers["origin"] = u.Scheme + "://" + u.Host;
		
		}
	}

	if (GoRequestContext.Request.CookiesEnable) {
		// Set Headers Cookies
		currentCookies := GoRequestContext.CookiesFetch(u.Host, "/");
		if (len(currentCookies) > 0) {
			var tmp1 []string;
			for key, value := range currentCookies {
				tmp1 = append(tmp1, key + "=" + strings.Replace(value, " ", "+", -1));
		
			}
			headers["Cookie"] = strings.Join(tmp1, "; ");
		}
	}

	if (GoRequestContext.Request.Referer != "") {
		headers["Referer"] = GoRequestContext.Request.Referer;
		
	}

	if (GoRequestContext.Request.AdditionalHeader) {
		headers["Upgrade-Insecure-Requests"] = "1";
		headers["Cache-Control"] = "max-age=0";
		
	}

	for _, hData := range GoRequestContext.Request.Header {
		if (strings.ToLower(hData.Name) == "content-type") {
			headers["Content-Type"] = hData.Value;

		} else if (strings.ToLower(hData.Name) == "accept") {
			headers["Accept"] = hData.Value;

		} else if (strings.ToLower(hData.Name) == "accept-encoding") {
			headers["Accept-Encoding"] = hData.Value;
	
		} else {
			var founded bool = false;
			for tmp1 := range headers {
				if (strings.EqualFold(tmp1, hData.Name)) {
					headers[tmp1] = hData.Value;
					founded = true;

				}
			}
			if (!founded) {
				headers[hData.Name] = hData.Value;

			}

		}
	}

	return headers;
}

func (GoRequestContext *GoRequestContext) GetPage(uri string,) (int, string, string, string) {
	// defer func() {
	// 	errorMessage := recover();
	// 	if (errorMessage != nil) {
	// 		fmt.Println("Something went wrong while do request! ");

	// 	}
	// }();

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
	if (GoRequestContext.Request.Body.Status) {	
		var postdata io.Reader;

		//
		if (len(GoRequestContext.Request.Body.FormData) > 0) {
			form := url.Values{}
			for k,v := range GoRequestContext.Request.Body.FormData {
				form.Add(k, v);
	
			}
			postdata = strings.NewReader(form.Encode());

		} else {
			postdata = strings.NewReader(GoRequestContext.Request.Body.Data);

		}

		//
		req, _ = http.NewRequest(GoRequestContext.Request.Method, u.String(), postdata);
		if (err != nil) {
			return 0, "", "", err.Error();
	
		}

	} else {
		req, _ = http.NewRequest(GoRequestContext.Request.Method, u.String(), nil);
	 	if (err != nil) {
			return 0, "", "", err.Error();
	
		}
	
	}

	// Headers
	requestHeader = GoRequestContext.GetHeaders(uri);
	for hName, hValue := range requestHeader {
		req.Header.Add(hName, hValue);

	}
	request, err := GoRequestContext.HTTPContext.Do(req);
	if (err != nil) {
		errMessage := err.Error();
		if (!strings.Contains(errMessage, "GOREQUEST_DO_REDIRECT")) {
			return 0, "", "", err.Error();
	
		}
	}

	var requestRaw string = "> " + request.Request.Method + " " + u.RequestURI() + " " + request.Proto + "\n";
	for key, value := range request.Request.Header {
		for _, value2 := range value {
			headerName := strings.ToLower(key);
			headerValue := value2;
			requestRaw = requestRaw + "> " + headerName + ": " + headerValue + "\n";

		}
	}
	if (GoRequestContext.Request.Body.Status) {	
		requestRaw = requestRaw + "> \n";
		if (len(GoRequestContext.Request.Body.FormData) > 0) {
			requestRaw = requestRaw + "Form Data: \n";
			for k,v := range GoRequestContext.Request.Body.FormData {
				requestRaw = requestRaw + "> " + k + ": " + v + "\n";
	
			}

		} else {
			requestRaw = requestRaw + "> " + GoRequestContext.Request.Body.Data + "\n";

		}
	}
	GoRequestContext.Request.RequestRAW = requestRaw;

	if (GoRequestContext.Request.EnableDebug) {
		GoRequestContext.Debug(strings.TrimSpace(GoRequestContext.Request.RequestRAW));
		GoRequestContext.Debug(">");
		GoRequestContext.Debug("< " + request.Proto + " " + strconv.Itoa(request.StatusCode));
		
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

			//
			if (GoRequestContext.Request.EnableDebug) {
				GoRequestContext.Debug("< " + headerName + ": " + headerValue);
				
			}

			//
			if (headerName == "set-cookie") {
				GoRequestContext.CookiesAdd(headerValue, request.Request.Host);
				
			} else if (headerName == "location") {
				newRedirectLink = strings.TrimSpace(headerValue);

			}
		}
	}
	if (newRedirectLink != "") {
		GoRequestContext.Request.Method = "GET";
		GoRequestContext.Request.Body = RequestBody{};
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

		GoRequestContext.Request.Referer = u.String();
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

	GoRequestContext.Request.Method = "POST";
	GoRequestContext.Request.Body = RequestBody {
		Status: true,
		Type: "application/x-www-form-urlencoded",
		Data: postdata,
		Length: int32(len(postdata)),
	};

	return GoRequestContext.GetPage(uri);	
}

func (GoRequestContext *GoRequestContext) GET(uri string) (int, string, string, string) {
	// defer func() {
	// 	errorMessage := recover();
	// 	if (errorMessage != nil) {
	// 		fmt.Println("Something went wrong while GET! ");

	// 	}
	// }();

	GoRequestContext.Request.Method = "GET";
	GoRequestContext.Request.Body = RequestBody {}

	return GoRequestContext.GetPage(uri);	
}