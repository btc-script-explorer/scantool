package btc

import (
	"fmt"
	"time"
	"sort"
	"strconv"
	"strings"
	"bytes"
"html/template"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"

	"btctx/app"
)

type Block struct {
	hash string
	previousHash string
	nextHash string
	height uint32
	timestamp int64
	txs [] Tx
}

func NewBlock (hash string, previous string, next string, height uint32, timestamp int64, txs [] Tx) Block {
	return Block { hash: hash, previousHash: previous, nextHash: next, height: height, timestamp: timestamp, txs: txs }
}

func (b *Block) IsNil () bool {
	return len (b.hash) == 0
}

func (b *Block) GetHash () string {
	return b.hash
}

func (b *Block) GetPreviousHash () string {
	return b.previousHash
}

func (b *Block) GetNextHash () string {
	return b.nextHash
}

func (b *Block) GetHeight () uint32 {
	return b.height
}

func (b *Block) GetTxs () [] Tx {
	return b.txs
}

func (b *Block) GetTx (index int) Tx {
	return b.txs [index]
}

type ElementTypeHTML struct {
	Label string
	Count uint32
	Percent string
}

// returns inputs, outputs
func (b *Block) GetInputOutputCounts () (int, int) {

	inputCount := 0
	outputCount := 0
	for _, tx := range b.txs {
		inputCount += tx.GetInputCount ()
		outputCount += tx.GetOutputCount ()
	}
	return inputCount, outputCount
}

func (b *Block) GetPendingPreviousOutputs () map [string] [] uint32 {

	unknownPreviousOutputTypes := make (map [string] [] uint32)
	for _, tx := range b.txs {
		for _, input := range tx.GetInputs () {
			if input.IsCoinbase () { continue }

			spendType := input.GetSpendType ()
			if len (spendType) == 0 {
				unknownPreviousOutputTypes [input.previousOutputTxId] = append (unknownPreviousOutputTypes [input.previousOutputTxId], input.GetPreviousOutputIndex ())
			}
		}
	}

	return unknownPreviousOutputTypes
}

func (b *Block) GetKnownSpendTypes () ([] ElementTypeHTML, int) {

	nonCoinbaseInputCount := 0
	knownSpendTypeCount := 1 // starting at 1 because the coinbase input always has a known spend type
	spendTypeMap := make (map [string] int)
	for _, tx := range b.txs {
		for _, input := range tx.GetInputs () {
			if input.IsCoinbase () { continue }

			nonCoinbaseInputCount++
			spendType := input.GetSpendType ()
			if len (spendType) > 0 {
				spendTypeMap [spendType]++
				knownSpendTypeCount++
			}
		}
	}

	var knownSpendTypes [] ElementTypeHTML
	for spendType, num := range spendTypeMap {
		knownSpendTypes = append (knownSpendTypes, ElementTypeHTML { Label: spendType, Count: uint32 (num), Percent: fmt.Sprintf ("%9.2f%%", float32 (num * 100) / float32 (nonCoinbaseInputCount)) })
	}

	return knownSpendTypes, knownSpendTypeCount
}

func (b *Block) GetOutputTypes () [] ElementTypeHTML {

	outputCount := 0
	outputTypeMap := make (map [string] int)
	for _, tx := range b.txs {
		for _, output := range tx.GetOutputs () {

			outputCount++
			outputType := output.GetOutputType ()
			if len (outputType) > 0 {
				outputTypeMap [outputType]++
			}
		}
	}

	var outputTypes [] ElementTypeHTML
	for outputType, num := range outputTypeMap {
		outputTypes = append (outputTypes, ElementTypeHTML { Label: outputType, Count: uint32 (num), Percent: fmt.Sprintf ("%9.2f%%", float32 (num * 100) / float32 (outputCount)) })
	}

	return outputTypes
}

type TxHTML struct {
	Index int
	TxId string
	Bip141 bool
	InputCount int
	OutputCount int
}

func (b *Block) GetHtmlData () map [string] interface {} {
	htmlData := make (map [string] interface {})

	// get block data
	htmlData ["Height"] = b.height
	htmlData ["Time"] = time.Unix (b.timestamp, 0).UTC ()
	htmlData ["Hash"] = b.hash
	if len (b.previousHash) > 0 { htmlData ["PreviousHash"] = b.previousHash }
	if len (b.nextHash) > 0 { htmlData ["NextHash"] = b.nextHash }
	htmlData ["TxCount"] = len (b.txs)

	// get the numbers of inputs and outputs and their types, and the tx detail
	var txDetail [] TxHTML
	bip141Count := 0
	inputCount := 0
	outputCount := float32 (0)

	witnessScriptMultisigCount := float32 (0)
	witnessScriptCount := float32 (0)

	tapScriptOrdinalCount := float32 (0)
	tapScriptCount := float32 (0)

	spendTypeMap := make (map [string] int)
	outputTypeMap := make (map [string] int)
	for t, tx := range b.txs {

		txDetail = append (txDetail, TxHTML { Index: t, TxId: tx.GetTxId (), Bip141: tx.SupportsBip141 (), InputCount: tx.GetInputCount (), OutputCount: tx.GetOutputCount () })
		if tx.SupportsBip141 () { bip141Count++ }

		for _, input := range tx.GetInputs () {

			st := input.spendType
			if st == OUTPUT_TYPE_P2WSH || st == SPEND_TYPE_P2SH_P2WSH {
				witnessScriptCount++
				if input.multisigWitnessScript { witnessScriptMultisigCount++ }
			} else if st == SPEND_TYPE_P2TR_Script {
				tapScriptCount++
				if input.ordinalTapScript {
					tapScriptOrdinalCount++
				} //else { fmt.Println (fmt.Sprintf ("Non-Ordinal Tap Script in tx %s, input %d.", tx.GetTxId (), i)) }
			}

			inputCount++
			spendTypeMap [input.GetSpendType ()]++
		}

		for _, output := range tx.GetOutputs () {
			outputCount++
			outputTypeMap [output.GetOutputType ()]++
		}
	}

	if witnessScriptCount > 0 || tapScriptCount > 0 {
		countStrLen := 0
		totalStrLen := 0

		wsCountStr := ""
		wsTotalStr := ""

		tsCountStr := ""
		tsTotalStr := ""

		// measuring the lengths of strings so the numbers line up on the UI
		if witnessScriptCount > 0 {
			wsCountStr = fmt.Sprintf ("%d", uint (witnessScriptMultisigCount))
			if len (wsCountStr) > countStrLen { countStrLen = len (wsCountStr) }

			wsTotalStr = fmt.Sprintf ("%d", uint (witnessScriptCount))
			if len (wsTotalStr) > totalStrLen { totalStrLen = len (wsTotalStr) }
		}

		if tapScriptCount > 0 {
			tsCountStr = fmt.Sprintf ("%d", uint (tapScriptOrdinalCount))
			if len (tsCountStr) > countStrLen { countStrLen = len (tsCountStr) }

			tsTotalStr = fmt.Sprintf ("%d", uint (tapScriptCount))
			if len (tsTotalStr) > totalStrLen { totalStrLen = len (tsTotalStr) }
		}

		// creating the messages
		if witnessScriptCount > 0 {
			wsMessage:= fmt.Sprintf ("%6.2f%% MultiSig (%*s/%*s)", (witnessScriptMultisigCount * 100) / witnessScriptCount, countStrLen, wsCountStr, totalStrLen, wsTotalStr)
			htmlData ["WitnessScriptMultiSigMessage"] = template.HTML (strings.Replace (wsMessage, " ", "&nbsp;", -1))
		}

		if tapScriptCount > 0 {
			tsMessage := fmt.Sprintf ("%6.2f%% Ordinals (%*s/%*s)", (tapScriptOrdinalCount * 100) / tapScriptCount, countStrLen, tsCountStr, totalStrLen, tsTotalStr)
			htmlData ["TapScriptOrdinalsMessage"] = template.HTML (strings.Replace (tsMessage, " ", "&nbsp;", -1))
		}
	}

	htmlData ["Bip141Percent"] = fmt.Sprintf ("%9.2f%%", float32 (bip141Count * 100) / float32 (len (b.txs)))

	htmlData ["BaseUrl"] = app.Settings.GetFullUrl ()
	htmlData ["TxDetail"] = txDetail

	htmlData ["InputCount"] = inputCount
	htmlData ["OutputCount"] = outputCount

	return htmlData
}

func extractBodyFromHTML (html string) string {

	bodyBegin := strings.Index (html, "<body>")
	if bodyBegin > -1 { bodyBegin += 6 } else { fmt.Println ("Body not found in chart HTML.") }

	bodyEnd := strings.Index (html, "</body>")
	if bodyEnd == -1 { fmt.Println ("Body not found in chart HTML.") }

	body := html [: bodyEnd]
	return body [bodyBegin :]
}

func GetBlockCharts (nonCoinbaseInputCount uint32, outputCount uint32, spendTypes map [string] uint32, outputTypes map [string] uint32) map [string] string {

	const pieRadius = 90
	const verticalPadding = 10
	longestLabel := 0

	// gather data for the spend type and output type charts

	spendTypeNames := [] string { OUTPUT_TYPE_P2PK, OUTPUT_TYPE_MultiSig, OUTPUT_TYPE_P2PKH, OUTPUT_TYPE_P2SH, SPEND_TYPE_P2SH_P2WPKH, SPEND_TYPE_P2SH_P2WSH, OUTPUT_TYPE_P2WPKH, OUTPUT_TYPE_P2WSH, SPEND_TYPE_P2TR_Key, SPEND_TYPE_P2TR_Script, SPEND_TYPE_NonStandard }

	var spendTypesHTML [] ElementTypeHTML
	for _, typeName := range spendTypeNames {
		if spendTypes [typeName] > 0 {
			if len (typeName) > longestLabel { longestLabel = len (typeName) }
			spendTypesHTML = append (spendTypesHTML, ElementTypeHTML { Label: typeName, Count: spendTypes [typeName], Percent: fmt.Sprintf ("%9.2f%%", float32 (spendTypes [typeName] * 100) / float32 (nonCoinbaseInputCount)) })
		}
	}

	outputTypeNames := [] string { OUTPUT_TYPE_P2PK, OUTPUT_TYPE_MultiSig, OUTPUT_TYPE_P2PKH, OUTPUT_TYPE_P2SH, OUTPUT_TYPE_P2WPKH, OUTPUT_TYPE_P2WSH, OUTPUT_TYPE_TAPROOT, OUTPUT_TYPE_OP_RETURN, OUTPUT_TYPE_WitnessUnknown, OUTPUT_TYPE_NonStandard }

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

