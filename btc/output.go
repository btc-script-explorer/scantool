package btc

import (
	"fmt"
	"os"
	"strconv"
	"strings"
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

func (o *Output) getMinimizedHTMLTemplate () string {
	fileBytes, err := os.ReadFile ("./html/output-minimized.html")
	if err != nil { fmt.Println (err.Error ()); return "" }
	return string (fileBytes)
}

func (o *Output) GetMinimizedHTML (outputIndex int) string {

	html := o.getMinimizedHTMLTemplate ()

	html = strings.Replace (html, "[[INDEX]]", strconv.Itoa (outputIndex), 1)
	html = strings.Replace (html, "[[TYPE]]", o.outputType, 1)
	html = strings.Replace (html, "[[VALUE]]", strconv.FormatUint (o.value, 10), 1)
	html = strings.Replace (html, "[[ADDRESS]]", o.address, 1)

	return html
}

