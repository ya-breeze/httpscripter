package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/fatih/color"
	"github.com/tidwall/gjson"
)

var Last *Storage

type Storage struct {
	BaseURL      string
	Client       *http.Client
	Request      *http.Request
	RequestBody  string
	Response     *http.Response
	ResponseBody string
}

func init() {
	Last = &Storage{
		Client: &http.Client{},
	}
}

func GET(urlString string, params ...string) {
	Send("GET", urlString, "", params...)
}

func POST(urlString, body string, params ...string) {
	Send("POST", urlString, body, params...)
}

func DELETE(urlString string, params ...string) {
	Send("POST", urlString, "", params...)
}

func PATCH(urlString, body string, params ...string) {
	Send("PATCH", urlString, body, params...)
}

func PUT(urlString, body string, params ...string) {
	Send("PUT", urlString, body, params...)
}

func Succeed(code int) bool {
	return code >= http.StatusOK && code < http.StatusMultipleChoices
}

func Failed(code int) bool {
	return !Succeed(code)
}

func JSON(obj map[string]interface{}) string {
	jsonStr, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}
	return string(jsonStr)
}

func Value(path string) gjson.Result {
	return gjson.Get(Last.ResponseBody, path)
}

func Send(method string, urlString, body string, params ...string) {
	headers := map[string]string{}
	queryParams := map[string]string{}
	Last.RequestBody = body

	for _, param := range params {
		switch {
		case strings.Contains(param, "=="):
			parts := strings.Split(param, "==")
			queryParams[parts[0]] = parts[1]
		case strings.Contains(param, ":"):
			parts := strings.Split(param, ":")
			headers[parts[0]] = parts[1]
		}
	}

	fullURL, err := url.Parse(Last.BaseURL + urlString)
	if err != nil {
		fmt.Println("Error parsing URL:", err)
		panic(err)
	}
	urlParams := url.Values{}
	for key, value := range queryParams {
		urlParams.Add(key, value)
	}
	fullURL.RawQuery = urlParams.Encode()

	Last.Request, err = http.NewRequest(method, fullURL.String(), bytes.NewBuffer([]byte(Last.RequestBody)))
	if err != nil {
		fmt.Println("Error:", err)
		panic(err)
	}

	for key, value := range headers {
		Last.Request.Header.Set(key, value)
	}
	printRequest(Last.Request, Last.RequestBody)

	Last.Response, err = Last.Client.Do(Last.Request)
	if err != nil {
		fmt.Println("Error:", err)
		panic(err)
	}
	defer Last.Response.Body.Close()

	respBody, err := io.ReadAll(Last.Response.Body)
	if err != nil {
		fmt.Println("Error:", err)
		panic(err)
	}
	Last.ResponseBody = string(respBody)

	printResponse(Last.Response, Last.ResponseBody)
}

func printResponse(resp *http.Response, body string) {
	respDump, err := httputil.DumpResponse(resp, false) // Don't include body here since we already captured it
	if err != nil {
		fmt.Println("Error dumping response:", err)
		return
	}
	printColoredResponse(string(respDump), []byte(body))
}

func printColoredResponse(headers string, body []byte) {
	// Define colors for headers
	httpColor := color.New(color.FgCyan, color.Bold)
	codeColor := color.New(color.FgYellow, color.Bold)
	messageColor := color.New(color.FgGreen, color.Bold)
	headerKeyColor := color.New(color.FgGreen)
	headerValueColor := color.New(color.FgWhite)

	// Print headers
	lines := bytes.Split([]byte(headers), []byte("\n"))
	for i, line := range lines {
		lineStr := string(line)
		if i == 0 {
			parts := strings.SplitN(lineStr, " ", 3)
			if len(parts) == 3 {
				httpColor.Printf("%s ", parts[0])
				codeColor.Printf("%s ", parts[1])
				messageColor.Printf("%s\n", parts[2])
			} else {
				fmt.Println(lineStr)
			}
		} else if lineStr == "" {
			// Print a blank line (between headers and body)
			// fmt.Println(lineStr)
			break
		} else {
			if bytes.Contains(line, []byte(":")) {
				headerParts := bytes.SplitN(line, []byte(":"), 2)
				headerKeyColor.Printf("%s:", headerParts[0])
				headerValueColor.Printf("%s\n", headerParts[1])
			} else {
				// Print other lines (e.g., HTTP status line)
				fmt.Println(lineStr)
			}
		}
	}

	// Print body (assume it's JSON for this example)
	if len(body) > 0 {
		prettyJSON, err := prettyPrintJSON(body)
		if err == nil {
			printColoredJSON(prettyJSON)
		} else {
			// If body is not JSON, print it as raw text
			fmt.Println(string(body))
		}
	}
}

func printRequest(req *http.Request, body string) {
	reqDump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		fmt.Println("Error dumping request:", err)
		return
	}
	printColoredRequest(string(reqDump))

	// Print body (assume it's JSON for this example)
	if len(body) > 0 {
		prettyJSON, err := prettyPrintJSON([]byte(body))
		if err == nil {
			printColoredJSON(prettyJSON)
		} else {
			// If body is not JSON, print it as raw text
			fmt.Println(string(body))
		}
	}
}

func printColoredRequest(request string) {
	// Define colors using the fatih/color package
	methodColor := color.New(color.FgCyan, color.Bold)
	urlColor := color.New(color.FgYellow, color.Bold)
	headerKeyColor := color.New(color.FgGreen)
	headerValueColor := color.New(color.FgWhite)
	// bodyColor := color.New(color.FgMagenta)

	// Split the request into lines
	lines := bytes.Split([]byte(request), []byte("\n"))
	for i, line := range lines {
		lineStr := strings.TrimSpace(string(line))

		if i == 0 {
			// Print the request line (Method + URL + HTTP version)
			parts := bytes.Fields(line)
			if len(parts) == 3 {
				methodColor.Printf("%s ", parts[0])
				urlColor.Printf("%s ", parts[1])
				fmt.Printf("%s\n", parts[2])
			} else {
				fmt.Println(lineStr)
			}
		} else if lineStr == "" {
			// Print a blank line (between headers and body)
			fmt.Println(lineStr)
			break
		} else if bytes.Contains(line, []byte(":")) {
			// Print headers (Key: Value)
			headerParts := bytes.SplitN(line, []byte(":"), 2)
			headerKeyColor.Printf("%s:", headerParts[0])
			headerValueColor.Printf("%s\n", headerParts[1])
		}
	}
}

// prettyPrintJSON takes raw JSON and returns a pretty-printed version
func prettyPrintJSON(rawJSON []byte) (string, error) {
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, rawJSON, "", "  ")
	if err != nil {
		return "", err
	}
	return prettyJSON.String(), nil
}

// printColoredJSON applies colors to different parts of a JSON string
func printColoredJSON(jsonStr string) {
	// keyColor := color.New(color.FgGreen).SprintFunc()
	stringColor := color.New(color.FgCyan).SprintFunc()
	numberColor := color.New(color.FgYellow).SprintFunc()
	boolColor := color.New(color.FgMagenta).SprintFunc()

	inQuotes := false
	escaped := false
	// isKey := false
	for _, r := range jsonStr {
		if escaped {
			escaped = false
			fmt.Print(stringColor(string(r)))
			continue
		}

		switch r {
		case '\\':
			escaped = true
			fmt.Print(stringColor(string(r)))
		case '"':
			// Toggle inQuotes state
			inQuotes = !inQuotes
			fmt.Print(stringColor(string(r)))

			// if !inQuotes {
			// If a colon follows the quoted section, it's a key
			// isKey = true
			// }
		case ':':
			// If we're out of quotes, it's the separator between key and value
			// isKey = false
			fmt.Print(string(r))
		case '{', '}', '[', ']':
			// Braces and brackets, print as-is
			fmt.Print(string(r))
		default:
			// Print either a key or a value, depending on the state
			if inQuotes {
				fmt.Print(stringColor(string(r)))
			} else if r >= '0' && r <= '9' || r == '.' || r == '-' {
				// Print numbers
				fmt.Print(numberColor(string(r)))
			} else if strings.Contains("truefalsenull", string(r)) {
				// Print boolean and null
				fmt.Print(boolColor(string(r)))
			} else {
				fmt.Print(string(r))
			}
		}
	}
	fmt.Println()
}
