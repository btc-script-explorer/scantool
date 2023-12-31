async function get_block_txs ()
{
console.log (block_tx_ids);

	var txs_loaded = 0;
	var bip141_count = 0;
	var input_count = 1; // starting with 1 for the coinbase input
	var output_count = 0;
	var tx_count = block_tx_ids.length;
	for (var t = 0; t < tx_count; t++)
	{
		const headers = new Headers ();
		headers.append ("Content-Type", "application/json");
		const response = await fetch (base_url_web + '/block-tx/' + block_tx_ids [t] + '/' + t);
		const data = await response.json ();
		$ ('#tx-count').html (++txs_loaded);
		if (data.bip141)
			++bip141_count;
		$ ('#bip141-percent').html (Number ((bip141_count * 100) / txs_loaded).toFixed (2));

		input_count += data.input_count;
		$ ('#input-count').html (input_count);

		output_count += data.output_count;
		$ ('#output-count').html (output_count);

		$ ('#txs').append (data.tx_html);

		var block_load_percent = Number (((t + 1) * 100) / tx_count).toFixed (2);
		$ ('#block-load-status-bar').css ('width', block_load_percent + '%');
		$ ('#block-load-status-percent').html (block_load_percent + '%');
	}

	$ ('#block-load-status').css ('display', 'none')
}

async function get_tx_inputs ()
{
	var input_count = tx_inputs.length;
	var tx_value_out = Number ($ ('#tx-value-out').html ());
	for (var i = 0; i < input_count; i++)
	{
		const headers = new Headers ();
		headers.append ("Content-Type", "application/json");
		var request_data = { method: 'POST', headers: headers, body: JSON.stringify ({ tx_id: tx_inputs [i].tx_id, input_index: tx_inputs [i].input_index }) };
		const response = await fetch (base_url_web + '/input', request_data);
		const data = await response.json ();

		$ ('#input-minimized-' + i + '-spend-type').html (data.spend_type)
		$ ('#input-minimized-' + i + '-value').html (get_value_html (data.value_in))
		$ ('#input-minimized-' + i + '-address').html (data.address)
		$ ('#input-maximized-' + i).html (data.input_html)

		if (data.spend_type != 'COINBASE')
		{
			var tx_value_in = Number ($ ('#tx-value-in').text ()) + Number (data.value_in);
			$ ('#tx-value-in').text (tx_value_in);
			var tx_fee = Number ($ ('#tx-fee').text ());
			if (tx_value_in >= tx_value_out)
				$ ('#tx-fee').html (tx_value_in - tx_value_out);
		}

		var tx_load_percent = Number (((i + 1) * 100) / input_count).toFixed (2);
		$ ('#tx-load-status-bar').css ('width', tx_load_percent + '%');
		$ ('#tx-load-status-percent').html (tx_load_percent + '%');
	}

	$ ('#tx-load-status').css ('display', 'none')

	$ ('#tx-value-in').html (get_value_html ($ ('#tx-value-in').text ()));
	$ ('#tx-value-out').html (get_value_html ($ ('#tx-value-out').text ()));
	$ ('#tx-fee').html (get_value_html ($ ('#tx-fee').text ()));
}

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

	$ ('#current-block').html (data.current_block_height);
}

$ (document).ready (
function ()
{
	if (typeof block_tx_ids !== 'undefined')
		get_block_txs ();
	else if (typeof tx_inputs !== 'undefined')
		get_tx_inputs ()
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

