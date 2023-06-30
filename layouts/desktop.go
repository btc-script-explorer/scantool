package layouts

import (
)

type PageLayout interface {
	GetMainLayout (pageContent string, javascript string) string
//	GetExplorerPage () string
//	GetTxBasicsPage () string
	SetHTMLContent (html string, javascript string) string

	GetHeaderBar () string
	GetFooterBar () string
}

func GetLayout (isDesktop bool) PageLayout {
	return Desktop {}
}

type Desktop struct {
}

func (d Desktop) GetMainLayout (pageContent string, javascript string) string {
	return `
		<!DOCTYPE html>
		<html>
			<head>
				<link rel="stylesheet" type="text/css" href="/css/btc-tx.css" />
				<script type="text/javascript" src="/js/jquery-3.7.0.min.js"></script>
				<script type="text/javascript">` +
					javascript +
				`</script>
				<script type="text/javascript" src="/js/btc-tx.js"></script>
			</head>
			<body>
				<div id="page">` +
					d.GetHeaderBar () +
					`<div class="page-content">` +
						pageContent +
					`</div>` +
					d.GetFooterBar () +
				`</div>
			</body>
		</html>`
}

func (d Desktop) SetHTMLContent (html string, javascript string) string {
	return d.GetMainLayout (html, javascript)
}

func (d Desktop) GetHeaderBar () string {
	return `
		<div id="page-header">
			<a class="menu-item" href="/">Home</a>
<!--
			<a class="menu-item" href="/tx-basics/">Transaction Basics</a>
			<a class="menu-item" href="/">Transactions of Interest</a>
			<a class="menu-item" href="/charts/">Charts</a>
			<a class="menu-item" href="/faw/">FAQ</a>
-->
		</div>`
}

func (d Desktop) GetFooterBar () string {
	return `
		<div id="page-footer">
			Copyright &#xA9;2023. All rights reserved.
		</div>`
}

