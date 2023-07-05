package themes

import (
	"fmt"
	"os"
	"strings"

//	"btctx/btc"
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

func (t *Theme) GetPath () string {
	return "themes/" + t.themeName + "/" + t.layoutName + "/"
}

func (t *Theme) getHtml (fileName string) string {
	fileBytes, err := os.ReadFile (t.GetPath () + fileName)
	if err != nil {
		fmt.Println (err.Error ())
		return ""
	}

	return string (fileBytes)
}

func (t *Theme) GetTxHtmlTemplate () string {
	return t.getHtml ("html/btc-objects/tx.html")
}

func (t *Theme) GetMinimizedInputHtmlTemplate () string {
	return t.getHtml ("html/btc-objects/input-minimized.html")
}

func (t *Theme) GetOutputHtmlTemplate (minimized bool) string {
	if minimized { return t.getHtml ("html/btc-objects/output-minimized.html") }
	return t.getHtml ("html/btc-objects/output-maximized.html")
}

func (t *Theme) GetExplorerPageHtml (queryId string, queryResult string, customJavascript string) string {
	pageHtml := t.getHtml ("html/pages/home.html")
	pageHtml = strings.Replace (pageHtml, "[[QUERY-ID]]", queryId, 1)
	pageHtml = strings.Replace (pageHtml, "[[QUERY-RESULT]]", queryResult, 1)

	layoutHtml := t.getHtml ("layout.html")
	layoutHtml = strings.Replace (layoutHtml, "[[CUSTOM-JAVASCRIPT]]", customJavascript, -1)
	layoutHtml = strings.Replace (layoutHtml, "[[LAYOUT-PATH]]", t.GetPath (), -1)
	layoutHtml = strings.Replace (layoutHtml, "[[LAYOUT-PAGE-CONTENT]]", pageHtml, 1)

	return layoutHtml
}

