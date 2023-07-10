package themes

import (
//	"fmt"
	"bytes"
	"strings"
	"html/template"

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
	if err := templ.ExecuteTemplate (&buff, "Layout", layoutData); err != nil {
		panic (err)
	}

	// return the html
	return buff.String ()
}

func (t *Theme) GetTxHtml (tx btc.Tx, customJavascript string) string {

	// get the data
	txPageHtmlData := tx.GetHtmlData ()
	explorerPageHtmlData := t.getExplorerPageHtmlData (tx.GetTxIdStr (), txPageHtmlData)
	layoutHtmlData := t.getLayoutHtmlData (customJavascript, explorerPageHtmlData)

	// parse the files
	layoutHtmlFiles := [] string {
		t.GetPath () + "html/layout.html",
		t.GetPath () + "html/page-explorer.html",
		t.GetPath () + "html/tx.html",
		t.GetPath () + "html/inputs-minimized.html",
		t.GetPath () + "html/input-maximized.html",
		t.GetPath () + "html/outputs-minimized.html",
		t.GetPath () + "html/output-maximized.html",
		t.GetPath () + "html/script.html",
		t.GetPath () + "html/segwit.html" }
	templ := template.Must (template.ParseFiles (layoutHtmlFiles...))

	// execute the templates
	var buff bytes.Buffer
	if err := templ.ExecuteTemplate (&buff, "Layout", layoutHtmlData); err != nil {
		panic (err)
	}

	// return the html
	return buff.String ()
}

func (t *Theme) GetPath () string {
	return "themes/" + t.themeName + "/" + t.layoutName + "/"
}

/*
func (t *Theme) getHtml (fileName string) string {
	fileBytes, err := os.ReadFile (t.GetPath () + "html/" + fileName)
	if err != nil {
		fmt.Println (err.Error ())
		return ""
	}

	return string (fileBytes)
}

func (t *Theme) GetTxHtmlTemplate () string {
	return t.getHtml ("html/tx.html")
}

func (t *Theme) GetInputHtmlTemplate (minimized bool) string {
	if minimized { return t.getHtml ("html/input-minimized.html") }
	return t.getHtml ("html/input-maximized.html")
}

func (t *Theme) GetMinimizedInputsTableHtmlTemplate () string {
	return t.getHtml ("html/inputs-minimized-table.html")
}

func (t *Theme) GetOutputHtmlTemplate (minimized bool) string {
	if minimized { return t.getHtml ("html/output-minimized.html") }
	return t.getHtml ("html/output-maximized.html")
}

func (t *Theme) GetMinimizedOutputsTableHtmlTemplate () string {
	return t.getHtml ("html/outputs-minimized-table.html")
}

func (t *Theme) GetScriptHtmlTemplate () string {
	return t.getHtml ("html/script.html")
}
*/

