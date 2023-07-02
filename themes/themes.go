package themes

import (
	"strings"

	"btctx/app"
)

type UiTheme interface {
//	GetLayout (pageContent string, javascript string) string
	GetHomePageHtml () string
}

type Theme struct {
	themeName string
	layoutName string
	layoutHtml string
}

func GetTheme () UiTheme {
	settings := app.GetSettings ()

	return Theme { themeName: strings.ToLower (settings.App.GetTheme ()), layoutName: strings.ToLower (settings.App.GetLayout ()), layoutHtml: "" }
}

func (t *Theme) getRootPath () string {
	return "themes/" + t.themeName + "/"
}

func (t *Theme) getLayoutPath () string {
	return t.getRootPath () + "html/" + t.layoutName + "/"
}

func (t Theme) GetHomePageHtml () string {
	html, err := os.ReadFile ("./html/page-home.html")
	if err != nil {
		fmt.Println (err.Error ())
		return
	}
}

