package app_demo

import (
	"httpr2/apps"
	"net/http"
)

var WriteBasicPage0 = apps.CreateAppResponse(http.StatusConflict, "OK")

func WriteBasicPage(http.ResponseWriter, *http.Request) {
	//apps.CreateAppResponse()

}
