function check_query_id_format (query_id)
{
	query_id = query_id.toLowerCase ();
	var allowed_chars = '0123456789abcdef';
	for (var i = 0; i < query_id.length; i++)
	{
		if (allowed_chars.indexOf (query_id.charAt (i)) == -1)
			return false;
	}

	return true;
}

function handle_search (query_id)
{
	if (typeof query_id != 'string' || query_id.length == 0 || !check_query_id_format (query_id))
	{
		console.log (query_id + ' is not a valid block id, transaction id or bitcoin address.');
		return;
	}

	window.location.href = base_url_web + '/search/' + query_id;
}

function get_pending_block_spend_types ()
{
	var pending_spend_types = {};

	// get the next set to request from the server
	if (typeof pending_block_spend_types != 'undefined')
	{
		var num = 0;
		for (var tx_id in pending_block_spend_types)
		{
			pending_spend_types [tx_id] = pending_block_spend_types [tx_id];
			if (++num >= 3)
				break;
		}
	}

	return pending_spend_types;
}

async function handle_pending_block_spend_types ()
{
	$ ('#toggle-charts-link').css ('display', 'none');
	$ ('#spend-type-status').css ('display', 'block');

	var spend_types_received = 0;

	var next_pending_spend_types = get_pending_block_spend_types ();
	while (!$.isEmptyObject (next_pending_spend_types))
	{
		var txs_to_delete = [];
		for (var tx_id in next_pending_spend_types)
			txs_to_delete.push (tx_id);

		if (txs_to_delete.length > 0)
		{
			// get the spend types from the server
			const headers = new Headers ();
			headers.append ("Content-Type", "application/json");
			var request_data = { method: 'POST', headers: headers, body: JSON.stringify (next_pending_spend_types) };
			const response = await fetch (base_url_web + '/legacy_spend_types', request_data);
			const data = await response.json ();

			// handle the response
			for (var outpoint in data)
			{
				if (typeof known_spend_types [data [outpoint]] != 'undefined')
					++known_spend_types [data [outpoint]];
				else
					known_spend_types [data [outpoint]] = 1;
				++known_spend_type_count;
				++spend_types_received;
				var spend_type_percent = ((spend_types_received * 100) / unknown_spend_type_count).toFixed (2);
				$ ('#spend-type-status-bar').css ('width', spend_type_percent + '%');
				$ ('#spend-type-status-percent').html (spend_type_percent + '%');
			}

			// get the next one
			// an interval could be used as a timer in case some of the responses are never received
			for (var i = 0; i < txs_to_delete.length; i++)
				delete pending_block_spend_types [txs_to_delete [i]];
		}

		next_pending_spend_types = get_pending_block_spend_types ();
	}

	// no more to get
	// we gather all the spend types, output types and non-coinbase input count
	// then we return them to the server to create the charts
	$ ('#spend-type-status').css ('display', 'none');

	var block_chart_data = { NonCoinbaseInputCount: known_spend_type_count, OutputCount: output_count, SpendTypes: known_spend_types, OutputTypes: output_types };

	const headers = new Headers ();
	headers.append ("Content-Type", "application/json");
	var request_data = { method: 'POST', headers: headers, body: JSON.stringify (block_chart_data) };
	const response = await fetch (base_url_web + '/block_charts', request_data);
	const data = await response.json ();

	$ ('#spend-types').html (data.SpendTypeChart);
	$ ('#output-types').html (data.OutputTypeChart);
	$ ('#type-charts').css ('display', 'block');
}

async function handle_pending_tx_previous_outputs ()
{
	if (pending_tx_previous_outputs.length == 0)
	{
		$ ('#tx-value-in').html (get_value_html ($ ('#tx-value-in').html ()))
		$ ('#tx-value-out').html (get_value_html ($ ('#tx-value-out').html ()))
		$ ('#tx-fee').html (get_value_html ($ ('#tx-fee').html ()))
		return;
	}

	while (pending_tx_previous_outputs.length > 0)
	{
		const headers = new Headers ();
		headers.append ("Content-Type", "application/json");
		var request_data = { method: 'POST', headers: headers, body: JSON.stringify (pending_tx_previous_outputs [0]) };
		const response = await fetch (base_url_web + '/previous_output', request_data);
		const data = await response.json ();

		// previous output type
		$ ('#input-maximized-' + data.InputIndex + '-previous-output-type').html (data.PrevOutType);

		// previous output value
		$ ('#input-minimized-' + data.InputIndex + '-value').html (get_value_html (data.PrevOutValue));
		$ ('#input-maximized-' + data.InputIndex + '-previous-output-value').html (get_value_html (data.PrevOutValue));

		// previous output address
		$ ('#input-minimized-' + data.InputIndex + '-address').html (data.PrevOutAddress);
		$ ('#input-maximized-' + data.InputIndex + '-previous-output-address').html (data.PrevOutAddress);

		// previous output script
		$ ('#input-maximized-' + data.InputIndex + '-previous-output-script').html (data.PrevOutScriptHtml);

		// if the spend type is empty, it uses the same name as the output type
		var spend_type = $ ('#input-minimized-' + data.InputIndex + '-spend-type').html ();
		if (spend_type.length == 0)
		{
			$ ('#input-minimized-' + data.InputIndex + '-spend-type').html (data.PrevOutType);
			$ ('#input-maximized-' + data.InputIndex + '-spend-type').html (data.PrevOutType);

/*
			if (data.PrevOutType == 'P2SH')
			{
				console.log ('Displaying alternate script for P2SH redemption: tx ' + $ ('#query-box').val () + ', input ' + data.InputIndex);
				var show_type = $ ('#input-maximized-' + data.InputIndex + '-input-script').css ('display');
				$ ('#input-maximized-' + data.InputIndex + '-input-script').css ('display', 'none');
				$ ('#input-maximized-' + data.InputIndex + '-input-script-alternate').css ('display', show_type);
				$ ('#input-maximized-' + data.InputIndex + '-redeem-script').css ('display', show_type);
			}
			else
			{
*/
				if (data.PrevOutType != 'P2PK' && data.PrevOutType != 'MultiSig' && data.PrevOutType != 'P2PKH')
				{
					$ ('#input-minimized-' + data.InputIndex + '-spend-type').html ('Non-Standard');
					$ ('#input-maximized-' + data.InputIndex + '-spend-type').html ('Non-Standard');
				}
//			}
		}

		// update the tx value in
		var value_in = parseInt ($ ('#tx-value-in').html ()) + data.PrevOutValue;
		$ ('#tx-value-in').html (value_in);

		// update the tx fee
		var value_out = parseInt ($ ('#tx-value-out').html ());
		if (value_in >= value_out)
			$ ('#tx-fee').html (value_in - value_out);

		// get the next one
		// an interval could be used as a timer in case some of the responses are never received
		pending_tx_previous_outputs.splice (0, 1);
	}

	$ ('#tx-value-in').html (get_value_html ($ ('#tx-value-in').html ()))
	$ ('#tx-value-out').html (get_value_html ($ ('#tx-value-out').html ()))
	$ ('#tx-fee').html (get_value_html ($ ('#tx-fee').html ()))
}

function get_value_html (value)
{
	var val_str = Number (value).toString ();
	if (val_str.length > 8)
	{
		var btc_digits = val_str.length - 8;
		val_str = '<span style="font-weight:bold;">' + val_str.substr (0, btc_digits) + '</span>' + val_str.substr (btc_digits);
	}

	return val_str;
}

async function copy_to_clipboard (data)
{
	await navigator.clipboard.writeText (data);
}

function toggle_inputs (event)
{
	var min = $ ('#inputs-minimized');
	var max = $ ('#inputs-maximized');
	if (min.css ('display') == 'block')
	{
		min.css ('display', 'none');
		max.css ('display', 'block');
		$ ('#input-toggle').html ('Hide');
	}
	else
	{
		min.css ('display', 'block');
		max.css ('display', 'none');
		$ ('#input-toggle').html ('Show');
	}
}

function toggle_outputs (event)
{
	var min = $ ('#outputs-minimized');
	var max = $ ('#outputs-maximized');
	if (min.css ('display') == 'block')
	{
		min.css ('display', 'none');
		max.css ('display', 'block');
		$ ('#output-toggle').html ('Hide');
	}
	else
	{
		min.css ('display', 'block');
		max.css ('display', 'none');
		$ ('#output-toggle').html ('Show');
	}
}

function toggle_script_view (html_id_prefix, off_class_prefix, on_class_prefix, view_type)
{
	var view_types = ['hex', 'text', 'type'];

	// hide all of the divs and turn off all of the buttons
	for (var t in view_types)
	{
		var element_id = html_id_prefix + view_types [t];
		var button_id = element_id + '-button';
		if (view_types [t] == view_type)
		{
			$ ('.' + element_id).css ('display', 'block');
			$ ('#' + button_id).removeClass (off_class_prefix);
			$ ('#' + button_id).addClass (on_class_prefix);
		}
		else
		{
			$ ('.' + element_id).css ('display', 'none');
			$ ('#' + button_id).removeClass (on_class_prefix);
			$ ('#' + button_id).addClass (off_class_prefix);
		}
	}
}

var current_block_interval = null;
async function check_for_new_block ()
{
	const headers = new Headers ();
	headers.append ("Content-Type", "application/json");
	var request_data = { method: 'GET', headers: headers };
	const response = await fetch (base_url_rest + '/current_block_height', request_data);
	const data = await response.json ();

	$ ('#current-block').html (data.Current_block_height);
}

$ (document).ready (
function ()
{
	if (typeof pending_tx_previous_outputs !== 'undefined' && Array.isArray (pending_tx_previous_outputs))
		handle_pending_tx_previous_outputs ();
	else if (typeof unknown_spend_type_count !== 'undefined' && unknown_spend_type_count > 0)
		$ ('#toggle-charts-link').css ('display', 'block');
	else
	{
		// for coinbase transactions
		$ ('#tx-value-in').html (get_value_html ($ ('#tx-value-in').html ()))
		$ ('#tx-value-out').html (get_value_html ($ ('#tx-value-out').html ()))
		$ ('#tx-fee').html (get_value_html ($ ('#tx-fee').html ()))
	}

	// set up the Enter key handler
	$ ('#query-box').on ('keypress', function (e) { if (e.which == 0x0d) handle_search ($ ('#query-box').val ()); })

	check_for_new_block ();
	current_block_interval = setInterval (check_for_new_block, 60000);
});

