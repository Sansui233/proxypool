package tool

import (
	"bytes"
	"errors"
	"github.com/robertkrimen/otto"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
)

// Find email playload
func GetCFEmailPayload(str string) string {
	s := strings.Split(str, "data-cfemail=")
	if len(s) > 1 {
		s = strings.Split(s[1], "\"")
		str = s[1]
		return str
	}
	return ""
}

// Remove cloudflare email protection
func CFEmailDecode(a string) (s string, err error) {
	if a == "" {
		return "", errors.New("CFEmailDecodeError: empty payload to decode")
	}
	var e bytes.Buffer
	r, _ := strconv.ParseInt(a[0:2], 16, 0)
	for n := 4; n < len(a)+2; n += 2 {
		i, _ := strconv.ParseInt(a[n-2:n], 16, 0)
		//e.WriteString(string(i ^ r))
		e.WriteString(string(rune(i ^ r)))
	}
	return e.String(), nil
}

// Return full accessible url from a script protected url. If not a script url, return input
func CFScriptRedirect(url string) (string, error) {
	resp, err := GetHttpClient().Get(url)
	if err != nil {
		return url, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return url, err
	}
	strbody := string(body)
	if len(strbody) < 7 {
		return url, nil
	}
	if strbody[:7] == "<script" {
		js := strings.Split(strbody, "javascript\">")[1]
		js = strings.Split(js, "</script>")[0]
		js = ScriptReplace(js, "strdecode")
		reUrl, err := ScriptGet(js, "strdecode")
		if err != nil {
			return url, err
		}
		if reUrl != "" {
			return reUrl, nil
		} else {
			return url, errors.New("RedirectionError: result from javascript")
		}
	}
	return url, nil
}

// Get result var of a js script
func ScriptGet(js string, varname string) (string, error) {
	vm := otto.New()
	_, err := vm.Run(js)
	if err != nil {
		return "", err
	}
	if value, err := vm.Get(varname); err == nil {
		if v, err := value.ToString(); err == nil {
			return v, nil
		}
	}
	return "", err
}

// Replace location with varname and remove window
func ScriptReplace(js string, varname string) string {
	strs := strings.Split(js, ";")
	varWindow := ""
	varLocation := ""
	bound := len(strs)

	if len(js) < 2 {
		return js
	}
	for i, _ := range strs {
		//replace location
		if varLocation != "" && strings.Contains(strs[i], varLocation) {
			re3, err := regexp.Compile(varLocation + ".*?[]]") // _LoKlO[_jzvXT]
			if err == nil {
				strs[i] = re3.ReplaceAllLiteralString(strs[i], varname)
			}
		}
		if strings.Contains(strs[i], "location") {
			strarr := strings.Split(strs[i], " = ")
			if len(strarr) >= 2 { // get varname, _jzvXT = location  or  return '/t' } _qf14P = location
				if strarr[len(strarr)-1] == "location" {
					index := strings.LastIndex(strs[i], "}")
					if index == -1 {
						varLocation = strarr[0]
						strs[i] = ""
					} else {
						strs[i] = strs[i][:index+1]
						varLocation = strings.Split(strs[i][index+1:], " = ")[0]
						varLocation = strings.TrimSpace(varLocation)
					}
				}
			} else { // set varname
				re, err := regexp.Compile("location.*?[]]=") // location[_jzvXT]=
				if err == nil {
					strs[i] = re.ReplaceAllLiteralString(strs[i], varname+"=")
				}
				re, err = regexp.Compile("location.*?[]]") // location[_jzvXT]
				if err == nil {
					strs[i] = re.ReplaceAllLiteralString(strs[i], varname+"=")
				}
				strs[i] = strings.ReplaceAll(strs[i], "location.replace = ", varname+"=")
				strs[i] = strings.ReplaceAll(strs[i], "location.replace=", varname+"=")
				strs[i] = strings.ReplaceAll(strs[i], "location.replace", varname+"=")
				strs[i] = strings.ReplaceAll(strs[i], "location.assign = ", varname+"=")
				strs[i] = strings.ReplaceAll(strs[i], "location.assign=", varname+"=")
				strs[i] = strings.ReplaceAll(strs[i], "location.assign", varname+"=")
				strs[i] = strings.ReplaceAll(strs[i], "location.href =", varname+"=")
				strs[i] = strings.ReplaceAll(strs[i], "location.href=", varname+"=")
				strs[i] = strings.ReplaceAll(strs[i], "location.href", varname+"=")
				strs[i] = strings.ReplaceAll(strs[i], "location=", varname+"=")
				strs[i] = strings.ReplaceAll(strs[i], "==", varname+"=")
			}
		}
		// remove window
		if strings.Contains(strs[i], "window") {
			index := strings.LastIndex(strs[i], "}")
			if index == -1 {
				varWindow = strings.Split(strs[i], " = window")[0]
				strs[i] = ""
			} else {
				varWindow = strings.Split(strs[i][index+1:], " = ")[0]
				varWindow = strings.TrimSpace(varWindow)
				strs[i] = strs[i][:index+1]
			}
		}
	}

	if varWindow != "" {
		for i, _ := range strs {
			if strings.Contains(strs[i], varWindow) {
				bound = i
				break
			}
		}
	}
	js = strings.Join(strs[:bound], ";")
	return js
}
