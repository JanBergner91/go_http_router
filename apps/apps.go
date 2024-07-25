package apps

import (
	"httpr2/mw_template"
	"net/http"
)

type App func(http.ResponseWriter, *http.Request)

var Default404 = CreateStatusPage("defaultErrors.html", "./html-templates", 404, struct {
	ErrorCode string
	ErrorName string
	ErrorMSG  string
}{ErrorCode: "404", ErrorName: "Not Found", ErrorMSG: "The 404 (Not Found) status code indicates that the origin server did not find a current representation for the target resource or is not willing to disclose that one exists. A 404 status code does not indicate whether this lack of representation is temporary or permanent; the 410 (Gone) status code is preferred over 404 if the origin server knows, presumably through some configurable means, that the condition is likely to be permanent."})

var Default405 = CreateStatusPage("defaultErrors.html", "./html-templates", 405, struct {
	ErrorCode string
	ErrorName string
	ErrorMSG  string
}{ErrorCode: "405", ErrorName: "Method Not Allowed", ErrorMSG: "The 405 (Method Not Allowed) status code indicates that the method received in the request-line is known by the origin server but not supported by the target resource. The origin server MUST generate an Allow header field in a 405 response containing a list of the target resource's currently supported methods."})

var Default500 = CreateStatusPage("defaultErrors.html", "./html-templates", 500, struct {
	ErrorCode string
	ErrorName string
	ErrorMSG  string
}{ErrorCode: "500", ErrorName: "Internal Server Error", ErrorMSG: "The 500 (Internal Server Error) status code indicates that the server encountered an unexpected condition that prevented it from fulfilling the request."})

func CreateStatusPage(tplname, tpldir string, statuscode int, data interface{}) App {
	return func(next_w http.ResponseWriter, next_r *http.Request) {
		mw_template.ProcessTemplate(next_w, tplname, tpldir, statuscode, data)
	}

}

func CreateAppResponse(statuscode int, statusmsg string) App {
	return func(next_w http.ResponseWriter, next_r *http.Request) {
		next_w.WriteHeader(statuscode)
		next_w.Write([]byte(statusmsg))
	}
}
