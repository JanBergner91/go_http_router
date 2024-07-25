package apps

import (
	"httpr2/mw_template"
	"net/http"
)

type App func(http.ResponseWriter, *http.Request)

var Default400 = CreateStatusPage("defaultErrors.html", "./html-templates", 400, struct {
	ErrorCode string
	ErrorName string
	ErrorMSG  string
}{ErrorCode: "400", ErrorName: "Bad Request", ErrorMSG: "The 400 (Bad Request) status code indicates that the server cannot or will not process the request due to something that is perceived to be a client error (e.g., malformed request syntax, invalid request message framing, or deceptive request routing)."})

var Default401 = CreateStatusPage("defaultErrors.html", "./html-templates", 401, struct {
	ErrorCode string
	ErrorName string
	ErrorMSG  string
}{ErrorCode: "401", ErrorName: "Unauthorized", ErrorMSG: "The 401 (Unauthorized) status code indicates that the request has not been applied because it lacks valid authentication credentials for the target resource. The server generating a 401 response MUST send a WWW-Authenticate header field containing at least one challenge applicable to the target resource."})

var Default402 = CreateStatusPage("defaultErrors.html", "./html-templates", 402, struct {
	ErrorCode string
	ErrorName string
	ErrorMSG  string
}{ErrorCode: "402", ErrorName: "Payment Required", ErrorMSG: "The 402 (Payment Required) status code is reserved for future use."})

var Default403 = CreateStatusPage("defaultErrors.html", "./html-templates", 403, struct {
	ErrorCode string
	ErrorName string
	ErrorMSG  string
}{ErrorCode: "403", ErrorName: "Forbidden", ErrorMSG: "The 403 (Forbidden) status code indicates that the server understood the request but refuses to fulfill it. A server that wishes to make public why the request has been forbidden can describe that reason in the response content (if any)."})

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

var Default406 = CreateStatusPage("defaultErrors.html", "./html-templates", 406, struct {
	ErrorCode string
	ErrorName string
	ErrorMSG  string
}{ErrorCode: "406", ErrorName: "Not Acceptable", ErrorMSG: "The 406 (Not Acceptable) status code indicates that the target resource does not have a current representation that would be acceptable to the user agent, according to the proactive negotiation header fields received in the request, and the server is unwilling to supply a default representation."})

var Default407 = CreateStatusPage("defaultErrors.html", "./html-templates", 407, struct {
	ErrorCode string
	ErrorName string
	ErrorMSG  string
}{ErrorCode: "407", ErrorName: "Proxy Authentication Required", ErrorMSG: "The 407 (Proxy Authentication Required) status code is similar to 401 (Unauthorized), but it indicates that the client needs to authenticate itself in order to use a proxy for this request. The proxy MUST send a Proxy-Authenticate header field containing a challenge applicable to that proxy for the request. The client MAY repeat the request with a new or replaced Proxy-Authorization header field."})

var Default408 = CreateStatusPage("defaultErrors.html", "./html-templates", 408, struct {
	ErrorCode string
	ErrorName string
	ErrorMSG  string
}{ErrorCode: "408", ErrorName: "Request Timeout", ErrorMSG: "The 408 (Request Timeout) status code indicates that the server did not receive a complete request message within the time that it was prepared to wait."})

var Default409 = CreateStatusPage("defaultErrors.html", "./html-templates", 409, struct {
	ErrorCode string
	ErrorName string
	ErrorMSG  string
}{ErrorCode: "409", ErrorName: "Conflict", ErrorMSG: "The 409 (Conflict) status code indicates that the request could not be completed due to a conflict with the current state of the target resource. This code is used in situations where the user might be able to resolve the conflict and resubmit the request. The server SHOULD generate content that includes enough information for a user to recognize the source of the conflict. Conflicts are most likely to occur in response to a PUT request. For example, if versioning were being used and the representation being PUT included changes to a resource that conflict with those made by an earlier (third-party) request, the origin server might use a 409 response to indicate that it can't complete the request. In this case, the response representation would likely contain information useful for merging the differences based on the revision history."})

var Default500 = CreateStatusPage("defaultErrors.html", "./html-templates", 500, struct {
	ErrorCode string
	ErrorName string
	ErrorMSG  string
}{ErrorCode: "500", ErrorName: "Internal Server Error", ErrorMSG: "The 500 (Internal Server Error) status code indicates that the server encountered an unexpected condition that prevented it from fulfilling the request."})

var Default501 = CreateStatusPage("defaultErrors.html", "./html-templates", 501, struct {
	ErrorCode string
	ErrorName string
	ErrorMSG  string
}{ErrorCode: "501", ErrorName: "Not Implemented", ErrorMSG: "The 501 (Not Implemented) status code indicates that the server does not support the functionality required to fulfill the request. This is the appropriate response when the server does not recognize the request method and is not capable of supporting it for any resource."})

var Default502 = CreateStatusPage("defaultErrors.html", "./html-templates", 502, struct {
	ErrorCode string
	ErrorName string
	ErrorMSG  string
}{ErrorCode: "502", ErrorName: "Bad Gateway", ErrorMSG: "The 502 (Bad Gateway) status code indicates that the server, while acting as a gateway or proxy, received an invalid response from an inbound server it accessed while attempting to fulfill the request."})

var Default503 = CreateStatusPage("defaultErrors.html", "./html-templates", 503, struct {
	ErrorCode string
	ErrorName string
	ErrorMSG  string
}{ErrorCode: "503", ErrorName: "Service Unavailable", ErrorMSG: "The 503 (Service Unavailable) status code indicates that the server is currently unable to handle the request due to a temporary overload or scheduled maintenance, which will likely be alleviated after some delay. The server MAY send a Retry-After header field to suggest an appropriate amount of time for the client to wait before retrying the request."})

var Default504 = CreateStatusPage("defaultErrors.html", "./html-templates", 504, struct {
	ErrorCode string
	ErrorName string
	ErrorMSG  string
}{ErrorCode: "504", ErrorName: "Gateway Timeout", ErrorMSG: "The 504 (Gateway Timeout) status code indicates that the server, while acting as a gateway or proxy, did not receive a timely response from an upstream server it needed to access in order to complete the request."})

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
