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

