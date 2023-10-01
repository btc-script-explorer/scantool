package web

import (
	"fmt"
	"net/http"
	"time"
	"io"
//	"sort"
	"strings"
	"strconv"
	"encoding/json"
	"encoding/hex"
	"bytes"
	"html/template"

//	"github.com/go-echarts/go-echarts/v2/charts"
//	"github.com/go-echarts/go-echarts/v2/opts"

	"github.com/btc-script-explorer/scantool/app"
	"github.com/btc-script-explorer/scantool/btc"
	"github.com/btc-script-explorer/scantool/btc/node"
	"github.com/btc-script-explorer/scantool/rest"
)

// html template structs

type ElementTypeHTML struct {
	Label string
	Count uint16
	Percent string
}

type FieldHtmlData struct {
	DisplayText template.HTML
	ShowCopyButton bool
	CopyText string
}

type FieldSetHtmlData struct {
	HtmlId string
	DisplayTypeClassPrefix string
	CharWidth uint
	HexFields [] FieldHtmlData
	TextFields [] FieldHtmlData
	TypeFields [] FieldHtmlData
	CopyImageUrl string
}

type ScriptHtmlData struct {
	FieldSet FieldSetHtmlData
	IsNil bool
	IsOrdinal bool
}

type InputHtmlData struct {
	InputIndex uint16
	DisplayTypeClassPrefix string
	IsCoinbase bool
	SpendType string
	ValueIn template.HTML
	BaseUrl string
	PreviousOutputTxId string
	PreviousOutputIndex uint16
	PreviousOutputAddress string
	PreviousOutputScript ScriptHtmlData
	Sequence uint32
	InputScript ScriptHtmlData
	RedeemScript ScriptHtmlData
	WitnessScript ScriptHtmlData
	TapScript ScriptHtmlData
	Bip141 bool
	Segwit SegwitHtmlData
}

type OutputHtmlData struct {
	OutputIndex uint16
	DisplayTypeClassPrefix string
	OutputType string
	Value template.HTML
	Address string
	OutputScript ScriptHtmlData
}

type SegwitHtmlData struct {
	FieldSet FieldSetHtmlData
	WitnessScript ScriptHtmlData
	TapScript ScriptHtmlData
	IsEmpty bool
}

// JSON structs
/*
type previousOutputJsonOut struct {
	InputTxId string
	InputIndex uint16
	PrevOutValue uint64
	PrevOutType string
	PrevOutAddress string
	PrevOutScriptHtml string
}

type BlockChartData struct {
	NonCoinbaseInputCount uint16
	OutputCount uint16
	SpendTypes map [string] uint16
	OutputTypes map [string] uint16
}
*/

// return an array of possible query types based on the search parameter
func determineQueryTypes (queryParam string) [] string {

	paramLen := len (queryParam)
	if paramLen == 64 {
		// it is a block or transaction hash
		_, err := hex.DecodeString (queryParam)
		if err != nil { fmt.Println (queryParam + " is not a valid hex string."); return [] string {} }
		return [] string { "tx", "block" }
	} else {
		// it could be a block height
		_, err := strconv.ParseUint (queryParam, 10, 32)
		if err != nil { fmt.Println (queryParam + " is not a valid block height."); return [] string {} }
		return [] string { "block" }
	}

	return [] string {}
}

func WebHandler (response http.ResponseWriter, request *http.Request) {

	modifiedPath := request.URL.Path
	if len (modifiedPath) > 0 && modifiedPath [0] == '/' { modifiedPath = modifiedPath [1:] }
	lastChar := len (modifiedPath) - 1
	if lastChar >= 0 && modifiedPath [lastChar] == '/' { modifiedPath = modifiedPath [: lastChar] }

	if len (modifiedPath) == 0 {
		errorMessage := fmt.Sprintf ("Invalid request: %s", request.URL.Path)
		fmt.Println (errorMessage)
		fmt.Fprint (response, errorMessage)
		return
	}

	requestParts := strings.Split (modifiedPath, "/")
	if requestParts [0] != "web" {
		errorMessage := fmt.Sprintf ("Invalid request: %s", request.URL.Path)
		fmt.Println (errorMessage)
		fmt.Fprint (response, errorMessage)
		return
	}

	params := requestParts [1:]
	paramCount := len (params)

	// web site currently using rest version 2
	restApi := rest.RestApiV2 {}

	nodeProxy, err := node.GetNodeProxy ()
	if err != nil {
		fmt.Println (err.Error ())
		return
	}

	html := ""
	customJavascript := fmt.Sprintf ("var base_url_web = '%s/web';\n", app.Settings.GetFullUrl ())
	customJavascript += fmt.Sprintf ("var base_url_rest = '%s/rest/v%d';\n", app.Settings.GetFullUrl (), restApi.GetVersion ())

	// about page
	if paramCount >= 1 && params [0] == "about" {
		fmt.Fprint (response, getAboutPageHtml (customJavascript))
		return
	}

	// here, a determination is made as to what the user is requesting by examining the parameters received
	possibleQueryTypes := [] string { "block" } // default query type
	if paramCount >= 1 { possibleQueryTypes [0] = params [0] }

	if possibleQueryTypes [0] == "search" {
		if paramCount < 2 { fmt.Println ("No search parameter provided."); return }
		possibleQueryTypes = determineQueryTypes (params [1])
	}

	for _, queryType := range possibleQueryTypes {
		switch queryType {

			// returns json
			case "block":

				if request.Method != "GET" { fmt.Println (fmt.Sprintf ("%s must be sent as a GET request.", queryType)); break }

				blockRequest := node.BlockRequest {}

				// get the block request data
				if paramCount >= 2 && len (params [1]) > 0 {
					blockRequest.BlockKey = params [1]
				}

//				options := make (map [string] interface {})
//				blockRequestData ["options"] = options

				block := nodeProxy.GetBlock (blockRequest)

				txIds := block.GetTxIds ()

				var txIdsBytes [] byte
				txIdsBytes, err = json.Marshal (txIds)
				if err != nil { fmt.Println (err.Error ()) }

				customJavascript += fmt.Sprintf ("var block_tx_ids = JSON.parse ('%s');\n", string (txIdsBytes))
				html = getBlockHtml (block, customJavascript)


			// block-tx is for the web interface to get HTML segments in real time
			// returns json
			case "block-tx":

				if request.Method != "GET" { fmt.Println (fmt.Sprintf ("%s must be sent as a GET request.", queryType)); break }

				// check the parameters
				if paramCount < 3 { fmt.Println ("No id provided for tx. Request ignored."); return }
				if len (params [1]) != 64 { fmt.Println (fmt.Sprintf ("%s is not a valid tx id", params [1])); return }

				blockIndex, err := strconv.Atoi (params [2])
				if err != nil { fmt.Println (fmt.Sprintf ("block index (%s) not formatted correctly, error: %s", params [2], err.Error ())); return }

				txRequest := node.TxRequest { TxId: params [1] }
				tx := nodeProxy.GetTx (txRequest)

				blockTxResponse := getBlockTxResponse (tx, uint16 (blockIndex))

				jsonBytes, err := json.Marshal (blockTxResponse)
				if err != nil { fmt.Println (err.Error ()) }

				fmt.Fprint (response, string (jsonBytes))

				return


			// returns html
			case "tx":

				if request.Method != "GET" { fmt.Println (fmt.Sprintf ("%s must be sent as a GET request.", queryType)); break }

				if paramCount < 2 { fmt.Println ("No id provided for tx. Request ignored."); return }
				if len (params [1]) != 64 { fmt.Println (fmt.Sprintf ("%s is not a valid tx id", params [1])); return }

				txRequest := node.TxRequest { TxId: params [1] }

				tx := nodeProxy.GetTx (txRequest)
				if tx.IsNil () {
					html = "Empty response from server."
					break
				}

				javascriptInputs := ""
				for i := uint16 (0); i < tx.GetInputCount (); i++ {
					if len (javascriptInputs) > 0 { javascriptInputs += "," }
					javascriptInputs += fmt.Sprintf ("{tx_id:\"%s\",input_index:%d}", params [1], i)
				}

				customJavascript += fmt.Sprintf ("var tx_inputs = [%s];", javascriptInputs)
				html = getTxHtml (tx, customJavascript)


//			case "address": // would probably require an electrum server for implementation


			// returns json
			case "input":

				if request.Method != "POST" { fmt.Println (fmt.Sprintf ("%s must be sent as a POST request.", queryType)); break }

				// get the parameters
				var params map [string] interface {}
				paramsBytes, err := io.ReadAll (request.Body)
				if err != nil {
					fmt.Println (err.Error ())
					fmt.Fprint (response, "")
					return
				}

				err = json.Unmarshal (paramsBytes, &params)
				if err != nil {
					fmt.Println (err.Error ())
					fmt.Fprint (response, "")
					return
				}

				if params ["tx_id"] == nil {
					fmt.Println ("tx_id parameter not found")
					fmt.Fprint (response, "")
					return
				}

				if params ["input_index"] == nil {
					fmt.Println ("input_index parameter not found")
					fmt.Fprint (response, "")
					return
				}

				txId := params ["tx_id"].(string)
				inputIndex := uint16 (params ["input_index"].(float64))

				// check for error from node in json
				if len (txId) != 64 {
					fmt.Println (fmt.Sprintf ("No tx id provided for input. Request ignored."))
					fmt.Fprint (response, "")
					return
				}

				// get the tx
				txRequest := node.TxRequest { TxId: txId }
				tx := nodeProxy.GetTx (txRequest)

				// check for errors
				if tx.IsNil () {
					fmt.Println (fmt.Sprintf ("Tx %s could not be found.", txId))
					fmt.Fprint (response, "")
					return
				}


				if inputIndex >= tx.GetInputCount () {
					fmt.Println (fmt.Sprintf ("Tx %s does not have an input %d.", txId, inputIndex))
					fmt.Fprint (response, "")
					return
				}

				// get the input
				input := tx.GetInput (inputIndex)
				var valueIn uint64
				var address string
				if input.IsCoinbase () {
					// value in is the total of all outputs for coinbase inputs
					valueIn = 0
					for _, output := range tx.GetOutputs () {
						valueIn += output.GetValue ()
					}
				} else {
					outputRequest := node.OutputRequest { TxId: input.GetPreviousOutputTxId (), OutputIndex: input.GetPreviousOutputIndex () }
					previousOutput := nodeProxy.GetOutput (outputRequest)
					input.SetPreviousOutput (previousOutput)
					address = previousOutput.GetAddress ()

					// value in comes from the previous output for non-coinbase inputs
					valueIn = previousOutput.GetValue ()
				}

				// return the response
				inputHtmlData := getInputHtmlData (input, inputIndex, valueIn, tx.SupportsBip141 ())
				inputHtml := getInputHtml (inputHtmlData)

				jsonInput := make (map [string] interface {})
				jsonInput ["spend_type"] = input.GetSpendType ()
				jsonInput ["address"] = address
				jsonInput ["value_in"] = valueIn
				jsonInput ["input_html"] = inputHtml

				jsonBytes, err := json.Marshal (jsonInput)
				if err != nil { fmt.Println (err) }

				fmt.Fprint (response, string (jsonBytes))
				return


/*
			case "block_charts":

				if request.Method != "POST" { fmt.Println (fmt.Sprintf ("%s must be sent as a POST request.", queryType)); return }

				// unpack the json
				var blockChartData BlockChartData
				err := json.NewDecoder (request.Body).Decode (&blockChartData)
				if err != nil { fmt.Println (err.Error()) }

				blockCharts := getBlockCharts (blockChartData.NonCoinbaseInputCount, blockChartData.OutputCount, blockChartData.SpendTypes, blockChartData.OutputTypes)

				chartsBytes, err := json.Marshal (blockCharts)
				if err != nil { fmt.Println (err.Error ()); return }

				fmt.Fprint (response, string (chartsBytes))
				return
*/
		}

		if len (html) > 0 { break }
	}

	if len (html) == 0 {
		html = getExplorerPageHtml ()
	}

	fmt.Fprint (response, html)
}

func ServeFile (response http.ResponseWriter, request *http.Request) {

	if request.URL.Path == "/favicon.ico" { return }
	http.ServeFile (response, request, GetPath () + request.URL.Path)
}

func getKnownSpendTypesHtmlData (nonCoinbaseInputCount uint16, knownSpendTypeMap map [string] uint16) ([] ElementTypeHTML, uint16) {

	var knownSpendTypes [] ElementTypeHTML
	knownSpendTypeCount := uint16 (1) // starting at 1 because the coinbase input always has a known spend type
	for spendType, num := range knownSpendTypeMap {
		knownSpendTypeCount += num
		knownSpendTypes = append (knownSpendTypes, ElementTypeHTML { Label: spendType, Count: uint16 (num), Percent: fmt.Sprintf ("%9.2f%%", float32 (num * 100) / float32 (nonCoinbaseInputCount)) })
	}

	return knownSpendTypes, knownSpendTypeCount
}
func getOutputTypesHtmlData (outputCount uint16, outputTypeMap map [string] uint16) [] ElementTypeHTML {

	var outputTypes [] ElementTypeHTML
	for outputType, num := range outputTypeMap {
		outputTypes = append (outputTypes, ElementTypeHTML { Label: outputType, Count: uint16 (num), Percent: fmt.Sprintf ("%9.2f%%", float32 (num * 100) / float32 (outputCount)) })
	}

	return outputTypes
}

func getExplorerPageHtmlData (queryText string, queryResults map [string] interface {}) map [string] interface {} {
	explorerPageData := make (map [string] interface {})
	explorerPageData ["QueryText"] = queryText
	if queryResults != nil {
		explorerPageData ["QueryResults"] = queryResults
	}

	return explorerPageData
}

func getLayoutHtmlData (customJavascript string, explorerPageData map [string] interface {}) map [string] interface {} {
	layoutData := make (map [string] interface {})
	layoutData ["CustomJavascript"] = template.HTML (`<script type="text/javascript">` + customJavascript + "</script>")
	layoutData ["ExplorerPage"] = explorerPageData

	nodeProxy, err := node.GetNodeProxy ()
	if err != nil { fmt.Println (err.Error ()) }

	layoutData ["NodeVersion"] = template.HTML (strings.Replace (nodeProxy.GetNodeVersion (), " ", "&nbsp;", -1))
	layoutData ["NodeUrl"] = template.HTML (app.Settings.GetNodeFullUrl ())

	return layoutData
}

func getExplorerPageHtml () string {

	// get the data
	explorerPageData := getExplorerPageHtmlData ("", nil)
	layoutData := getLayoutHtmlData ("", explorerPageData)

	// parse the files
	files := [] string {
		GetPath () + "html/layout.html",
		GetPath () + "html/page-explorer.html" }

	templ := template.Must (template.ParseFiles (files...))
	templ.Parse (`{{ define "QueryResults" }}{{ end }}`)

	// execute the templates
	var buff bytes.Buffer
	if err := templ.ExecuteTemplate (&buff, "Layout", layoutData); err != nil { panic (err) }

	// return the html
	return buff.String ()
}

func getBlockHtml (block btc.Block, customJavascript string) string {

	// get the data
	blockHtmlData := make (map [string] interface {})

	blockHash := block.GetHash ()

	// get block data
	blockHtmlData ["BaseUrl"] = app.Settings.GetFullUrl () + "/web"
	blockHtmlData ["Height"] = block.GetHeight ()
	blockHtmlData ["Time"] = time.Unix (block.GetTimestamp (), 0).UTC ()
	blockHtmlData ["Hash"] = blockHash

	previousHash := block.GetPreviousHash ()
	if len (previousHash) > 0 { blockHtmlData ["PreviousHash"] = previousHash }
	nextHash := block.GetNextHash ()
	if len (nextHash) > 0 { blockHtmlData ["NextHash"] = nextHash }

	// create the html page
	explorerPageHtmlData := getExplorerPageHtmlData (blockHash, blockHtmlData)
	layoutHtmlData := getLayoutHtmlData (customJavascript, explorerPageHtmlData)

	// parse the files
	layoutHtmlFiles := [] string {
		GetPath () + "html/layout.html",
		GetPath () + "html/page-explorer.html",
		GetPath () + "html/type-chart-detail.html",
		GetPath () + "html/block.html" }
	templ := template.Must (template.ParseFiles (layoutHtmlFiles...))

	// execute the templates
	var buff bytes.Buffer
	if err := templ.ExecuteTemplate (&buff, "Layout", layoutHtmlData); err != nil { panic (err) }

	// return the html
	return buff.String ()
}

type BlockTxHtmlData struct {
	Id string
	Bip141 bool
	InputCount uint16
	OutputCount uint16
	BlockIndex uint16
	BaseUrl string
}

type BlockTxResponse struct {
	Bip141 bool `json:"bip141"`
	InputCount uint16 `json:"input_count"`
	OutputCount uint16 `json:"output_count"`
	TxHtml string `json:"tx_html"`
}

func getBlockTxResponse (tx btc.Tx, blockIndex uint16) BlockTxResponse {

	blockTxData := BlockTxHtmlData {	Id: tx.GetTxId (),
										Bip141: tx.SupportsBip141 (),
										InputCount: tx.GetInputCount (),
										OutputCount: tx.GetOutputCount (),
										BlockIndex: blockIndex,
										BaseUrl: app.Settings.GetFullUrl () + "/web" }

	// parse the file
	htmlFiles := [] string { GetPath () + "html/block-tx.html" }
	templ := template.Must (template.ParseFiles (htmlFiles...))

	// execute the template
	var buff bytes.Buffer
	if err := templ.ExecuteTemplate (&buff, "BlockTx", blockTxData); err != nil { panic (err) }

	blockTxResponse := BlockTxResponse {	Bip141: tx.SupportsBip141 (),
											InputCount: tx.GetInputCount (),
											OutputCount: tx.GetOutputCount (),
											TxHtml: buff.String () }
	return blockTxResponse
}

func getTxHtml (tx btc.Tx, customJavascript string) string {

	txPageHtmlData := make (map [string] interface {})

	// transaction data
	txPageHtmlData ["BaseUrl"] = app.Settings.GetFullUrl () + "/web"
	txPageHtmlData ["BlockTime"] = time.Unix (tx.GetBlockTime (), 0).UTC ()

	txPageHtmlData ["BlockHash"] = tx.GetBlockHash ()
	txPageHtmlData ["IsCoinbase"] = tx.IsCoinbase ()
	txPageHtmlData ["SupportsBip141"] = tx.SupportsBip141 ()
	txPageHtmlData ["LockTime"] = tx.GetLockTime ()

	// outputs
	totalOut := uint64 (0)
	outputs := tx.GetOutputs ()
	outputCount := len (outputs)

	outputCountLabel := fmt.Sprintf ("%d Output", outputCount)
	if outputCount > 1 { outputCountLabel += "s" }
	txPageHtmlData ["OutputCountLabel"] = outputCountLabel

	outputHtmlData := make ([] OutputHtmlData, outputCount)
	for o, output := range outputs {
		totalOut += output.GetValue ()
		scriptHtmlId := fmt.Sprintf ("output-script-%d", o)
		outputHtmlData [o] = getOutputHtmlData (outputs [o], scriptHtmlId, "", uint16 (o))
	}
	txPageHtmlData ["OutputData"] = outputHtmlData

	// totals for the transaction
	txPageHtmlData ["ValueOut"] = totalOut
	txPageHtmlData ["ValueIn"] = 0
	if tx.IsCoinbase () {
		txPageHtmlData ["ValueIn"] = totalOut
	}
	txPageHtmlData ["Fee"] = 0

	// inputs
	inputs := tx.GetInputs ()
	inputCount := len (inputs)

	inputCountLabel := fmt.Sprintf ("%d Input", inputCount)
	if inputCount > 1 { inputCountLabel += "s" }
	txPageHtmlData ["InputCountLabel"] = inputCountLabel

	inputHtmlData := make ([] InputHtmlData, inputCount)
	for i := uint16 (0); i < uint16 (inputCount); i++ {
//		valueIn := uint64 (0); if tx.IsCoinbase () && i == 0 { valueIn = totalOut }
//		inputHtmlData [i] = getInputHtmlData (inputs [i], i, valueIn, tx.SupportsBip141 ())
		inputHtmlData [i] = InputHtmlData { InputIndex: i }
	}
	txPageHtmlData ["InputData"] = inputHtmlData

	// add the tx html data to the page and layout html data
	explorerPageHtmlData := getExplorerPageHtmlData (tx.GetTxId (), txPageHtmlData)
	layoutHtmlData := getLayoutHtmlData (customJavascript, explorerPageHtmlData)

	// parse the files
	layoutHtmlFiles := [] string {
		GetPath () + "html/layout.html",
		GetPath () + "html/page-explorer.html",
		GetPath () + "html/tx.html",
		GetPath () + "html/input-minimized.html",
		GetPath () + "html/input-maximized.html",
		GetPath () + "html/output-minimized.html",
		GetPath () + "html/output-maximized.html",
		GetPath () + "html/field-set.html" }
	templ := template.Must (template.ParseFiles (layoutHtmlFiles...))

	// execute the templates
	var buff bytes.Buffer
	if err := templ.ExecuteTemplate (&buff, "Layout", layoutHtmlData); err != nil { panic (err) }

	// return the html
	return buff.String ()
}

func getInputHtml (htmlData InputHtmlData) string {

	htmlFiles := [] string {
		GetPath () + "html/input-detail.html",
		GetPath () + "html/field-set.html" }
	templ := template.Must (template.ParseFiles (htmlFiles...))

	var buff bytes.Buffer
	if err := templ.ExecuteTemplate (&buff, "InputDetail", htmlData); err != nil { panic (err) }

	// return the html
	return buff.String ()
}

func getAboutPageHtml (customJavascript string) string {
	layoutHtmlData := getLayoutHtmlData (customJavascript, map [string] interface {} { "AppVersion": app.GetVersion () })

	// parse the files
	layoutHtmlFiles := [] string {
		GetPath () + "html/layout.html",
		GetPath () + "html/page-about.html" }
	templ := template.Must (template.ParseFiles (layoutHtmlFiles...))

	// execute the templates
	var buff bytes.Buffer
	if err := templ.ExecuteTemplate (&buff, "Layout", layoutHtmlData); err != nil { panic (err) }

	// return the html
	return buff.String ()
}

func getInputHtmlData (input btc.Input, txIndex uint16, satoshis uint64, bip141 bool) InputHtmlData {

	displayTypeClassPrefix := fmt.Sprintf ("input-%d", txIndex)
	htmlData := InputHtmlData { InputIndex: txIndex, DisplayTypeClassPrefix: displayTypeClassPrefix, SpendType: input.GetSpendType (), Sequence: input.GetSequence (), Bip141: bip141 }

	htmlId := fmt.Sprintf ("input-script-%d", txIndex)

	previousOutput := input.GetPreviousOutput ()

	if input.IsCoinbase () {
		htmlData.IsCoinbase = true
		htmlData.ValueIn = template.HTML (getValueHtml (satoshis))
	} else {
		htmlData.BaseUrl = app.Settings.GetFullUrl () + "/web"
		htmlData.PreviousOutputTxId = input.GetPreviousOutputTxId ()
		htmlData.PreviousOutputIndex = input.GetPreviousOutputIndex ()
		htmlData.ValueIn = template.HTML (getValueHtml (previousOutput.GetValue ()))
	}
	htmlData.InputScript = getScriptHtmlData (input.GetInputScript (), htmlId, displayTypeClassPrefix)

	// previous output
	if len (previousOutput.GetAddress ()) > 0 {
		htmlData.PreviousOutputScript = getScriptHtmlData (previousOutput.GetOutputScript (), fmt.Sprintf ("previous-output-script-%d", txIndex), displayTypeClassPrefix)
		htmlData.PreviousOutputAddress = previousOutput.GetAddress ()
	}

	// redeem script and segwit
	redeemScript := input.GetRedeemScript ()
	htmlData.RedeemScript = getScriptHtmlData (redeemScript, fmt.Sprintf ("redeem-script-%d", txIndex), displayTypeClassPrefix)

	segwit := input.GetSegwit ()
	htmlData.Segwit = getSegwitHtmlData (segwit, txIndex, displayTypeClassPrefix)

	return htmlData
}

func getOutputHtmlData (output btc.Output, scriptHtmlId string, displayTypeClassPrefix string, outputIndex uint16) OutputHtmlData {

	if len (displayTypeClassPrefix) == 0 {
		displayTypeClassPrefix = fmt.Sprintf ("output-%d", outputIndex)
	}
	outputScriptHtml := getScriptHtmlData (output.GetOutputScript (), scriptHtmlId, displayTypeClassPrefix)
	return OutputHtmlData { OutputIndex: outputIndex, DisplayTypeClassPrefix: displayTypeClassPrefix, OutputType: output.GetOutputType (), Value: template.HTML (getValueHtml (output.GetValue ())), Address: output.GetAddress (), OutputScript: outputScriptHtml }
}

func shortenField (fieldText string, length uint, dotCount uint) string {

	fieldLength := uint (len (fieldText))
	if fieldLength <= length { return fieldText }

	if dotCount >= length { return fieldText }

	charsToInclude := length - dotCount
	if charsToInclude < 2 { return fieldText }

	charsInEachPart := charsToInclude / 2
	part1End := charsInEachPart + (charsToInclude % 2)
	part2Begin := fieldLength - charsInEachPart

	// build the final result
	shortenedField := fieldText [0 : part1End]
	for i := uint (0); i < dotCount; i++ { shortenedField += "." }
	shortenedField += fieldText [part2Begin :]

	return shortenedField
}

const FIELD_MAX_WIDTH = uint (89)
const FIELD_DOT_COUNT = uint (5)

func getSegwitHtmlData (segwit btc.Segwit, inputIndex uint16, displayTypeClassPrefix string) SegwitHtmlData {

	if segwit.IsNil () { return SegwitHtmlData { IsEmpty: true} }

	htmlId := fmt.Sprintf ("input-%d-segwit", inputIndex)

	var hexFieldsHtml [] FieldHtmlData
	var textFieldsHtml [] FieldHtmlData
	var typeFieldsHtml [] FieldHtmlData

	fields := segwit.GetFields ()
	fieldCount := len (fields)
	if fieldCount > 0 {

		hexFieldsHtml = make ([] FieldHtmlData, fieldCount)
		textFieldsHtml = make ([] FieldHtmlData, fieldCount)
		typeFieldsHtml = make ([] FieldHtmlData, fieldCount)

		for f, field := range fields {

			// hex strings
			entireHexField := field.AsHex ()
			hexFieldsHtml [f] = FieldHtmlData { DisplayText: template.HTML (shortenField (entireHexField, FIELD_MAX_WIDTH, FIELD_DOT_COUNT)), ShowCopyButton: uint (len (entireHexField)) > FIELD_MAX_WIDTH }
			if hexFieldsHtml [f].ShowCopyButton {
				hexFieldsHtml [f].CopyText = entireHexField
			}

			// text strings
			bytes, _ := hex.DecodeString (entireHexField)
			entireTextField := string (bytes)
			shortenedField := shortenField (entireTextField, FIELD_MAX_WIDTH, FIELD_DOT_COUNT)
			finalText := hexToText (shortenedField)
			textFieldsHtml [f] = FieldHtmlData { DisplayText: template.HTML (finalText), ShowCopyButton: uint (len (entireTextField)) > FIELD_MAX_WIDTH }
			if shortenedField != entireTextField {
				textFieldsHtml [f].ShowCopyButton = true
				textFieldsHtml [f].CopyText = entireTextField
			}

			// field types
			typeFieldsHtml [f] = FieldHtmlData { DisplayText: template.HTML (field.AsType ()), ShowCopyButton: false }
		}
	}

	copyImageUrl := app.Settings.GetFullUrl () + "/web/image/clipboard-copy.png"

	fieldSet := FieldSetHtmlData { HtmlId: htmlId, DisplayTypeClassPrefix: displayTypeClassPrefix, CharWidth: FIELD_MAX_WIDTH, HexFields: hexFieldsHtml, TextFields: textFieldsHtml, TypeFields: typeFieldsHtml, CopyImageUrl: copyImageUrl }
	htmlData := SegwitHtmlData { FieldSet: fieldSet, IsEmpty: fieldCount == 0 }

	witnessScript := segwit.GetWitnessScript ()
	htmlData.WitnessScript = getScriptHtmlData (witnessScript, htmlId + "-witness-script", displayTypeClassPrefix)

	tapScript, _ := segwit.GetTapScript ()
	htmlData.TapScript = getScriptHtmlData (tapScript, htmlId + "-tap-script", displayTypeClassPrefix)

	return htmlData
}

func getScriptHtmlData (script btc.Script, htmlId string, displayTypeClassPrefix string) ScriptHtmlData {

	if script.IsNil () { return ScriptHtmlData { IsNil: true } }

	scriptHtmlData := ScriptHtmlData { FieldSet: FieldSetHtmlData { HtmlId: htmlId, DisplayTypeClassPrefix: displayTypeClassPrefix, CharWidth: FIELD_MAX_WIDTH }, IsNil: false, IsOrdinal: script.IsOrdinal () }

	scriptFields := script.GetFields ()
	fieldCount := len (scriptFields)
	if script.HasParseError () { fieldCount++ }

	if len (scriptFields) == 0 {
		scriptHtmlData.FieldSet.HexFields = [] FieldHtmlData { FieldHtmlData { DisplayText: "Empty", ShowCopyButton: false } }
		scriptHtmlData.FieldSet.TextFields = [] FieldHtmlData { FieldHtmlData { DisplayText: "Empty", ShowCopyButton: false } }
		scriptHtmlData.FieldSet.TypeFields = [] FieldHtmlData { FieldHtmlData { DisplayText: "Empty", ShowCopyButton: false } }
		return scriptHtmlData
	}

	hexFieldsHtml := make ([] FieldHtmlData, fieldCount)
	textFieldsHtml := make ([] FieldHtmlData, fieldCount)
	typeFieldsHtml := make ([] FieldHtmlData, fieldCount)

	for f, field := range scriptFields {

		// hex strings
		entireHexField := field.AsHex ()
		shortenedHex := shortenField (entireHexField, FIELD_MAX_WIDTH, FIELD_DOT_COUNT)
		hexFieldsHtml [f] = FieldHtmlData { DisplayText: template.HTML (shortenedHex), ShowCopyButton: uint (len (entireHexField)) > FIELD_MAX_WIDTH }
		if shortenedHex != entireHexField {
			hexFieldsHtml [f].ShowCopyButton = true
			hexFieldsHtml [f].CopyText = entireHexField
		}

		// text strings
		bytes, err := hex.DecodeString (entireHexField)
		isOpcode := err != nil

		entireTextField := string (bytes)
		shortenedField := shortenField (entireTextField, FIELD_MAX_WIDTH, FIELD_DOT_COUNT)
		finalText := entireHexField
		if !isOpcode {
			finalText = hexToText (shortenedField)
		}
		textFieldsHtml [f] = FieldHtmlData { DisplayText: template.HTML (finalText), ShowCopyButton: !isOpcode && uint (len (entireTextField)) > FIELD_MAX_WIDTH }
		if shortenedField != entireTextField {
			textFieldsHtml [f].ShowCopyButton = true
			textFieldsHtml [f].CopyText = entireTextField
		}

		// field types
		typeFieldsHtml [f] = FieldHtmlData { DisplayText: template.HTML (field.AsType ()), ShowCopyButton: false }
	}

	if script.HasParseError () {
		parseErrorStr := template.HTML ("<<< PARSE ERROR >>>")
		hexFieldsHtml [fieldCount - 1] = FieldHtmlData { DisplayText: parseErrorStr, ShowCopyButton: false }
		textFieldsHtml [fieldCount - 1] = FieldHtmlData { DisplayText: parseErrorStr, ShowCopyButton: false }
		typeFieldsHtml [fieldCount - 1] = FieldHtmlData { DisplayText: parseErrorStr, ShowCopyButton: false }
	}

	copyImageUrl := app.Settings.GetFullUrl () + "/web/image/clipboard-copy.png"

	scriptHtmlData.FieldSet.HexFields = hexFieldsHtml
	scriptHtmlData.FieldSet.TextFields = textFieldsHtml
	scriptHtmlData.FieldSet.TypeFields = typeFieldsHtml
	scriptHtmlData.FieldSet.CopyImageUrl = copyImageUrl

	return scriptHtmlData
}

func hexToText (originalStr string) string {

	result := originalStr
	result = strings.ReplaceAll (result, "<", "&lt;")
	result = strings.ReplaceAll (result, ">", "&gt;")
	result = strings.ReplaceAll (result, " ", "&nbsp;")

	return result
}

func getValueHtml (satoshis uint64) string {
	satoshisStr := strconv.FormatUint (satoshis, 10)
	digitCount := len (satoshisStr)
	if digitCount > 8 {
		btcDigits := digitCount - 8
		satoshisStr = "<span style=\"font-weight:bold;\">" + satoshisStr [0 : btcDigits] + "</span>" + satoshisStr [btcDigits :]
	}
	return satoshisStr
}

func extractBodyFromHTML (html string) string {

	bodyBegin := strings.Index (html, "<body>")
	if bodyBegin > -1 { bodyBegin += 6 } else { fmt.Println ("Body not found in chart HTML.") }

	bodyEnd := strings.Index (html, "</body>")
	if bodyEnd == -1 { fmt.Println ("Body not found in chart HTML.") }

	body := html [: bodyEnd]
	return body [bodyBegin :]
}

/*
func getBlockCharts (nonCoinbaseInputCount uint16, outputCount uint16, spendTypes map [string] uint16, outputTypes map [string] uint16) map [string] string {

	const pieRadius = 90
	const verticalPadding = 10
	longestLabel := 0

	// gather data for the spend type and output type charts

	spendTypeNames := [] string { btc.OUTPUT_TYPE_P2PK, btc.OUTPUT_TYPE_MultiSig, btc.OUTPUT_TYPE_P2PKH, btc.OUTPUT_TYPE_P2SH, btc.SPEND_TYPE_P2SH_P2WPKH, btc.SPEND_TYPE_P2SH_P2WSH, btc.OUTPUT_TYPE_P2WPKH, btc.OUTPUT_TYPE_P2WSH, btc.SPEND_TYPE_P2TR_Key, btc.SPEND_TYPE_P2TR_Script, btc.SPEND_TYPE_NonStandard }

	var spendTypesHTML [] ElementTypeHTML
	for _, typeName := range spendTypeNames {
		if spendTypes [typeName] > 0 {
			if len (typeName) > longestLabel { longestLabel = len (typeName) }
			spendTypesHTML = append (spendTypesHTML, ElementTypeHTML { Label: typeName, Count: spendTypes [typeName], Percent: fmt.Sprintf ("%9.2f%%", float32 (spendTypes [typeName] * 100) / float32 (nonCoinbaseInputCount)) })
		}
	}

	outputTypeNames := [] string { btc.OUTPUT_TYPE_P2PK, btc.OUTPUT_TYPE_MultiSig, btc.OUTPUT_TYPE_P2PKH, btc.OUTPUT_TYPE_P2SH, btc.OUTPUT_TYPE_P2WPKH, btc.OUTPUT_TYPE_P2WSH, btc.OUTPUT_TYPE_TAPROOT, btc.OUTPUT_TYPE_OP_RETURN, btc.OUTPUT_TYPE_WitnessUnknown, btc.OUTPUT_TYPE_NonStandard }

	var outputTypesHTML [] ElementTypeHTML
	for _, typeName := range outputTypeNames {
		if outputTypes [typeName] > 0 {
			if len (typeName) > longestLabel { longestLabel = len (typeName) }
			outputTypesHTML = append (outputTypesHTML, ElementTypeHTML { Label: typeName, Count: outputTypes [typeName], Percent: fmt.Sprintf ("%9.2f%%", float32 (outputTypes [typeName] * 100) / float32 (outputCount)) })
		}
	}

	outputTypeCount := len (outputTypesHTML)
	legendHeight := outputTypeCount
	legendHeight *= 22

	boxDimension := ((pieRadius + verticalPadding) * 2) + legendHeight
	boxDimensionStr := strconv.Itoa (boxDimension)

	htmlData := make (map [string] string)

	if len (spendTypeNames) > 0 {

		sort.SliceStable (spendTypesHTML, func (i, j int) bool { return spendTypesHTML [i].Count > spendTypesHTML [j].Count })

		// spend type values
		var spendTypeValues [] opts.PieData
		for _, elementData := range spendTypesHTML {
			if spendTypes [elementData.Label] > 0 {
				fmtStr := fmt.Sprintf ("%%-%ds %%6d %%7s", longestLabel + 3)
				elementLabel := fmt.Sprintf (fmtStr, elementData.Label, elementData.Count, elementData.Percent)
				spendTypeValues = append (spendTypeValues, opts.PieData { Name: elementLabel, Value: elementData.Count })
			}
		}

		spendTypeCount := len (spendTypesHTML)
		if spendTypeCount > outputTypeCount {
			legendHeight = spendTypeCount
			legendHeight *= 22

			boxDimension = ((pieRadius + verticalPadding) * 2) + legendHeight
			boxDimensionStr = strconv.Itoa (boxDimension)
		}

		// create the spend type chart

		pie := charts.NewPie ()
		pie.AddSeries ("Spend Types", spendTypeValues)
		pie.SetSeriesOptions (charts.WithLabelOpts (opts.Label { Show: false }), charts.WithPieChartOpts (opts.PieChart { Center: [] int { boxDimension / 2, legendHeight + verticalPadding + pieRadius }, Radius: [] int { 0, pieRadius } }))

		pie.Legend.Orient = "vertical"
		pie.Legend.Top = "0"
		pie.Legend.Height = strconv.Itoa (legendHeight)
		pie.Legend.Width = boxDimensionStr
		pie.Legend.ItemWidth = 12
		pie.Legend.ItemHeight = 12
		pie.Legend.TextStyle = new (opts.TextStyle)
		pie.Legend.TextStyle.FontFamily = "monospace"
		pie.Legend.TextStyle.FontStyle = "bold"

		pie.Initialization.PageTitle = "Spend Types"
		pie.Initialization.Width = boxDimensionStr + "px"
		pie.Initialization.Height = boxDimensionStr + "px"

		var buff bytes.Buffer
		pie.Render (&buff)
		htmlData ["SpendTypeChart"] = extractBodyFromHTML (buff.String ())
	}

	sort.SliceStable (outputTypesHTML, func (i, j int) bool { return outputTypesHTML [i].Count > outputTypesHTML [j].Count })

	// output type values
	var outputTypeValues [] opts.PieData
	for _, elementData := range outputTypesHTML {
		if outputTypes [elementData.Label] > 0 {
			fmtStr := fmt.Sprintf ("%%-%ds %%6d %%7s", longestLabel + 3)
			elementLabel := fmt.Sprintf (fmtStr, elementData.Label, elementData.Count, elementData.Percent)
			outputTypeValues = append (outputTypeValues, opts.PieData { Name: elementLabel, Value: elementData.Count })
		}
	}

	// create the output type chart

	pie := charts.NewPie ()
	pie.AddSeries ("Output Types", outputTypeValues)
	pie.SetSeriesOptions (charts.WithLabelOpts (opts.Label { Show: false }), charts.WithPieChartOpts (opts.PieChart { Center: [] int { boxDimension / 2, legendHeight + verticalPadding + pieRadius }, Radius: [] int { 0, pieRadius } }))

	pie.Legend.Orient = "vertical"
	pie.Legend.Top = "0"
	pie.Legend.Height = strconv.Itoa (legendHeight)
	pie.Legend.Width = boxDimensionStr
	pie.Legend.ItemWidth = 12
	pie.Legend.ItemHeight = 12
	pie.Legend.TextStyle = new (opts.TextStyle)
	pie.Legend.TextStyle.FontFamily = "monospace"
	pie.Legend.TextStyle.FontStyle = "bold"

	pie.Initialization.PageTitle = "Output Types"
	pie.Initialization.Width = boxDimensionStr + "px"
	pie.Initialization.Height = boxDimensionStr + "px"

	var buff bytes.Buffer
	pie.Render (&buff)
	htmlData ["OutputTypeChart"] = extractBodyFromHTML (buff.String ())

	return htmlData
}
*/

// TODO: this needs to be replaced by a web-dir setting
func GetPath () string {
	return "web/"
}

