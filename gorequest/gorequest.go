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

func (GoRequestContext *GoRequestContext) Debug(Request *NewRequest, message string) {
	if (Request.DebuggerContext.Enable) {
		Request.DebuggerContext.Data = append(Request.DebuggerContext.Data, message);
		fmt.Println(message);
		
	}
}
func (GoRequestContext *GoRequestContext) Init() {
	if (GoRequestContext.Timeout <= 0) {
		GoRequestContext.Timeout = 15;

	}

	if (GoRequestContext.HTTPContext != nil) {
		GoRequestContext.HTTPContext.CloseIdleConnections();

	}

	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.MaxIdleConns = 100000;
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify:true};
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

	// create a custom error to know if a redirect happened.
	var RedirectAttemptedError = errors.New("GOREQUEST_DO_REDIRECT");

	// return the error, so client won't attempt redirects.
	GoRequestContext.HTTPContext.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if (len(via) > GoRequestContext.MaxRedirect) {
			return errors.New("too many redirects");
			
		}

		// Not Important
		// lastUrlQuery := req.URL.RequestURI();
		// GoRequestContext.Request.URLStack = append(GoRequestContext.Request.URLStack, lastUrlQuery);
		// GoRequestContext.Request.URLLast = lastUrlQuery;

		return RedirectAttemptedError
	}
}

func sanitizeOrWarn(fieldName string, valid func(byte) bool, v string) string {
	ok := true
	for i := 0; i < len(v); i++ {
		if valid(v[i]) {
			continue
		}
		ok = false
		break
	}
	if ok {
		return v
	}
	buf := make([]byte, 0, len(v))
	for i := 0; i < len(v); i++ {
		if b := v[i]; valid(b) {
			buf = append(buf, b)
		}
	}
	return string(buf)
}

func validCookieValueByte(b byte) bool {
	return 0x20 <= b && b < 0x7f && b != '"' && b != ';' && b != '\\'
}
func sanitizeCookieValue(v string) string {
	v = sanitizeOrWarn("Cookie.Value", validCookieValueByte, v)
	if len(v) == 0 {
		return v
	}
	if strings.IndexByte(v, ' ') >= 0 || strings.IndexByte(v, ',') >= 0 {
		return `"` + v + `"`
	}
	return v
}

func (GoRequestContext *GoRequestContext) GetHeaders(Request *NewRequest, uri string) map[string]string {
	headers := make(map[string]string);

	u, err := url.Parse(uri)
    if (err != nil) {
		return headers;
		
    }
	

	if (Request.AdditionalHeader) {
		headers["Host"] = u.Host;
		headers["User-Agent"] = "GoStepper";
		headers["Accept"] = "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9";
		headers["Accept-Language"] = "en-US,en;q=0.9,mt;q=0.8";
		headers["Accept-Encoding"] = "gzip, deflate";
		// headers["upgrade-insecure-requests"] = "1";
		// headers["sec-fetch-user"] = "?1";
		// headers["sec-fetch-site"] = "none";
		// headers["sec-fetch-mode"] = "navigate";
		// headers["sec-fetch-dest"] = "document";
		// headers["sec-ch-ua-mobile"] = "?0";
		// headers["sec-ch-ua"] = "\"Chromium\";v=\"92\", \" Not A;Brand\";v=\"99\", \"Google Chrome\";v=\"92\"";
		// headers["upgrade-insecure-requests"] = "1";
		headers["cache-control"] = "no-cache";
		headers["Pragma"] = "no-cache"; // Disable Web Caching
		// headers["Connection"] = "Keep-Alive"; // Not Important
	}

	if (Request.Body.Status) {
		headers["Content-Type"] = Request.Body.Type;
		headers["Content-Length"] = strconv.Itoa(int(Request.Body.Length));
		if (Request.AdditionalHeader) {
			headers["origin"] = u.Scheme + "://" + u.Host;
		
		}
	}

	if (Request.CookiesEnable) {
		// Set Headers Cookies
		currentCookies := GoRequestContext.CookiesFetch(Request, u.Host, "/");
		if (len(currentCookies) > 0) {
			var tmp1 []string;
			for key, value := range currentCookies {
				tmp1 = append(tmp1, key + "=" + sanitizeCookieValue(strings.Replace(value, " ", "+", -1)));
		
			}
			headers["Cookie"] = strings.Join(tmp1, "; ");
		}
	}

	if (Request.Referer != "") {
		headers["Referer"] = Request.Referer;
		
	}

	for _, hData := range Request.Header {
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

func (GoRequestContext *GoRequestContext) GetPage(Request *NewRequest, uri string,) (int, string, string, string) {
	// defer func() {
	// 	errorMessage := recover();
	// 	if (errorMessage != nil) {
	// 		fmt.Println("Something went wrong while do request! ");

	// 	}
	// }();

	Request.RequestLastUrl = uri;

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
	if (Request.Body.Status) {	
		req, _ = http.NewRequest(Request.Method, u.String(), strings.NewReader(Request.Body.Data));
		if (err != nil) {
			return 0, "", "", err.Error();
	
		}

		// Change to form data
		if (len(Request.Body.FormData) > 0) {
			form := url.Values{}
			for k,v := range Request.Body.FormData {
				form.Add(k, v);
	
			}
			req.PostForm = form;
			req.ParseForm();

		}

	} else {
		req, _ = http.NewRequest(Request.Method, u.String(), nil);
	 	if (err != nil) {
			return 0, "", "", err.Error();
	
		}

	}

	// Headers
	requestHeader = GoRequestContext.GetHeaders(Request, uri);
	for hName, hValue := range requestHeader {
		if (strings.ToLower(hName) == "accept-encoding") {
			tmp1 := strings.Split(hValue, ",");
			for _, v := range tmp1 {
				tmp2 := strings.TrimSpace(v);
				req.TransferEncoding = append(req.TransferEncoding, tmp2);

			}

		} else {
			req.Header.Add(hName, hValue);

		}
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
	if (Request.Body.Status) {	
		requestRaw = requestRaw + "> \n";
		if (len(Request.Body.FormData) > 0) {
			requestRaw = requestRaw + "Form Data: \n";
			for k,v := range Request.Body.FormData {
				requestRaw = requestRaw + "> " + k + ": " + v + "\n";
	
			}

		} else {
			requestRaw = requestRaw + "> " + Request.Body.Data + "\n";

		}
	}
	Request.RequestRAW = requestRaw;

	if (Request.DebuggerContext.Enable) {
		GoRequestContext.Debug(Request, strings.TrimSpace(Request.RequestRAW));
		GoRequestContext.Debug(Request, ">");
		GoRequestContext.Debug(Request, "< " + request.Proto + " " + strconv.Itoa(request.StatusCode));
		
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
			if (Request.DebuggerContext.Enable) {
				GoRequestContext.Debug(Request, "< " + headerName + ": " + headerValue);
				
			}

			//
			if (headerName == "set-cookie") {
				GoRequestContext.CookiesAdd(Request, headerValue, request.Request.Host);
				
			} else if (headerName == "location") {
				newRedirectLink = strings.TrimSpace(headerValue);

			}
		}
	}
	if (newRedirectLink != "") {
		Request.Method = "GET";
		Request.Body = RequestBody{};
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

		Request.Referer = u.String();
		return GoRequestContext.GetPage(Request, newRedirectLink);
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

func (GoRequestContext *GoRequestContext) POST(Request *NewRequest, uri string, postdata string) (int, string, string, string) {
	defer func() {
		errorMessage := recover();
		if (errorMessage != nil) {
			fmt.Println("Something went wrong while POST! ");

		}
	}();

	Request.Method = "POST";
	Request.Body = RequestBody {
		Status: true,
		Type: "application/x-www-form-urlencoded",
		Data: postdata,
		Length: int32(len(postdata)),
	};

	return GoRequestContext.GetPage(Request, uri);	
}

func (GoRequestContext *GoRequestContext) GET(Request *NewRequest, uri string) (int, string, string, string) {
	// defer func() {
	// 	errorMessage := recover();
	// 	if (errorMessage != nil) {
	// 		fmt.Println("Something went wrong while GET! ");

	// 	}
	// }();

	Request.Method = "GET";
	Request.Body = RequestBody {}

	return GoRequestContext.GetPage(Request, uri);	
}