package web

import (
	"fmt"
	"net/http"
	"time"
	"sort"
	"strings"
	"strconv"
	"encoding/json"
	"encoding/hex"
	"bytes"
	"html/template"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"

	"btctx/app"
	"btctx/btc"
	"btctx/rest"
)

const WEB_REST_VERSION = 1

// html template structs

type ElementTypeHTML struct {
	Label string
	Count uint32
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
	InputIndex uint32
	DisplayTypeClassPrefix string
	IsCoinbase bool
	SpendType string
	ValueIn template.HTML
	BaseUrl string
	PreviousOutputTxId string
	PreviousOutputIndex uint32
	Sequence uint32
	InputScript ScriptHtmlData
	InputScriptAlternate ScriptHtmlData
	RedeemScript ScriptHtmlData
	WitnessScript ScriptHtmlData
	TapScript ScriptHtmlData
	Bip141 bool
	Segwit SegwitHtmlData
	IncludeAlternateInputScript bool
}

type OutputHtmlData struct {
	OutputIndex uint32
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

type previousOutputJsonOut struct {
	InputTxId string
	InputIndex uint32
	PrevOutValue uint64
	PrevOutType string
	PrevOutAddress string
	PrevOutScriptHtml string
}

type BlockChartData struct {
	NonCoinbaseInputCount uint32
	OutputCount uint32
	SpendTypes map [string] uint32
	OutputTypes map [string] uint32
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

	var params [] string
	paramCount := 0

	requestParts := strings.Split (modifiedPath, "/")
	if requestParts [0] != "web" {
		errorMessage := fmt.Sprintf ("Invalid request: %s", request.URL.Path)
		fmt.Println (errorMessage)
		fmt.Fprint (response, errorMessage)
		return
	}

	params = requestParts [1:]
	paramCount = len (params)

	html := ""

	// here, duck typing is used to determine what the user is requesting by examining the parameters received
	possibleQueryTypes := [] string { "block" } // default query type
	if paramCount >= 1 { possibleQueryTypes [0] = params [0] }

	if possibleQueryTypes [0] == "search" {
		if paramCount < 2 { fmt.Println ("No search parameter provided."); return }
		possibleQueryTypes = determineQueryTypes (params [1])
	}

	restApi := rest.RestApiV1 {}

	customJavascript := fmt.Sprintf ("var base_url_web = '%s/web';\n", app.Settings.GetFullUrl ())
	customJavascript += fmt.Sprintf ("var base_url_rest = '%s/rest/v%d';\n", app.Settings.GetFullUrl (), WEB_REST_VERSION)

	for _, queryType := range possibleQueryTypes {
		switch queryType {

			// returns html
			case "block":

				if request.Method != "GET" { fmt.Println (fmt.Sprintf ("%s must be sent as a GET request.", queryType)); break }

				blockRequestData := make (map [string] interface {})

				// get the block request data
				blockParam := ""
				if paramCount >= 2 {
					blockParam = params [1]
					paramLen := len (blockParam)
					if paramLen > 0 {
						height, err := strconv.Atoi (blockParam)
						if err == nil {
							blockRequestData ["height"] = uint32 (height)
						} else if paramLen == 64 {
							blockRequestData ["hash"] = blockParam
						} else {
							break // it isn't a block hash or height
						}
					}
				}

				options := make (map [string] interface {})
				options ["WScriptUsage"] = true
				blockRequestData ["options"] = options
				blockData := restApi.GetBlockData (blockRequestData)

				// spend types
				knownSpendTypes, knownSpendTypeCount := getKnownSpendTypesHtmlData (blockData ["InputCount"].(uint16) - 1, blockData ["KnownSpendTypeMap"].(map [string] uint16))
				knownSpendTypeJs := ""
				for _, knownSpendType := range knownSpendTypes {
					if len (knownSpendTypeJs) > 0 { knownSpendTypeJs += "," }
					knownSpendTypeJs += fmt.Sprintf ("'%s':%d", knownSpendType.Label, knownSpendType.Count)
				}
				customJavascript += fmt.Sprintf ("var known_spend_types = {%s};\n", knownSpendTypeJs)
				customJavascript += fmt.Sprintf ("var known_spend_type_count = %d;\n", knownSpendTypeCount)
				customJavascript += fmt.Sprintf ("var unknown_spend_type_count = %d;\n", blockData ["InputCount"].(uint16) - knownSpendTypeCount)

				// output types
				outputTypes := getOutputTypesHtmlData (blockData ["OutputCount"].(uint16), blockData ["OutputTypeMap"].(map [string] uint16))
				outputTypeJs := ""
				for _, outputType := range outputTypes {
					if len (outputTypeJs) > 0 { outputTypeJs += "," }
					outputTypeJs += fmt.Sprintf ("'%s':%d", outputType.Label, outputType.Count)
				}
				customJavascript += fmt.Sprintf ("var output_types = {%s};\n", outputTypeJs)
				customJavascript += fmt.Sprintf ("var output_count = %d;\n", blockData ["OutputCount"].(uint16))

				// unknown spend types
				pendingPreviousOutputs := blockData ["UnknownSpendTypeMap"].(map [string] [] uint32)
				pendingPrevOutBytes, err := json.Marshal (pendingPreviousOutputs)
				if err != nil { fmt.Println (err.Error ()) }

				pendingPrevOutJson := string (pendingPrevOutBytes)
				customJavascript += "var pending_block_spend_types = JSON.parse ('" + pendingPrevOutJson + "');"

				html = getBlockHtml (blockData, customJavascript)

			// returns html
			case "tx":

				if request.Method != "GET" { fmt.Println (fmt.Sprintf ("%s must be sent as a GET request.", queryType)); break }

				if paramCount < 2 { fmt.Println ("No id provided for tx. Request ignored."); return }
				if len (params [1]) != 64 { fmt.Println (fmt.Sprintf ("Parameter %s is not a valid tx id.", params [1])); return }

				txRequestData := make (map [string] interface {})
				txRequestData ["id"] = params [1]
				txRequestData ["options"] = map [string] interface {} { "PreviousOutputs": true }
				txData := restApi.GetTxData (txRequestData)
				if txData == nil {
					// TODO: need better error handling
					html = "Empty response from server."
					break
				}

				if txData ["PreviousOutputRequests"] != nil {
					pendingPreviousOutputsBytes, err := json.Marshal (txData ["PreviousOutputRequests"].([] rest.PreviousOutputRequest))
					if err != nil { fmt.Println (err.Error ()) }

					pendingPreviousOutputsJson := string (pendingPreviousOutputsBytes)
					customJavascript += "var pending_tx_previous_outputs = JSON.parse ('" + pendingPreviousOutputsJson + "');"
				}

				html = getTxHtml (txData, customJavascript)

			case "address":
				// requires an electrum server
				break


			// returns json
			case "prevout":

				if request.Method != "POST" { fmt.Println (fmt.Sprintf ("%s must be sent as a POST request.", queryType)); return }

				// unpack the json
				var previousOutputJsonIn rest.PreviousOutputRequest
				err := json.NewDecoder (request.Body).Decode (&previousOutputJsonIn)
				if err != nil { fmt.Println (err.Error()) }

				txId := previousOutputJsonIn.PrevOutTxId
				outputIndex := previousOutputJsonIn.PrevOutIndex
				inputIndex := previousOutputJsonIn.InputIndex
				inputTxId := previousOutputJsonIn.InputTxId

				previousOutput := restApi.GetPreviousOutputResponseData (txId, uint32 (outputIndex))
				idPrefix := fmt.Sprintf ("previous-output-%d", inputIndex)
				classPrefix := fmt.Sprintf ("input-%d", inputIndex)
				previousOutputScriptHtml := getPreviousOutputScriptHtml (previousOutput.OutputScript, idPrefix, classPrefix)

				// return the json response
				previousOutputResponse := previousOutputJsonOut { InputTxId: inputTxId,
																	InputIndex: uint32 (inputIndex),
																	PrevOutValue: previousOutput.Value,
																	PrevOutType: previousOutput.OutputType,
																	PrevOutAddress: previousOutput.Address,
																	PrevOutScriptHtml: previousOutputScriptHtml }

				jsonBytes, err := json.Marshal (previousOutputResponse)
				if err != nil { fmt.Println (err) }

				fmt.Fprint (response, string (jsonBytes))
				return

			// returns json
			case "legacy_spend_types":

				if request.Method != "POST" { fmt.Println (fmt.Sprintf ("%s must be sent as a POST request.", queryType)); return }

				// unpack the json
				var requestedPreviousOutputs map [string] [] uint32
				err := json.NewDecoder (request.Body).Decode (&requestedPreviousOutputs)
				if err != nil { fmt.Println (err.Error()) }

				prevOutMap := restApi.GetPreviousOutputTypes (requestedPreviousOutputs)

				// if this is not a legacy type, it must be a non-standard input
				for outpoint, outputType := range prevOutMap {
					if outputType != btc.OUTPUT_TYPE_P2PK && outputType != btc.OUTPUT_TYPE_MultiSig && outputType != btc.OUTPUT_TYPE_P2PKH && outputType != btc.OUTPUT_TYPE_P2SH {
						prevOutMap [outpoint] = btc.SPEND_TYPE_NonStandard
					}
				}

				prevOutsBytes, err := json.Marshal (prevOutMap)
				if err != nil { fmt.Println (err.Error ()) }

				fmt.Fprint (response, string (prevOutsBytes))
				return

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

func getKnownSpendTypesHtmlData (nonCoinbaseInputCount uint16, knownSpendTypeMap map [string] uint16) ([] ElementTypeHTML, uint16) {

	var knownSpendTypes [] ElementTypeHTML
	knownSpendTypeCount := uint16 (1) // starting at 1 because the coinbase input always has a known spend type
	for spendType, num := range knownSpendTypeMap {
		knownSpendTypeCount += num
		knownSpendTypes = append (knownSpendTypes, ElementTypeHTML { Label: spendType, Count: uint32 (num), Percent: fmt.Sprintf ("%9.2f%%", float32 (num * 100) / float32 (nonCoinbaseInputCount)) })
	}

	return knownSpendTypes, knownSpendTypeCount
}
func getOutputTypesHtmlData (outputCount uint16, outputTypeMap map [string] uint16) [] ElementTypeHTML {

	var outputTypes [] ElementTypeHTML
	for outputType, num := range outputTypeMap {
		outputTypes = append (outputTypes, ElementTypeHTML { Label: outputType, Count: uint32 (num), Percent: fmt.Sprintf ("%9.2f%%", float32 (num * 100) / float32 (outputCount)) })
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

	nodeClient := btc.GetNodeClient ()
	layoutData ["NodeVersion"] = nodeClient.GetVersionString ()
	layoutData ["NodeUrl"] = app.Settings.GetNodeFullUrl ()

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

func getBlockHtml (blockData map [string] interface {}, customJavascript string) string {

	// get the data
	blockHtmlData := make (map [string] interface {})

	// get block data
	blockHtmlData ["Height"] = blockData ["Height"].(uint32)
	blockHtmlData ["Time"] = time.Unix (blockData ["Timestamp"].(int64), 0).UTC ()
	blockHtmlData ["Hash"] = blockData ["Hash"].(string)

	if blockData ["PreviousHash"] != nil { blockHtmlData ["PreviousHash"] = blockData ["PreviousHash"].(string) }
	if blockData ["NextHash"] != nil { blockHtmlData ["NextHash"] = blockData ["NextHash"].(string) }
	blockHtmlData ["TxCount"] = uint16 (len (blockData ["Txs"].([] rest.BlockTxData)))

	// get the numbers of inputs and outputs and their types, and the tx detail
	if blockData ["WitnessScriptCount"] != nil && blockData ["WitnessScriptMultisigCount"] != nil {
		witnessScriptCount := blockData ["WitnessScriptCount"].(uint16)
		if witnessScriptCount > 0 {
			witnessScriptMultisigCount := blockData ["WitnessScriptMultisigCount"].(uint16)
			wsMessage:= fmt.Sprintf ("%6.2f%% MultiSig (%d/%d)", float32 (witnessScriptMultisigCount) * 100 / float32 (witnessScriptCount), witnessScriptMultisigCount, witnessScriptCount)
			blockHtmlData ["WitnessScriptMultiSigMessage"] = template.HTML (strings.Replace (wsMessage, " ", "&nbsp;", -1))
		}
	}

	if blockData ["TapScriptCount"] != nil && blockData ["TapScriptOrdinalCount"] != nil {
		tapScriptCount := blockData ["TapScriptCount"].(uint16)
		if tapScriptCount > 0 {
			tapScriptOrdinalCount := blockData ["TapScriptOrdinalCount"].(uint16)
			tsMessage:= fmt.Sprintf ("%6.2f%% Ordinals (%d/%d)", float32 (tapScriptOrdinalCount) * 100 / float32 (tapScriptCount), tapScriptOrdinalCount, tapScriptCount)
			blockHtmlData ["TapScriptOrdinalsMessage"] = template.HTML (strings.Replace (tsMessage, " ", "&nbsp;", -1))
		}
	}

	blockHtmlData ["Bip141Percent"] = fmt.Sprintf ("%9.2f%%", float32 (blockData ["Bip141Count"].(uint16)) * 100 / float32 (blockHtmlData ["TxCount"].(uint16)))

	blockHtmlData ["BaseUrl"] = app.Settings.GetFullUrl () + "/web"
	blockHtmlData ["TxData"] = blockData ["Txs"].([] rest.BlockTxData)

	blockHtmlData ["InputCount"] = blockData ["InputCount"].(uint16)
	blockHtmlData ["OutputCount"] = blockData ["OutputCount"].(uint16)

	// create the html page
	explorerPageHtmlData := getExplorerPageHtmlData (blockData ["Hash"].(string), blockHtmlData)
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

func getTxHtml (txData map [string] interface {}, customJavascript string) string {

	txPageHtmlData := make (map [string] interface {})

	// transaction data
	txPageHtmlData ["BaseUrl"] = app.Settings.GetFullUrl () + "/web"
	txPageHtmlData ["BlockHeight"] = txData ["BlockHeight"].(uint32)
	txPageHtmlData ["BlockTime"] = time.Unix (txData ["BlockTime"].(int64), 0).UTC ()

	txPageHtmlData ["BlockHash"] = txData ["BlockHash"].(string)
	txPageHtmlData ["IsCoinbase"] = txData ["IsCoinbase"].(bool)
	txPageHtmlData ["SupportsBip141"] = txData ["SupportsBip141"].(bool)
	txPageHtmlData ["LockTime"] = txData ["LockTime"].(uint32)

	// outputs
	totalOut := uint64 (0)
	outputs := txData ["Outputs"].([] rest.OutputData)
	outputCount := len (outputs)

	outputCountLabel := fmt.Sprintf ("%d Output", outputCount)
	if outputCount > 1 { outputCountLabel += "s" }
	txPageHtmlData ["OutputCountLabel"] = outputCountLabel

	outputHtmlData := make ([] OutputHtmlData, outputCount)
	for o, output := range outputs {
		totalOut += output.Value
		scriptHtmlId := fmt.Sprintf ("output-script-%d", o)
		outputHtmlData [o] = getOutputHtmlData (outputs [o], scriptHtmlId, "", uint32 (o))
	}
	txPageHtmlData ["OutputData"] = outputHtmlData

	// totals for the transaction
	txPageHtmlData ["ValueOut"] = totalOut
	txPageHtmlData ["ValueIn"] = 0; if txData ["IsCoinbase"].(bool) { txPageHtmlData ["ValueIn"] = totalOut }
	txPageHtmlData ["Fee"] = 0

	// inputs
	inputs := txData ["Inputs"].([] map [string] interface {})
	inputCount := len (inputs)

	inputCountLabel := fmt.Sprintf ("%d Input", inputCount)
	if inputCount > 1 { inputCountLabel += "s" }
	txPageHtmlData ["InputCountLabel"] = inputCountLabel

	inputHtmlData := make ([] InputHtmlData, inputCount)
	for i := uint32 (0); i < uint32 (inputCount); i++ {
		valueIn := uint64 (0); if txData ["IsCoinbase"].(bool) && i == 0 { valueIn = totalOut }
		inputHtmlData [i] = getInputHtmlData (inputs [i], valueIn, txData ["SupportsBip141"].(bool))
	}
	txPageHtmlData ["InputData"] = inputHtmlData

	// add the tx html data to the page and layout html data
	explorerPageHtmlData := getExplorerPageHtmlData (txData ["Id"].(string), txPageHtmlData)
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

func getInputHtmlData (input map [string] interface {}, satoshis uint64, bip141 bool) InputHtmlData {

	inputIndex := input ["InputIndex"].(uint32)
	spendType := input ["SpendType"].(string)

	displayTypeClassPrefix := fmt.Sprintf ("input-%d", inputIndex)
	htmlData := InputHtmlData { InputIndex: inputIndex, DisplayTypeClassPrefix: displayTypeClassPrefix, SpendType: spendType, Sequence: input ["Sequence"].(uint32), Bip141: bip141 }
	htmlId := fmt.Sprintf ("input-script-%d", inputIndex)

	if input ["Coinbase"].(bool) {
		htmlData.IsCoinbase = true
		htmlData.ValueIn = template.HTML (getValueHtml (satoshis))
	} else {
		htmlData.BaseUrl = app.Settings.GetFullUrl () + "/web"
		htmlData.PreviousOutputTxId = input ["PreviousOutputTxId"].(string)
		htmlData.PreviousOutputIndex = input ["PreviousOutputIndex"].(uint32)
	}
	htmlData.InputScript = getScriptHtmlData (input ["InputScript"].(map [string] interface {}), htmlId, displayTypeClassPrefix)

	// redeem script and segwit
	redeemScript := map [string] interface {} (nil)
	if input ["RedeemScript"] != nil { redeemScript = input ["RedeemScript"].(map [string] interface {}) }
	htmlData.RedeemScript = getScriptHtmlData (redeemScript, fmt.Sprintf ("redeem-script-%d", inputIndex), displayTypeClassPrefix)

	segwit := map [string] interface {} (nil)
	if input ["Segwit"] != nil { segwit = input ["Segwit"].(map [string] interface {}) }
	htmlData.Segwit = getSegwitHtmlData (segwit, inputIndex, displayTypeClassPrefix)

	return htmlData
}

func getOutputHtmlData (output rest.OutputData, scriptHtmlId string, displayTypeClassPrefix string, outputIndex uint32) OutputHtmlData {

	if len (displayTypeClassPrefix) == 0 {
		displayTypeClassPrefix = fmt.Sprintf ("output-%d", outputIndex)
	}
	outputScriptHtml := getScriptHtmlData (output.OutputScript, scriptHtmlId, displayTypeClassPrefix)
	return OutputHtmlData { OutputIndex: outputIndex, DisplayTypeClassPrefix: displayTypeClassPrefix, OutputType: output.OutputType, Value: template.HTML (getValueHtml (output.Value)), Address: output.Address, OutputScript: outputScriptHtml }
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

func getSegwitHtmlData (segwit map [string] interface {}, inputIndex uint32, displayTypeClassPrefix string) SegwitHtmlData {

	if segwit == nil { return SegwitHtmlData { IsEmpty: true} }

	htmlId := fmt.Sprintf ("input-%d-segwit", inputIndex)

	var hexFieldsHtml [] FieldHtmlData
	var textFieldsHtml [] FieldHtmlData
	var typeFieldsHtml [] FieldHtmlData

	fields := segwit ["Fields"].([] rest.FieldData)
	fieldCount := len (fields)
	if fieldCount > 0 {

		hexFieldsHtml = make ([] FieldHtmlData, fieldCount)
		textFieldsHtml = make ([] FieldHtmlData, fieldCount)
		typeFieldsHtml = make ([] FieldHtmlData, fieldCount)

		for f, field := range fields {

			// hex strings
			entireHexField := field.Hex
			hexFieldsHtml [f] = FieldHtmlData { DisplayText: template.HTML (shortenField (field.Hex, FIELD_MAX_WIDTH, FIELD_DOT_COUNT)), ShowCopyButton: uint (len (field.Hex)) > FIELD_MAX_WIDTH }
			if hexFieldsHtml [f].ShowCopyButton {
				hexFieldsHtml [f].CopyText = entireHexField
			}

			// text strings
			bytes, _ := hex.DecodeString (field.Hex)
			entireTextField := string (bytes)
			shortenedField := shortenField (entireTextField, FIELD_MAX_WIDTH, FIELD_DOT_COUNT)
			finalText := hexToText (shortenedField)
			textFieldsHtml [f] = FieldHtmlData { DisplayText: template.HTML (finalText), ShowCopyButton: uint (len (entireTextField)) > FIELD_MAX_WIDTH }
			if shortenedField != entireTextField {
				textFieldsHtml [f].ShowCopyButton = true
				textFieldsHtml [f].CopyText = entireTextField
			}

			// field types
			typeFieldsHtml [f] = FieldHtmlData { DisplayText: template.HTML (field.Type), ShowCopyButton: false }
		}
	}

	copyImageUrl := app.Settings.GetFullUrl () + "/web/image/clipboard-copy.png"

	fieldSet := FieldSetHtmlData { HtmlId: htmlId, DisplayTypeClassPrefix: displayTypeClassPrefix, CharWidth: FIELD_MAX_WIDTH, HexFields: hexFieldsHtml, TextFields: textFieldsHtml, TypeFields: typeFieldsHtml, CopyImageUrl: copyImageUrl }
	htmlData := SegwitHtmlData { FieldSet: fieldSet, IsEmpty: fieldCount == 0 }

	witnessScript := map [string] interface {} (nil)
	if segwit ["WitnessScript"] != nil { witnessScript = segwit ["WitnessScript"].(map [string] interface {}) }
	htmlData.WitnessScript = getScriptHtmlData (witnessScript, htmlId + "-witness-script", displayTypeClassPrefix)

	tapScript := map [string] interface {} (nil)
	if segwit ["TapScript"] != nil { tapScript = segwit ["TapScript"].(map [string] interface {}) }
	htmlData.TapScript = getScriptHtmlData (tapScript, htmlId + "-tap-script", displayTypeClassPrefix)

	return htmlData
}

func getScriptHtmlData (script map [string] interface {}, htmlId string, displayTypeClassPrefix string) ScriptHtmlData {

	if script == nil { return ScriptHtmlData { IsNil: true } }

	isOrdinal := script ["Ordinal"] != nil && script ["Ordinal"].(bool)
	scriptHtmlData := ScriptHtmlData { FieldSet: FieldSetHtmlData { HtmlId: htmlId, DisplayTypeClassPrefix: displayTypeClassPrefix, CharWidth: FIELD_MAX_WIDTH }, IsNil: false, IsOrdinal: isOrdinal }

	scriptFields := script ["Fields"].([] rest.FieldData)

	if len (scriptFields) == 0 {
		scriptHtmlData.FieldSet.HexFields = [] FieldHtmlData { FieldHtmlData { DisplayText: "Empty", ShowCopyButton: false } }
		scriptHtmlData.FieldSet.TextFields = [] FieldHtmlData { FieldHtmlData { DisplayText: "Empty", ShowCopyButton: false } }
		scriptHtmlData.FieldSet.TypeFields = [] FieldHtmlData { FieldHtmlData { DisplayText: "Empty", ShowCopyButton: false } }
		return scriptHtmlData
	}

	fields := script ["Fields"].([] rest.FieldData)
	fieldCount := len (fields)
	if script ["ParseError"] != nil && script ["ParseError"].(bool) { fieldCount++ }

	hexFieldsHtml := make ([] FieldHtmlData, fieldCount)
	textFieldsHtml := make ([] FieldHtmlData, fieldCount)
	typeFieldsHtml := make ([] FieldHtmlData, fieldCount)

	for f, field := range fields {

		// hex strings
		entireHexField := field.Hex
		shortenedHex := shortenField (field.Hex, FIELD_MAX_WIDTH, FIELD_DOT_COUNT)
		hexFieldsHtml [f] = FieldHtmlData { DisplayText: template.HTML (shortenedHex), ShowCopyButton: uint (len (entireHexField)) > FIELD_MAX_WIDTH }
		if shortenedHex != entireHexField {
			hexFieldsHtml [f].ShowCopyButton = true
			hexFieldsHtml [f].CopyText = entireHexField
		}

		// text strings
		bytes, err := hex.DecodeString (field.Hex)
		isOpcode := err != nil

		entireTextField := string (bytes)
		shortenedField := shortenField (entireTextField, FIELD_MAX_WIDTH, FIELD_DOT_COUNT)
		finalText := field.Hex
		if !isOpcode {
			finalText = hexToText (shortenedField)
		}
		textFieldsHtml [f] = FieldHtmlData { DisplayText: template.HTML (finalText), ShowCopyButton: !isOpcode && uint (len (entireTextField)) > FIELD_MAX_WIDTH }
		if shortenedField != entireTextField {
			textFieldsHtml [f].ShowCopyButton = true
			textFieldsHtml [f].CopyText = entireTextField
		}

		// field types
		typeFieldsHtml [f] = FieldHtmlData { DisplayText: template.HTML (field.Type), ShowCopyButton: false }
	}

	if script ["ParseError"] != nil && script ["ParseError"].(bool) {
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

func getPreviousOutputScriptHtml (script map [string] interface {}, htmlId string, displayTypeClassPrefix string) string {

	// get the data
	previousOutputHtmlData := getScriptHtmlData (script, htmlId, displayTypeClassPrefix)
	
	// parse the file
	layoutHtmlFiles := [] string {
		GetPath () + "html/field-set.html" }
	templ := template.Must (template.ParseFiles (layoutHtmlFiles...))

	// execute the template
	var buff bytes.Buffer
	if err := templ.ExecuteTemplate (&buff, "FieldSet", previousOutputHtmlData.FieldSet); err != nil { panic (err) }

	// return the html
	return buff.String ()
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

func getBlockCharts (nonCoinbaseInputCount uint32, outputCount uint32, spendTypes map [string] uint32, outputTypes map [string] uint32) map [string] string {

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

// TODO: this needs to be replaced by a web-dir setting
func GetPath () string {
	return "web/"
}

