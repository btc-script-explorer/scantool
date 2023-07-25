package btc

import (
	"fmt"
	"strconv"
	"time"
	"bytes"
	"sort"
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
	Count int
	Percent string
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
	inputCount := 0
	outputCount := float32 (0)
	spendTypeMap := make (map [string] int)
	outputTypeMap := make (map [string] int)
	for i, tx := range b.txs {

		txDetail = append (txDetail, TxHTML { Index: i, TxId: tx.GetTxId (), Bip141: tx.SupportsBip141 (), InputCount: tx.GetInputCount (), OutputCount: tx.GetOutputCount () })

		for _, input := range tx.GetInputs () {
			inputCount++
if input.GetSpendType () == "Non-Standard" {
	fmt.Println (tx.GetTxId ())
}
			spendTypeMap [input.GetSpendType ()]++
		}

		for _, output := range tx.GetOutputs () {
			outputCount++
			outputTypeMap [output.GetOutputType ()]++
		}
	}

	settings := app.GetSettings ()
	htmlData ["BaseUrl"] = "http://" + settings.Website.GetFullUrl ()
	htmlData ["TxDetail"] = txDetail

	htmlData ["InputCount"] = inputCount
	htmlData ["OutputCount"] = outputCount
	nonCoinbaseInputCount := float32 (inputCount - 1)

	// for the pie charts
	const pieRadius = 90
	const verticalPadding = 10
	longestLabel := 0

	// gather data for the spend type and output type charts

	spendTypeNames := [] string { OUTPUT_TYPE_P2PK, OUTPUT_TYPE_MultiSig, OUTPUT_TYPE_P2PKH, OUTPUT_TYPE_P2SH, SPEND_TYPE_P2SH_P2WPKH, SPEND_TYPE_P2SH_P2WSH, OUTPUT_TYPE_P2WPKH, OUTPUT_TYPE_P2WSH, SPEND_TYPE_P2TR_Key, SPEND_TYPE_P2TR_Script, SPEND_TYPE_NonStandard }

	var spendTypesHTML [] ElementTypeHTML
	for _, typeName := range spendTypeNames {
		if spendTypeMap [typeName] > 0 {
			if len (typeName) > longestLabel { longestLabel = len (typeName) }
			spendTypesHTML = append (spendTypesHTML, ElementTypeHTML { Label: typeName, Count: spendTypeMap [typeName], Percent: fmt.Sprintf ("%9.2f%%", float32 (spendTypeMap [typeName] * 100) / nonCoinbaseInputCount) })
		}
	}

	outputTypeNames := [] string { OUTPUT_TYPE_P2PK, OUTPUT_TYPE_MultiSig, OUTPUT_TYPE_P2PKH, OUTPUT_TYPE_P2SH, OUTPUT_TYPE_P2WPKH, OUTPUT_TYPE_P2WSH, OUTPUT_TYPE_TAPROOT, OUTPUT_TYPE_OP_RETURN, OUTPUT_TYPE_WitnessUnknown, OUTPUT_TYPE_NonStandard }

	var outputTypesHTML [] ElementTypeHTML
	for _, typeName := range outputTypeNames {
		if outputTypeMap [typeName] > 0 {
			if len (typeName) > longestLabel { longestLabel = len (typeName) }
			outputTypesHTML = append (outputTypesHTML, ElementTypeHTML { Label: typeName, Count: outputTypeMap [typeName], Percent: fmt.Sprintf ("%9.2f%%", float32 (outputTypeMap [typeName] * 100) / outputCount) })
		}
	}

	outputTypeCount := len (outputTypesHTML)
	legendHeight := outputTypeCount
	legendHeight *= 22

	boxDimension := ((pieRadius + verticalPadding) * 2) + legendHeight
	boxDimensionStr := strconv.Itoa (boxDimension)


	if len (spendTypeNames) > 0 {

		sort.SliceStable (spendTypesHTML, func (i, j int) bool { return spendTypesHTML [i].Count > spendTypesHTML [j].Count })

		// spend type values
		var spendTypeValues [] opts.PieData
		for _, elementData := range spendTypesHTML {
			if spendTypeMap [elementData.Label] > 0 {
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
		htmlData ["SpendTypeChart"] = template.HTML (buff.String ())
	}

	sort.SliceStable (outputTypesHTML, func (i, j int) bool { return outputTypesHTML [i].Count > outputTypesHTML [j].Count })

	// output type values
	var outputTypeValues [] opts.PieData
	for _, elementData := range outputTypesHTML {
		if outputTypeMap [elementData.Label] > 0 {
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

//fmt.Printf ("%+v\n", pie)

	var buff bytes.Buffer
	pie.Render (&buff)
	htmlData ["OutputTypeChart"] = template.HTML (buff.String ())

	return htmlData
}

