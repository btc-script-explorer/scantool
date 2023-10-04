package main

/*
Other explorers

https://www.lopp.net/bitcoin-information/block-explorers.html
*/

import (
	"fmt"
	"net/http"
	"log"

	"github.com/btc-script-explorer/scantool/app"
	"github.com/btc-script-explorer/scantool/btc/node"
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

	nodeProxy, _ := node.GetNodeProxy ()

	// create the data lines of the message
	messageLines := make ([] string, 0)

	messageLines = append (messageLines, "")
	messageLines = append (messageLines, "SCANTOOL " + app.GetVersion ())
	messageLines = append (messageLines, "")

	messageLines = append (messageLines, "Node: " + nodeProxy.GetNodeVersion ())
	messageLines = append (messageLines, "      " + app.Settings.GetNodeFullUrl ())
	messageLines = append (messageLines, "")

	webLine := " Web: "; if app.Settings.IsWebOn () { webLine += app.Settings.GetFullUrl () + "/web/" } else { webLine += "Off" }
	messageLines = append (messageLines, webLine)
	messageLines = append (messageLines, "")

	messageLines = append (messageLines, "REST: curl -X GET " + app.Settings.GetFullUrl () + "/rest/v1/current_block_height")
	messageLines = append (messageLines, "      curl -X POST -d '{}' " + app.Settings.GetFullUrl () + "/rest/v1/block")
	messageLines = append (messageLines, "")

	settingsLineCaching := "Caching: O"; if app.Settings.IsCachingOn () { settingsLineCaching += "n" } else { settingsLineCaching += "ff" }
	messageLines = append (messageLines, settingsLineCaching)
	messageLines = append (messageLines, "")

	// first make sure every line is an even number of characters
	for l := 0; l < len (messageLines); l++ {
		if len (messageLines [l]) % 2 != 0 {
			messageLines [l] += " "
		}
	}

	// get the maximum length, starting with the fourth line
	maxLineLen := 0
	for l := 3; l < len (messageLines); l++ {
		if len (messageLines [l]) > maxLineLen {
			maxLineLen = len (messageLines [l])
		}
	}

	// add padding so the lines are all the same length, starting with the fourth line
	for l := 3; l < len (messageLines); l++ {
		for c := len (messageLines [l]) - 1; c < maxLineLen; c++ {
			messageLines [l] += " "
		}
	}

	// calculate the width of the message and add padding as necessary
	bannerWidth := 0
	for l := 0; l < len (messageLines); l++ {

		// make sure every line is an even number of characters
		if len (messageLines [l]) % 2 != 0 {
			messageLines [l] += " "
		}

		if len (messageLines [l]) + 6 > bannerWidth {
			bannerWidth = len (messageLines [l]) + 6
		}
	}

	topAndBottom := ""
	for a := 0; a < bannerWidth; a++ {
		topAndBottom += "*"
	}

	// pad the ones that need to be padded
	for l := 0; l < len (messageLines); l++ {
		for len (messageLines [l]) < bannerWidth - 2 {
			messageLines [l] = " " + messageLines [l] + " "
		}
	}

	// print the message
	fmt.Println ()
	fmt.Println (topAndBottom)
	for l := 0; l < len (messageLines); l++ {
		fmt.Println ("*" + messageLines [l] + "*")
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

	// make sure the node is connected and start the cache if it is being used
	_, err := node.GetNodeProxy ()
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

	log.Fatal (http.ListenAndServe (app.Settings.GetBaseUrl (true), mux))
}
