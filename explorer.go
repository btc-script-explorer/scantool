package main

/*
Other explorers

https://www.lopp.net/bitcoin-information/block-explorers.html
*/

import (
	"fmt"
	"os"
	"net/http"
	"log"

	"btctx/app"
	"btctx/rest"
	"btctx/web"

	"btctx/test"
)

func homeHandler (response http.ResponseWriter, request *http.Request) {

	// redirect web requests to most recent version if web enabled
	if app.Settings.IsWebOn () {
		http.Redirect (response, request, app.Settings.GetFullUrl () + "/web", http.StatusMovedPermanently)
		return
	}

	// web is disabled, this request is not supported
	errorMessage := "Invalid REST URL. Web server is turned off."
	fmt.Println (errorMessage)
	fmt.Fprint (response, errorMessage)
}

func main () {

	app.ParseSettings ()

	if app.Settings.GetTestMode () != "" {
		test.RunTests ()
		os.Exit (0)
	}

	app.Settings.PrintListeningMessage ()

	mux := http.NewServeMux ()

	mux.HandleFunc ("/", homeHandler)

	if app.Settings.IsWebOn () {
		mux.HandleFunc ("/favicon.ico", web.ServeFile)
		mux.HandleFunc ("/css/", web.ServeFile)
		mux.HandleFunc ("/js/", web.ServeFile)
		mux.HandleFunc ("/image/", web.ServeFile)
		mux.HandleFunc ("/web/", web.WebHandler)
	}

	mux.HandleFunc ("/rest/", rest.RestHandler)

	log.Fatal (http.ListenAndServe (app.Settings.GetBaseUrl (), mux))
}
