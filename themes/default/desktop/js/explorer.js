function check_query_id_format (query_id)
{
	if (query_id.length != 64)
		return false;

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
	if (!check_query_id_format (query_id))
	{
		console.log (query_id + ' is not a valid block id, transaction id or bitcoin address.');
		return;
	}

	var url = window.location.protocol + '//' + window.location.host + '/tx/' + query_id;
	window.location.href = url;
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
	ajax_data.method = 'getpreviousoutput';
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

			// previous output value in the minimized input window
			$ ('#input-minimized-value-' + data.Input_index).html (get_value_html (data.Prev_out_value));

			// previous output address in the minimized input window
			var has_address_format = data.Prev_out_type == 'Taproot' || data.Prev_out_type == 'P2WPKH' || data.Prev_out_type == 'P2WSH' || data.Prev_out_type == 'P2PKH' || data.Prev_out_type == 'P2SH';
			$ ('#input-minimized-address-' + data.Input_index).html (has_address_format ? data.Prev_out_Address : 'No Address Format');

			// previous output box in the maximized input
			$ ('#input-' + data.Input_index + '-previous-output').html (data.Prev_out_html);

			// update the tx value in
			var value_in = parseInt ($ ('#tx-value-in').html ()) + data.Prev_out_value;
			$ ('#tx-value-in').html (value_in);

			// update the tx fee
			var value_out = parseInt ($ ('#tx-value-out').html ());
			if (value_in >= value_out)
				$ ('#tx-fee').html (value_in - value_out);

			var segwit_field_count = $ ('#segwit-' + data.Input_index).children ().length;

			// spend type
			var predicted_spend_type = $ ('#input-minimized-spend-type-' + data.Input_index).html ();
			if (data.Prev_out_type == 'Taproot')
			{
				// tap script
				if (predicted_spend_type == 'Taproot Script Path')
				{
					if (pending_inputs [0].Tap_script_index < segwit_field_count - 1)
						$ ('#segwit-' + data.Input_index + '-field-' + pending_inputs [0].Tap_script_index).html ('&lt;&lt;&lt; SERIALIZED TAP SCRIPT &gt;&gt;&gt;');
					else if (predicted_spend_type != 'Taproot Key Path' && predicted_spend_type != 'Taproot Script Path')
						console.log ('Expecting ' + (pending_inputs [0].Tap_script_index + 1) + ' segwit fields, but only found ' + segwit_field_count + ' for ' + predicted_spend_type + ' input.');
				}
				else if (predicted_spend_type != 'Taproot Key Path' && predicted_spend_type != 'Taproot Script Path')
					console.log ('Input ' + data.Input_index + ' has a ' + data.Prev_out_type + ' previous output type with ' + predicted_spend_type + ' spend type.');
			}
			else if (data.Prev_out_type == 'P2SH')
			{
/*
				var input_script_field_count = $ ('#input-script-' + data.Input_index).children ().length;

				// redeem script
				if (typeof $ ('#redeem-script-' + data.Input_index) !== 'undefined' && input_script_field_count > 0)
					$ ('#input-script-' + data.Input_index + '-field-' + (input_script_field_count - 1)).html ('&lt;&lt;&lt; SERIALIZED REDEEM SCRIPT &gt;&gt;&gt;');
				else
					console.log ('No redeem script for ' + predicted_spend_type + ' input.');
*/

				// witness script
				if (predicted_spend_type == 'P2SH-P2WSH')
				{
					if (segwit_field_count > 0)
						$ ('#segwit-' + data.Input_index + '-field-' + (segwit_field_count - 1)).html ('&lt;&lt;&lt; SERIALIZED WITNESS SCRIPT &gt;&gt;&gt;');
					else
						console.log ('No segwit fields for ' + predicted_spend_type + ' input.');
				}

				if (predicted_spend_type != 'P2SH-P2WPKH' && predicted_spend_type != 'P2SH-P2WSH' && predicted_spend_type != 'P2SH')
					console.log (data.Prev_out_type + ' output type with ' + predicted_spend_type + ' spend type.');
			}
			else if (data.Prev_out_type == 'P2WSH')
			{
				if (segwit_field_count > 0)
					$ ('#segwit-' + data.Input_index + '-field-' + (segwit_field_count - 1)).html ('&lt;&lt;&lt; SERIALIZED WITNESS SCRIPT &gt;&gt;&gt;');
				else
					console.log ('No segwit fields for ' + data.Prev_out_type + ' input.');
			}
			else if (data.Prev_out_type == 'P2WPKH')
			{
			}
			else // P2PK, MultiSig, P2PKH, Non-Standard
			{
				$ ('#input-minimized-spend-type-' + data.Input_index).html (data.Prev_out_type);
				$ ('#input-spend-type-' + data.Input_index).html (data.Prev_out_type);
			}

/*
			if (predicted_spend_type.length == 0)
			{
			}
			else
			{
				var p2sh_wrapped = (predicted_spend_type == 'P2SH-P2WPKH' || predicted_spend_type == 'P2SH-P2WSH') && data.Prev_out_type == 'P2SH';
				var taproot_script_path = predicted_spend_type == 'Taproot Script Path' && data.Prev_out_type == 'Taproot';
			}
*/

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

function toggle_script_view (html_id, view_type)
{
	var view_types = ['hex', 'text', 'type'];

	// hide the old type
	for (var t in view_types)
	{
		var old_element = $ ('#' + html_id + '-' + view_types [t]);
		if (old_element.css ('display') == 'block')
		{
			if (view_types [t] == view_type)
				return;

			old_element.css ('display', 'none');

			var old_element_button = $ ('#' + html_id + '-' + view_types [t] + '-button');
			old_element_button.css ('font-weight', 'normal');
			old_element_button.css ('cursor', 'pointer');

			break;
		}
	}

	// show the new type
	$ ('#' + html_id + '-' + view_type).css ('display', 'block');
	var new_element_button = $ ('#' + html_id + '-' + view_type + '-button');
	new_element_button.css ('font-weight', 'bold');
	new_element_button.css ('cursor', 'normal');
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

$ (document).ready (
function ()
{
//	handle_resize ();
//	$ (window).resize (handle_resize);

	if (typeof pending_inputs !== 'undefined' && Array.isArray (pending_inputs))
		handle_pending_inputs ();
	else
	{
		$ ('#tx-value-in').html (get_value_html ($ ('#tx-value-in').html ()))
		$ ('#tx-value-out').html (get_value_html ($ ('#tx-value-out').html ()))
		$ ('#tx-fee').html (get_value_html ($ ('#tx-fee').html ()))
	}

	// set up the Enter key handler
	$ ('#query-box').on ('keypress', function (e) { if (e.which == 0x0d) get_transaction ($ ('#query-box').val ()); })
});

