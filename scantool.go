package main

/*
Other explorers

https://www.lopp.net/bitcoin-information/block-explorers.html
*/

import (
	"fmt"
//"runtime"
	"net/http"
	"log"

	"github.com/btc-script-explorer/scantool/app"
	"github.com/btc-script-explorer/scantool/btc"
	"github.com/btc-script-explorer/scantool/rest"
	"github.com/btc-script-explorer/scantool/web"
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

func printListeningMessage () {

	nodeProxy, err := btc.GetNodeProxy ()
	if err != nil {
		fmt.Println (err.Error ())
		return
	}

	// create the data lines of the message
	lines := make ([] string, 0)
	lines = append (lines, "")
	lines = append (lines, "SCANTOOL " + app.GetVersion ())
	lines = append (lines, "")
	lines = append (lines, nodeProxy.GetNodeVersion ())
	lines = append (lines, app.Settings.GetNodeFullUrl ())
	lines = append (lines, "")
	lines = append (lines, "Web Access:")
	lines = append (lines, app.Settings.GetFullUrl () + "/web/")
	lines = append (lines, "")
	lines = append (lines, "REST API Example:")
	lines = append (lines, "curl -X GET " + app.Settings.GetFullUrl () + "/rest/v2/current_block_height")
	lines = append (lines, "")

	// calculate the width of the message and add padding as necessary
	bannerWidth := 0
	for l := 0; l < len (lines); l++ {
		if len (lines [l]) % 2 != 0 {
			lines [l] += " "
		}

		if len (lines [l]) + 6 > bannerWidth {
			bannerWidth = len (lines [l]) + 6
		}
	}

	topAndBottom := ""
	for a := 0; a < bannerWidth; a++ {
		topAndBottom += "*"
	}

	// pad the ones that need to be padded
	for l := 0; l < len (lines); l++ {
		for len (lines [l]) < bannerWidth - 2 {
			lines [l] = " " + lines [l] + " "
		}
	}

	// print the message
	fmt.Println ()
	fmt.Println (topAndBottom)
	for l := 0; l < len (lines); l++ {
		fmt.Println ("*" + lines [l] + "*")
	}
	fmt.Println (topAndBottom)
	fmt.Println ()
}

func main () {

	app.ParseSettings ()
	if app.Settings.IsVersionRequest () {
		fmt.Println (fmt.Sprintf ("scantool %s", app.GetVersion ()))
		return
	}

	proxy, err := btc.StartNodeProxy ()
_=proxy
	if err != nil {
		fmt.Println (err.Error ())
		return
	}

	printListeningMessage ()

	mux := http.NewServeMux ()

	mux.HandleFunc ("/", homeHandler)

	if app.Settings.IsWebOn () {
		mux.HandleFunc ("/favicon.ico", web.ServeFile)
		mux.HandleFunc ("/css/", web.ServeFile)
		mux.HandleFunc ("/js/", web.ServeFile)
		mux.HandleFunc ("/web/", web.WebHandler)
	}

	mux.HandleFunc ("/rest/", rest.RestHandler)

//var buf [] byte
//runtime.Stack (&buf, true)

	log.Fatal (http.ListenAndServe (app.Settings.GetBaseUrl (true), mux))
}
