package themes

import (
)

type PageLayout interface {
	GetMainLayout (pageContent string, javascript string) string
	SetHTMLContent (html string, javascript string) string
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

