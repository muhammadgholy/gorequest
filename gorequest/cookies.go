package gorequest

import (
	"fmt"
	"strings"
	"time"
)

func (GoRequestContext *GoRequestContext) CookiesAdd(data string, host string) {
	list := strings.Split(data, ";");
	if (!strings.Contains(list[0], "=")) {
		return;

	}

	tmp1 := strings.SplitN(list[0], "=", 2);
	cookieName := tmp1[0];
	cookieValue := tmp1[1];

	path := "/";
	domain := host;
	isExpired := false;
	
	
	for _, value := range list {
		value = strings.TrimSpace(value);

		var tName string;
		var tValue string;

		if (strings.Contains(value, "=")) {
			tmp1 := strings.SplitN(value, "=", 2);
			tName = strings.TrimSpace(tmp1[0]);
			tValue = strings.TrimSpace(tmp1[1]);

		}
		
		if (strings.ToLower(tName) == "path") {
			if (tValue != "") {
				path = tValue;

			}

		} else if (strings.ToLower(tName) == "domain") {
			if (tValue != "") {
				if (tValue[0:1] == ".") {
					domain = tValue;

				} else {
					domain = "." + tValue;

				}
			}

		} else if (strings.ToLower(tName) == "expires") {
			layout := "Mon, 02-Jan-2006 15:04:05 MST"
			t, err := time.Parse(layout, tValue)
			
			if err != nil {
				fmt.Println(err)
			}

			var expiryTime int = int(t.Unix());
			var currentTime int = int(time.Now().Unix());
			if (expiryTime < currentTime) {
				GoRequestContext.CookiesDelete(cookieName, domain, path);
				isExpired = true;

			}
		}
	}

	if (!isExpired) {
		newCookieData := CookiesData{
			Domain: domain,
			Name: cookieName,
			Value: cookieValue,
			Path: path,
		};
		GoRequestContext.Cookies = append(GoRequestContext.Cookies, newCookieData);

	}
}

func (GoRequestContext *GoRequestContext) CookiesDelete(name string, domain string, path string) {
	for key, value := range GoRequestContext.Cookies {
		if (value.Domain == domain && value.Name == name && value.Path == path) {
			GoRequestContext.Cookies = append(GoRequestContext.Cookies[:key], GoRequestContext.Cookies[key+1:]...);

		}
	}
}

func (GoRequestContext *GoRequestContext) CookiesFetch(domain string, path string) map[string]string {
	var cookies = make(map[string]string);
	if (path == "") {
		path = "/";

	}
	if (domain == "") {
		return cookies;

	}
	for _, value := range GoRequestContext.Cookies {
		if (value.Domain[0:1] == ".") {
			if (len(value.Domain) >= len(domain)) {
				if ((domain[len(domain)-len(value.Domain):] == value.Domain[1:]) || (value.Domain == domain)) {
					cookies[value.Name] = value.Value;
	
				}
			}

		} else if (value.Domain == domain) {
			cookies[value.Name] = value.Value;

		}
	}
	return cookies;
}