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

	// redirect web requests if web enabled
	if (request.Method == "GET" || len (request.Method) == 0) && app.Settings.IsWebOn () {
		http.Redirect (response, request, app.Settings.GetFullUrl () + "/web", http.StatusMovedPermanently)
		return
	}

	// this request is not supported
	errorMessage := "Invalid REST URL. No rest version provided. No function name provided."
	fmt.Println (errorMessage)
	fmt.Fprint (response, errorMessage)
}

func main () {

	app.ParseSettings ()

	// test the connection
	restApi := rest.RestApiV1 {}
	currentBlockJson := restApi.GetCurrentBlockHeight ()
	if len (currentBlockJson) == 0 {
		fmt.Println ("Failed to connect to node.")
		return
	}

	// used only for testing
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
		mux.HandleFunc ("/web/", web.WebHandler)
	}

	mux.HandleFunc ("/rest/", rest.RestHandler)

	log.Fatal (http.ListenAndServe (app.Settings.GetBaseUrl (true), mux))
}
