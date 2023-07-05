package btc

import (
	"strconv"
	"strings"

	"btctx/themes"
)

type Output struct {
	value uint64
	outputScript Script
	outputType string
	address string
}

func (o *Output) GetSatoshis () uint64 {
	return o.value
}

func (o *Output) GetOutputScript () Script {
	return o.outputScript
}

func (o *Output) GetOutputType () string {
	return o.outputType
}

func (o *Output) GetAddress () string {
	return o.address
}

func (o *Output) GetMinimizedHtml (outputIndex int, theme themes.Theme) string {

	html := theme.GetMinimizedOutputHtmlTemplate ()

	html = strings.Replace (html, "[[INDEX]]", strconv.Itoa (outputIndex), 1)
	html = strings.Replace (html, "[[TYPE]]", o.outputType, 1)
	html = strings.Replace (html, "[[VALUE]]", strconv.FormatUint (o.value, 10), 1)
	html = strings.Replace (html, "[[ADDRESS]]", o.address, 1)

	return html
}

