package web

import (
	"fmt"
	"net/http"
	"time"
	"strings"
	"strconv"
	"encoding/json"
	"encoding/hex"
	"bytes"
	"html/template"

	"btctx/app"
	"btctx/btc"
)

// html template structs

type ScriptFieldHtmlData struct {
	DisplayField string
	ShowCopyButton bool
	CopyText string
}

type FieldHtmlData struct {
	DisplayText string
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


func WebV1Handler (response http.ResponseWriter, request *http.Request) {

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

	nodeClient := btc.GetNodeClient ()
	html := ""

	// here, duck typing is used to determine what the user is requesting by examining the parameters received
	possibleQueryTypes := [] string { "block" } // default query type
	if paramCount >= 1 { possibleQueryTypes [0] = params [0] }

	if possibleQueryTypes [0] == "search" {
		if paramCount < 2 { fmt.Println ("No search parameter provided."); return }
		possibleQueryTypes = determineQueryTypes (params [1])
	}

	customJavascript := fmt.Sprintf ("var base_url_web = '%s';\n", app.Settings.GetFullUrl ())

	for _, queryType := range possibleQueryTypes {
		switch queryType {
			case "block":

				if request.Method != "GET" { fmt.Println (fmt.Sprintf ("%s must be sent as a GET request.", queryType)); break }

				blockParam := ""
				if paramCount >= 2 { blockParam = params [1] }
				paramLen := len (blockParam)

				// determine the block hash
				// the params could be a hash, height, height range or it could be left empty (current block)
				blockHash := ""
				if paramLen == 0 {
					// it is the default block, the current block
					blockHash = nodeClient.GetCurrentBlockHash ()
				} else if paramLen == 64 {
					// it could be a block hash
					blockHash = blockParam
				} else {
					// it could be a block height
					blockHeight, err := strconv.Atoi (blockParam)
					if err == nil {
						blockHash = nodeClient.GetBlockHash (blockHeight)
					}
				}

				if len (blockHash) > 0 {
					block := nodeClient.GetBlock (blockHash, true)
					if !block.IsNil () {

						inputCount, outputCount := block.GetInputOutputCounts ()

						// known spend types
						knownSpendTypes, knownSpendTypeCount := block.GetKnownSpendTypes ()
						knownSpendTypeJs := ""
						for _, knownSpendType := range knownSpendTypes {
							if len (knownSpendTypeJs) > 0 { knownSpendTypeJs += "," }
							knownSpendTypeJs += fmt.Sprintf ("'%s':%d", knownSpendType.Label, knownSpendType.Count)
						}
						customJavascript += fmt.Sprintf ("var known_spend_types = {%s};\n", knownSpendTypeJs)
						customJavascript += fmt.Sprintf ("var known_spend_type_count = %d;\n", knownSpendTypeCount)
						customJavascript += fmt.Sprintf ("var unknown_spend_type_count = %d;\n", inputCount - knownSpendTypeCount)

						// output types
						outputTypes := block.GetOutputTypes ()
						outputTypeJs := ""
						for _, outputType := range outputTypes {
							if len (outputTypeJs) > 0 { outputTypeJs += "," }
							outputTypeJs += fmt.Sprintf ("'%s':%d", outputType.Label, outputType.Count)
						}
						customJavascript += fmt.Sprintf ("var output_types = {%s};\n", outputTypeJs)
						customJavascript += fmt.Sprintf ("var output_count = %d;\n", outputCount)

						// unknown spend types
						pendingPreviousOutputs := block.GetPendingPreviousOutputs ()
						pendingPrevOutBytes, err := json.Marshal (pendingPreviousOutputs)
						if err != nil { fmt.Println (err.Error ()) }

						pendingPrevOutJson := string (pendingPrevOutBytes)
						customJavascript += "var pending_block_spend_types = JSON.parse ('" + pendingPrevOutJson + "');"

						html = getBlockHtml (block, customJavascript)
					}
				}

			case "tx":

				if request.Method != "GET" { fmt.Println (fmt.Sprintf ("%s must be sent as a GET request.", queryType)); break }

				if paramCount < 2 { fmt.Println ("Wrong number of parameters for tx. Request ignored."); return }

				txId := params [1]
				txIdBytes, err := hex.DecodeString (txId)
				if err != nil { panic (err.Error ()) }

				tx := nodeClient.GetTx (hex.EncodeToString (txIdBytes))
				if !tx.IsNil () {
					pendingPreviousOutputsBytes, err := json.Marshal (tx.GetPendingPreviousOutputs ())
					if err != nil { fmt.Println (err.Error ()) }

					pendingPreviousOutputsJson := string (pendingPreviousOutputsBytes)
					customJavascript += "var pending_tx_previous_outputs = JSON.parse ('" + pendingPreviousOutputsJson + "');"
					html = getTxHtml (tx, customJavascript)
				}

			case "address":
				break


			case "prevout":

				// this one works more like a rest api or ajax call, only json is returned, which contains some html
				if request.Method != "POST" { fmt.Println (fmt.Sprintf ("%s must be sent as a POST request.", queryType)); return }

				// unpack the json
				var previousOutputJsonIn btc.PendingPreviousOutput
				err := json.NewDecoder (request.Body).Decode (&previousOutputJsonIn)
				if err != nil { fmt.Println (err.Error()) }

				txId := previousOutputJsonIn.PrevOutTxId
				outputIndex := previousOutputJsonIn.PrevOutIndex
				inputIndex := previousOutputJsonIn.InputIndex
				inputTxId := previousOutputJsonIn.InputTxId

				previousOutput := nodeClient.GetPreviousOutput (txId, uint32 (outputIndex))
				idPrefix := fmt.Sprintf ("previous-output-%d", inputIndex)
				classPrefix := fmt.Sprintf ("input-%d", inputIndex)
				previousOutputScriptHtml := getPreviousOutputScriptHtml (previousOutput.GetOutputScript (), idPrefix, classPrefix)

				// return the json response
				satoshis := previousOutput.GetValue ()
				previousOutputResponse := previousOutputJsonOut { InputTxId: inputTxId,
																	InputIndex: uint32 (inputIndex),
																	PrevOutValue: satoshis,
																	PrevOutType: previousOutput.GetOutputType (),
																	PrevOutAddress: previousOutput.GetAddress (),
																	PrevOutScriptHtml: previousOutputScriptHtml }

				jsonBytes, err := json.Marshal (previousOutputResponse)
				if err != nil { fmt.Println (err) }

				fmt.Fprint (response, string (jsonBytes))
				return

			case "get_block_charts":

				if request.Method != "POST" { fmt.Println (fmt.Sprintf ("%s must be sent as a POST request.", queryType)); return }

				// unpack the json
				var blockChartData BlockChartData
				err := json.NewDecoder (request.Body).Decode (&blockChartData)
				if err != nil { fmt.Println (err.Error()) }

				blockCharts := btc.GetBlockCharts (blockChartData.NonCoinbaseInputCount, blockChartData.OutputCount, blockChartData.SpendTypes, blockChartData.OutputTypes)

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

func getBlockHtml (block btc.Block, customJavascript string) string {

	// get the data
	blockHtmlData := block.GetHtmlData ()
	explorerPageHtmlData := getExplorerPageHtmlData (block.GetHash (), blockHtmlData)
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

func getTxHtml (tx btc.Tx, customJavascript string) string {

	txPageHtmlData := make (map [string] interface {})

	// transaction data
	txPageHtmlData ["BaseUrl"] = app.Settings.GetFullUrl ()
	txPageHtmlData ["BlockHeight"] = tx.GetBlockHeight ()
	txPageHtmlData ["BlockTime"] = time.Unix (tx.GetBlockTime (), 0).UTC ()
	txPageHtmlData ["BlockHash"] = tx.GetBlockHash ()
	txPageHtmlData ["IsCoinbase"] = tx.IsCoinbase ()
	txPageHtmlData ["SupportsBip141"] = tx.SupportsBip141 ()
	txPageHtmlData ["LockTime"] = tx.GetLockTime ()

	// outputs
	totalOut := uint64 (0)
	outputCount := tx.GetOutputCount ()

	outputCountLabel := strconv.Itoa (outputCount) + " Output"
	if outputCount > 1 { outputCountLabel += "s" }
	txPageHtmlData ["OutputCountLabel"] = outputCountLabel

	outputs := tx.GetOutputs ()
	outputHtmlData := make ([] OutputHtmlData, outputCount)
	for o := uint32 (0); o < uint32 (outputCount); o++ {
		totalOut += outputs [o].GetValue ()
		scriptHtmlId := "output-script-" + strconv.FormatUint (uint64 (o), 10)
		outputHtmlData [o] = getOutputHtmlData (outputs [o], scriptHtmlId, "", o)
	}
	txPageHtmlData ["OutputData"] = outputHtmlData

	// totals for the transaction
	txPageHtmlData ["ValueOut"] = totalOut
	txPageHtmlData ["ValueIn"] = 0; if tx.IsCoinbase () { txPageHtmlData ["ValueIn"] = totalOut }
	txPageHtmlData ["Fee"] = 0

	// inputs
	inputCount := tx.GetInputCount ()

	inputCountLabel := strconv.Itoa (inputCount) + " Input"
	if inputCount > 1 { inputCountLabel += "s" }
	txPageHtmlData ["InputCountLabel"] = inputCountLabel

	inputs := tx.GetInputs ()
	inputHtmlData := make ([] InputHtmlData, inputCount)
	for i := uint32 (0); i < uint32 (inputCount); i++ {
		valueIn := uint64 (0); if tx.IsCoinbase () && i == 0 { valueIn = totalOut }
		inputHtmlData [i] = getInputHtmlData (inputs [i], i, valueIn, tx.SupportsBip141 ())
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

func getInputHtmlData (input btc.Input, inputIndex uint32, satoshis uint64, bip141 bool) InputHtmlData {

	displayTypeClassPrefix := fmt.Sprintf ("input-%d", inputIndex)
	htmlData := InputHtmlData { InputIndex: inputIndex, DisplayTypeClassPrefix: displayTypeClassPrefix, SpendType: input.GetSpendType (), Sequence: input.GetSequence (), Bip141: bip141 }
	htmlId := "input-script-" + strconv.FormatUint (uint64 (inputIndex), 10)

	if input.IsCoinbase () {
		htmlData.IsCoinbase = true
		htmlData.ValueIn = template.HTML (getValueHtml (satoshis))
	} else {
		htmlData.BaseUrl = app.Settings.GetFullUrl ()
		htmlData.PreviousOutputTxId = input.GetPreviousOutputTxId ()
		htmlData.PreviousOutputIndex = input.GetPreviousOutputIndex ()
	}
	htmlData.InputScript = getScriptHtmlData (input.GetInputScript (), htmlId, displayTypeClassPrefix)

	// if the spend type is empty, it redeems a legacy output type
	// therefore, we must include an alternate input script to account for the possibility of a P2SH output
	if len (input.GetSpendType ()) == 0 {
		inputScript := input.GetInputScript ()
		redeemScript := inputScript.GetSerializedScript ()
		if !redeemScript.HasParseError () {
			input.SetRedeemScript (redeemScript)
			inputScriptAlternate := btc.NewScript (inputScript.AsBytes ())
			if !inputScriptAlternate.IsEmpty () {
				inputScriptAlternate.SetFieldType (inputScriptAlternate.GetParsedFieldCount () - 1, "<<< SERIALIZED REDEEM SCRIPT >>>")

				// check for a zero-length redeem script
				alternateScriptFields := inputScriptAlternate.GetFields ()
				serializedScriptIndex := len (alternateScriptFields) - 1
				serializedScriptBytes := alternateScriptFields [serializedScriptIndex].AsBytes ()
				if len (serializedScriptBytes) == 1 && serializedScriptBytes [0] == 0x00 {
					alternateScriptFields [serializedScriptIndex].SetIsOpcode (false)
					alternateScriptFields [serializedScriptIndex].SetBytes ([] byte {})
					input.SetRedeemScript (btc.NewScript ([] byte {}))
				}

				htmlData.InputScriptAlternate = getScriptHtmlData (inputScriptAlternate, htmlId, displayTypeClassPrefix)
				htmlData.IncludeAlternateInputScript = true
			}
		}
	}

	// redeem script and segwit
	htmlData.RedeemScript = getScriptHtmlData (input.GetRedeemScript (), "redeem-script-" + strconv.FormatUint (uint64 (inputIndex), 10), displayTypeClassPrefix)
	htmlData.Segwit = getSegwitHtmlData (input.GetSegwit (), inputIndex, displayTypeClassPrefix, input.GetSpendType () == btc.SPEND_TYPE_P2TR_Key || input.GetSpendType () == btc.SPEND_TYPE_P2TR_Script)

	return htmlData
}

func getOutputHtmlData (output btc.Output, scriptHtmlId string, displayTypeClassPrefix string, outputIndex uint32) OutputHtmlData {

	if len (displayTypeClassPrefix) == 0 {
		displayTypeClassPrefix = fmt.Sprintf ("output-%d", outputIndex)
	}
	outputScriptHtml := getScriptHtmlData (output.GetOutputScript (), scriptHtmlId, displayTypeClassPrefix)
	return OutputHtmlData { OutputIndex: outputIndex, DisplayTypeClassPrefix: displayTypeClassPrefix, OutputType: output.GetOutputType (), Value: template.HTML (getValueHtml (output.GetValue ())), Address: output.GetAddress (), OutputScript: outputScriptHtml }
}

func getSegwitHtmlData (segwit btc.Segwit, inputIndex uint32, displayTypeClassPrefix string, usingSchnorrSignatures bool) SegwitHtmlData {

	if segwit.IsNil () {
		return SegwitHtmlData { IsEmpty: true}
	}

	htmlId := fmt.Sprintf ("input-%d-segwit", inputIndex)
	const maxCharWidth = uint (89)

	var hexFieldsHtml [] FieldHtmlData
	var textFieldsHtml [] FieldHtmlData
	var typeFieldsHtml [] FieldHtmlData

	if !segwit.IsEmpty () {

		fields := segwit.GetFields ()
		fieldCount := len (fields);

		// segwit does not know what some of its types are, so it identifies data types when the HTML is rendered
		witnessScript := segwit.GetWitnessScript ()
		if !witnessScript.IsNil () { fields [fieldCount - 1].SetType ("<<< SERIALIZED WITNESS SCRIPT >>>") }

		tapScript, tapScriptIndex := segwit.GetTapScript ()
		if !tapScript.IsNil () {

			cbIndex := segwit.GetControlBlockIndex ()
			cbLeafCount := 0
			if cbIndex != -1 {
				cbLeafCount = (len (fields [cbIndex].AsBytes ()) - 1) / 32
			} else {
				fmt.Println ("Segwit has tap script but no control block.")
			}

			// set the field types for the Taproot Segwit fields
			if segwit.HasAnnex () {
				annexIndex := fieldCount - 1
				fields [annexIndex].SetType (fmt.Sprintf ("Annex (%d Bytes)", len (fields [annexIndex].AsBytes ())))
			}

			fields [tapScriptIndex].SetType ("<<< SERIALIZED TAP SCRIPT >>>")

			leafCountLabel := "TapLea"
			if cbLeafCount == 1 { leafCountLabel += "f" } else { leafCountLabel += "ves" }
			fields [cbIndex].SetType (fmt.Sprintf ("Control Block (%d %s)", cbLeafCount, leafCountLabel))

			// set the field types for the Tap Script
			tapScriptFields := tapScript.GetFields ()
			for f, field := range tapScriptFields {
				if !field.IsOpcode () {
					tapScriptFields [f].SetType (btc.GetStackItemType (field.AsBytes (), true))
				}
			}
		}

		hexFieldsHtml = make ([] FieldHtmlData, fieldCount)
		textFieldsHtml = make ([] FieldHtmlData, fieldCount)
		typeFieldsHtml = make ([] FieldHtmlData, fieldCount)

		for f, field := range fields {

			// set any field types that aren't already set
			if len (fields [f].AsType ()) == 0 {
				fields [f].SetType (btc.GetStackItemType (field.AsBytes (), usingSchnorrSignatures))
			}

			// hex strings
			entireHexField := field.AsHex (0)
			hexFieldsHtml [f] = FieldHtmlData { DisplayText: field.AsHex (maxCharWidth), ShowCopyButton: false }
			if hexFieldsHtml [f].DisplayText != entireHexField {
				hexFieldsHtml [f].ShowCopyButton = true
				hexFieldsHtml [f].CopyText = entireHexField
			}

			// text strings
			entireTextField := field.AsText (0)
			textFieldsHtml [f] = FieldHtmlData { DisplayText: field.AsText (maxCharWidth), ShowCopyButton: false }
			if textFieldsHtml [f].DisplayText != entireTextField {
				textFieldsHtml [f].ShowCopyButton = true
				textFieldsHtml [f].CopyText = entireTextField
			}

			// field types
			typeFieldsHtml [f] = FieldHtmlData { DisplayText: fields [f].AsType (), ShowCopyButton: false }
		}
	}

	copyImageUrl := app.Settings.GetFullUrl () + "/image/clipboard-copy.png"

	fieldSet := FieldSetHtmlData { HtmlId: htmlId, DisplayTypeClassPrefix: displayTypeClassPrefix, CharWidth: maxCharWidth, HexFields: hexFieldsHtml, TextFields: textFieldsHtml, TypeFields: typeFieldsHtml, CopyImageUrl: copyImageUrl }
	htmlData := SegwitHtmlData { FieldSet: fieldSet, IsEmpty: segwit.IsEmpty () }

	htmlData.WitnessScript = getScriptHtmlData (segwit.GetWitnessScript (), htmlId + "-witness-script", displayTypeClassPrefix)
	tapScript, _ := segwit.GetTapScript ()
	htmlData.TapScript = getScriptHtmlData (tapScript, htmlId + "-tap-script", displayTypeClassPrefix)

	return htmlData
}

func getScriptHtmlData (script btc.Script, htmlId string, displayTypeClassPrefix string) ScriptHtmlData {

	if script.IsNil () {
		return ScriptHtmlData { IsNil: true}
	}

	const maxCharWidth = uint (89)

	scriptHtmlData := ScriptHtmlData { FieldSet: FieldSetHtmlData { HtmlId: htmlId, DisplayTypeClassPrefix: displayTypeClassPrefix, CharWidth: maxCharWidth }, IsNil: false, IsOrdinal: script.IsOrdinal () }

	if script.IsEmpty () {
		scriptHtmlData.FieldSet.HexFields = [] FieldHtmlData { FieldHtmlData { DisplayText: "Empty", ShowCopyButton: false } }
		scriptHtmlData.FieldSet.TextFields = [] FieldHtmlData { FieldHtmlData { DisplayText: "Empty", ShowCopyButton: false } }
		scriptHtmlData.FieldSet.TypeFields = [] FieldHtmlData { FieldHtmlData { DisplayText: "Empty", ShowCopyButton: false } }
		return scriptHtmlData
	}

	fieldCount := script.GetFieldCount ()
	if script.HasParseError () { fieldCount++ }

	hexFieldsHtml := make ([] FieldHtmlData, fieldCount)
	textFieldsHtml := make ([] FieldHtmlData, fieldCount)
	typeFieldsHtml := make ([] FieldHtmlData, fieldCount)

	for f, field := range script.GetFields () {

		// hex strings
		entireHexField := field.AsHex (0)
		hexFieldsHtml [f] = FieldHtmlData { DisplayText: field.AsHex (maxCharWidth), ShowCopyButton: false }
		if hexFieldsHtml [f].DisplayText != entireHexField {
			hexFieldsHtml [f].ShowCopyButton = true
			hexFieldsHtml [f].CopyText = entireHexField
		}

		// text strings
		entireTextField := field.AsText (0)
		textFieldsHtml [f] = FieldHtmlData { DisplayText: field.AsText (maxCharWidth), ShowCopyButton: false }
		if textFieldsHtml [f].DisplayText != entireTextField {
			textFieldsHtml [f].ShowCopyButton = true
			textFieldsHtml [f].CopyText = entireTextField
		}

		// field types
		typeFieldsHtml [f] = FieldHtmlData { DisplayText: field.AsType (), ShowCopyButton: false }
	}

	if script.HasParseError () {
		parseErrorStr := "<<< PARSE ERROR >>>"
		hexFieldsHtml [fieldCount - 1] = FieldHtmlData { DisplayText: parseErrorStr, ShowCopyButton: false }
		textFieldsHtml [fieldCount - 1] = FieldHtmlData { DisplayText: parseErrorStr, ShowCopyButton: false }
		typeFieldsHtml [fieldCount - 1] = FieldHtmlData { DisplayText: parseErrorStr, ShowCopyButton: false }
	}

	copyImageUrl := app.Settings.GetFullUrl () + "/image/clipboard-copy.png"

	scriptHtmlData.FieldSet.HexFields = hexFieldsHtml
	scriptHtmlData.FieldSet.TextFields = textFieldsHtml
	scriptHtmlData.FieldSet.TypeFields = typeFieldsHtml
	scriptHtmlData.FieldSet.CopyImageUrl = copyImageUrl

	return scriptHtmlData
}

func getPreviousOutputScriptHtml (script btc.Script, htmlId string, displayTypeClassPrefix string) string {

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

// TODO: this needs to be replaced by a web-dir setting
func GetPath () string {
	return "web/"
}

