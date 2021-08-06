package utility

import (
	"strings"
)

func StringToFormData(str string) map[string]string {
	fmtPostdata := make(map[string]string);
	for _,dd := range strings.Split(str, "&") {
		dds := strings.SplitN(dd, "=", 2);
		fmtPostdata[dds[0]] = dds[1];

	}
	return fmtPostdata;

}