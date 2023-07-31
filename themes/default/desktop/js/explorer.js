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

function get_transaction (query_id)
{
	if (typeof query_id != 'string' || query_id.length == 0 || !check_query_id_format (query_id))
	{
console.log (typeof query_id)
		console.log (query_id + ' is not a valid block id, transaction id or bitcoin address.');
		return;
	}

	var url = window.location.protocol + '//' + window.location.host + '/search/' + query_id;
	window.location.href = url;
}

function handle_pending_previous_outputs ()
{
	var next_pending_previous_output = {};
	var txs_to_delete = [];
	if (typeof pending_previous_outputs != 'undefined')
	{
		for (var tx_id in pending_previous_outputs)
		{
			next_pending_previous_output [tx_id] = JSON.stringify (pending_previous_outputs [tx_id]);

			txs_to_delete.push (tx_id);
			if (txs_to_delete.length >= 3)
				break;
		}
	}

	if (txs_to_delete.length == 0)
	{
		$ ('.spend-type-percent-loaded').css ('visibility', 'hidden');

		// we gather all the spend types, output types and non-coinbase input count
		// then we return them to the server to create the charts
		var spend_types = {};
		var spend_type_divs = $ ('#spend-types').children ();
		for (var s = 0; s < spend_type_divs.length; s++)
		{
			var spend_type_parts = $ (spend_type_divs [s]).children ();
			spend_types [$ (spend_type_parts [0]).html ()] = Number ($ (spend_type_parts [1]).html ());
		}

		var output_types = {};
		var output_type_divs = $ ('#output-types').children ();
		for (var o = 0; o < output_type_divs.length; o++)
		{
			var output_type_parts = $ (output_type_divs [o]).children ();
			output_types [$ (output_type_parts [0]).html ()] = Number ($ (output_type_parts [1]).html ());
		}

		var block_chart_data = { NonCoinbaseInputCount: noncoinbase_input_count, OutputCount: output_count, SpendTypes: JSON.stringify (spend_types), OutputTypes: JSON.stringify (output_types) };
		block_chart_data.method = 'get_block_charts';

		$.ajax (
		{
			type: 'post',
			url: window.location.protocol + '//' + window.location.host + '/ajax',
			data: block_chart_data,
			dataType: 'json',
			success: function (data, textStatus, jqXHR)
			{
				$ ('#spend-types').html (data.SpendTypeChart);
				$ ('#output-types').html (data.OutputTypeChart);
			}
		});

		return;
	}

	next_pending_previous_output.method = 'get_pending_previous_outputs';
//console.log ('get_pending_previous_outputs request:', next_pending_previous_output);

	$.ajax (
	{
		type: 'post',
		url: window.location.protocol + '//' + window.location.host + '/ajax',
		data: next_pending_previous_output,
		dataType: 'json',
//		error: function (jqXHR, textStatus, errorThrown) {},
		success: function (data, textStatus, jqXHR)
		{
//console.log ('get_pending_previous_outputs response:', data);

			for (var output_type in data)
			{
				var type_count = data [output_type];
				if (output_type != 'P2PK' && output_type != 'MultiSig' && output_type != 'P2PKH' && output_type != 'P2SH')
					output_type = 'Non-Standard';
				

				var html_id_prefix = 'spend-type-' + output_type;
				var spend_type_div = $ ('#' + html_id_prefix);
				if (spend_type_div.length == 0)
				{
					// we need to add a new type
					$ ('#spend-types').append ('<div id="' + html_id_prefix + '" style="height:22px;"></div>');
					$ ('#' + html_id_prefix).append ('<div style="display:inline-block; font-family:monospace; width:20ch; text-align:left; padding-right:2ch;">' + output_type + '</div>');
					$ ('#' + html_id_prefix).append ('<div id="' + html_id_prefix + '-count" style="display:inline-block; font-family:monospace; padding:2px 2ch 2px 0; width:15ch; text-align:right;">' + type_count + '</div>');
				}
				else
					$ ('#' + html_id_prefix + '-count').html (Number ($ ('#' + html_id_prefix + '-count').html ()) + type_count);

				known_spend_type_count += type_count;
				spend_type_percent = ((known_spend_type_count * 100) / noncoinbase_input_count).toFixed (2);
				$ ('.spend-type-percent-loaded').html ('(' + spend_type_percent + '%)');
			}

			// get the next one
			// an interval could be used as a timer in case some of the responses are never received
//			delete pending_previous_outputs [tx_id];
			for (var i = 0; i < txs_to_delete.length; i++)
				delete pending_previous_outputs [txs_to_delete [i]];

			handle_pending_previous_outputs ();
/*
			else
			{
				$ ('#tx-value-in').html (get_value_html ($ ('#tx-value-in').html ()))
				$ ('#tx-value-out').html (get_value_html ($ ('#tx-value-out').html ()))
				$ ('#tx-fee').html (get_value_html ($ ('#tx-fee').html ()))
			}
*/
		}
//		complete: function (jqXHR, textStatus) {}
	});
}

function handle_pending_inputs ()
{
	if (pending_inputs.length == 0)
	{
		$ ('#tx-value-in').html (get_value_html ($ ('#tx-value-in').html ()))
		$ ('#tx-value-out').html (get_value_html ($ ('#tx-value-out').html ()))
		$ ('#tx-fee').html (get_value_html ($ ('#tx-fee').html ()))
		return;
	}

	var ajax_data = pending_inputs [0];
	ajax_data.method = 'get_previous_output';
//console.log ('getpreviousoutput request:', ajax_data);

	$.ajax (
	{
		type: 'post',
		url: window.location.protocol + '//' + window.location.host + '/ajax',
		data: ajax_data,
		dataType: 'json',
//		error: function (jqXHR, textStatus, errorThrown) {},
		success: function (data, textStatus, jqXHR)
		{
//console.log ('getpreviousoutput response:', data);

			// previous output type
			$ ('#input-maximized-' + data.Input_index + '-previous-output-type').html (data.Prev_out_type);

			// previous output value
			$ ('#input-minimized-' + data.Input_index + '-value').html (get_value_html (data.Prev_out_value));
			$ ('#input-maximized-' + data.Input_index + '-previous-output-value').html (get_value_html (data.Prev_out_value));

			// previous output address
			$ ('#input-minimized-' + data.Input_index + '-address').html (data.Prev_out_address);
			$ ('#input-maximized-' + data.Input_index + '-previous-output-address').html (data.Prev_out_address);

			// previous output script
			$ ('#input-maximized-' + data.Input_index + '-previous-output-script').html (data.Prev_out_script_html);

			// if the spend type is empty, it uses the same name as the output type
			var spend_type = $ ('#input-minimized-' + data.Input_index + '-spend-type').html ();
			if (spend_type.length == 0)
			{
				$ ('#input-minimized-' + data.Input_index + '-spend-type').html (data.Prev_out_type);
				$ ('#input-maximized-' + data.Input_index + '-spend-type').html (data.Prev_out_type);

				if (data.Prev_out_type == 'P2SH')
				{
					console.log ('Displaying alternate script for P2SH redemption.');
					var show_type = $ ('#input-maximized-' + data.Input_index + '-input-script').css ('display');
					$ ('#input-maximized-' + data.Input_index + '-input-script').css ('display', 'none');
					$ ('#input-maximized-' + data.Input_index + '-input-script-alternate').css ('display', show_type);
					$ ('#input-maximized-' + data.Input_index + '-redeem-script').css ('display', show_type);
				}
				else
				{
					if (data.Prev_out_type != 'P2PK' && data.Prev_out_type != 'MultiSig' && data.Prev_out_type != 'P2PKH')
					{
						$ ('#input-minimized-' + data.Input_index + '-spend-type').html ('Non-Standard');
						$ ('#input-maximized-' + data.Input_index + '-spend-type').html ('Non-Standard');
					}
				}
			}

			// update the tx value in
			var value_in = parseInt ($ ('#tx-value-in').html ()) + data.Prev_out_value;
			$ ('#tx-value-in').html (value_in);

			// update the tx fee
			var value_out = parseInt ($ ('#tx-value-out').html ());
			if (value_in >= value_out)
				$ ('#tx-fee').html (value_in - value_out);

			// get the next one
			// an interval could be used as a timer in case some of the responses are never received
			pending_inputs.splice (0, 1);
			if (pending_inputs.length > 0)
				handle_pending_inputs ();
			else
			{
				$ ('#tx-value-in').html (get_value_html ($ ('#tx-value-in').html ()))
				$ ('#tx-value-out').html (get_value_html ($ ('#tx-value-out').html ()))
				$ ('#tx-fee').html (get_value_html ($ ('#tx-fee').html ()))
			}
		}
//		complete: function (jqXHR, textStatus) {}
	});
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

function handle_resize ()
{
	var body_margin_top_css = $ ('body').css ('margin-top');
	var body_margin_top = Number (body_margin_top_css.substring (0, body_margin_top_css.length - 2));
	var body_margin_bottom_css = $ ('body').css ('margin-bottom');
	var body_margin_bottom = Number (body_margin_top_css.substring (0, body_margin_bottom_css.length - 2));

	var body_vertical_margin = body_margin_top + body_margin_bottom;

	var win_height = $ (window).outerHeight ();
	var page_height = $ ('#page').outerHeight ();
	if (page_height < win_height)
		$ ('#page').css ('height', (win_height - body_vertical_margin) + 'px');
}

var current_block_interval = null;
function check_for_new_block ()
{
	$.ajax (
	{
		type: 'post',
		url: window.location.protocol + '//' + window.location.host + '/ajax',
		data: { method: 'get_current_block' },
		dataType: 'json',
		success: function (data, textStatus, jqXHR)
		{
			$ ('#current-block').html (data.Current_block_height);
		}
	});
}

$ (document).ready (
function ()
{
//	handle_resize ();
//	$ (window).resize (handle_resize);

	if (typeof pending_inputs !== 'undefined' && Array.isArray (pending_inputs))
		handle_pending_inputs ();
	else if (typeof noncoinbase_input_count === 'number')
		handle_pending_previous_outputs ();
	else
	{
		// for coinbase transactions
		$ ('#tx-value-in').html (get_value_html ($ ('#tx-value-in').html ()))
		$ ('#tx-value-out').html (get_value_html ($ ('#tx-value-out').html ()))
		$ ('#tx-fee').html (get_value_html ($ ('#tx-fee').html ()))
	}

	// set up the Enter key handler
	$ ('#query-box').on ('keypress', function (e) { if (e.which == 0x0d) get_transaction ($ ('#query-box').val ()); })

	check_for_new_block ();
	current_block_interval = setInterval (check_for_new_block, 60000);
});

