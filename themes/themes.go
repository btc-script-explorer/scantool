package themes

import (
//	"fmt"
	"bytes"
	"strings"
	"html/template"

	"btctx/app"
	"btctx/btc"
)

type Theme struct {
	themeName string
	layoutName string
}

func GetTheme (themeName string, layoutName string) Theme {
	return Theme { themeName: strings.ToLower (themeName), layoutName: strings.ToLower (layoutName) }
}

func GetThemeForUserAgent (userAgent string) Theme {
	return GetTheme ("default", "desktop")
}

func (t *Theme) getExplorerPageHtmlData (queryText string, queryResults map [string] interface {}) map [string] interface {} {
	explorerPageData := make (map [string] interface {})
	explorerPageData ["QueryText"] = queryText
	if queryResults != nil {
		explorerPageData ["QueryResults"] = queryResults
	}

	return explorerPageData
}

func (t *Theme) getLayoutHtmlData (customJavascript string, explorerPageData map [string] interface {}) map [string] interface {} {
	layoutData := make (map [string] interface {})
	layoutData ["CustomJavascript"] = template.HTML (`<script type="text/javascript">` + customJavascript + "</script>")
	layoutData ["ExplorerPage"] = explorerPageData

	nodeClient := btc.GetNodeClient ()
	layoutData ["NodeVersion"] = nodeClient.GetVersionString ()

	settings := app.GetSettings ()
	layoutData ["NodeUrl"] = settings.Node.GetFullUrl ()

	return layoutData
}

func (t *Theme) GetExplorerPageHtml () string {

	// get the data
	explorerPageData := t.getExplorerPageHtmlData ("", nil)
	layoutData := t.getLayoutHtmlData ("", explorerPageData)

	// parse the files
	files := [] string {
		t.GetPath () + "html/layout.html",
		t.GetPath () + "html/page-explorer.html" }

	templ := template.Must (template.ParseFiles (files...))
	templ.Parse (`{{ define "QueryResults" }}{{ end }}`)

	// execute the templates
	var buff bytes.Buffer
	if err := templ.ExecuteTemplate (&buff, "Layout", layoutData); err != nil { panic (err) }

	// return the html
	return buff.String ()
}

func (t *Theme) GetBlockHtml (block btc.Block, customJavascript string) string {

	// get the data
	blockHtmlData := block.GetHtmlData ()
	explorerPageHtmlData := t.getExplorerPageHtmlData (block.GetHash (), blockHtmlData)
	layoutHtmlData := t.getLayoutHtmlData (customJavascript, explorerPageHtmlData)

	// parse the files
	layoutHtmlFiles := [] string {
		t.GetPath () + "html/layout.html",
		t.GetPath () + "html/page-explorer.html",
		t.GetPath () + "html/block.html" }
	templ := template.Must (template.ParseFiles (layoutHtmlFiles...))

	// execute the templates
	var buff bytes.Buffer
	if err := templ.ExecuteTemplate (&buff, "Layout", layoutHtmlData); err != nil { panic (err) }

	// return the html
	return buff.String ()
}

func (t *Theme) GetTxHtml (tx btc.Tx, customJavascript string) string {

	// get the data
	txPageHtmlData := tx.GetHtmlData ()
	explorerPageHtmlData := t.getExplorerPageHtmlData (tx.GetTxId (), txPageHtmlData)
	layoutHtmlData := t.getLayoutHtmlData (customJavascript, explorerPageHtmlData)

	// parse the files
	layoutHtmlFiles := [] string {
		t.GetPath () + "html/layout.html",
		t.GetPath () + "html/page-explorer.html",
		t.GetPath () + "html/tx.html",
		t.GetPath () + "html/input-minimized.html",
		t.GetPath () + "html/input-maximized.html",
		t.GetPath () + "html/output-minimized.html",
		t.GetPath () + "html/output-maximized.html",
		t.GetPath () + "html/field-set.html" }
	templ := template.Must (template.ParseFiles (layoutHtmlFiles...))

	// execute the templates
	var buff bytes.Buffer
	if err := templ.ExecuteTemplate (&buff, "Layout", layoutHtmlData); err != nil { panic (err) }

	// return the html
	return buff.String ()
}

func (t *Theme) GetPreviousOutputScriptHtml (script btc.Script, htmlId string, displayTypeClassPrefix string) string {

	// get the data
	previousOutputHtmlData := script.GetHtmlData (htmlId, displayTypeClassPrefix)
	
	// parse the file
	layoutHtmlFiles := [] string {
		t.GetPath () + "html/field-set.html" }
	templ := template.Must (template.ParseFiles (layoutHtmlFiles...))

	// execute the template
	var buff bytes.Buffer
	if err := templ.ExecuteTemplate (&buff, "FieldSet", previousOutputHtmlData.FieldSet); err != nil { panic (err) }

	// return the html
	return buff.String ()
}

func (t *Theme) GetPath () string {
	return "themes/" + t.themeName + "/" + t.layoutName + "/"
}

