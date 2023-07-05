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
		return;

	var ajax_data = pending_inputs [0];
	ajax_data.method = 'getpreviousoutput';
//console.log ('getpreviousoutput request:'); console.log (ajax_data);

	$.ajax (
	{
		type: 'post',
		url: window.location.protocol + '//' + window.location.host + '/ajax',
		data: ajax_data,
		dataType: 'json',
//		error: function (jqXHR, textStatus, errorThrown) {},
		success: function (data, textStatus, jqXHR)
		{
//console.log ('getpreviousoutput response:'); console.log (data);

			$ ('#input-minimized-address-' + data.Input_index).html (data.Address);

			$ ('#input-minimized-value-' + data.Input_index).html (data.Value);
			var value_in = parseInt ($ ('#tx-value-in').html ()) + data.Value;
			$ ('#tx-value-in').html (value_in);

			var value_out = parseInt ($ ('#tx-value-out').html ());
			if (value_in >= value_out)
				$ ('#tx-fee').html (value_in - value_out);

			var input_tx_type = $ ('#input-minimized-tx-type-' + data.Input_index).html ();
			if (input_tx_type.length == 0)
			{
				if (data.Output_type == 'Taproot')
					$ ('#input-minimized-tx-type-' + data.Input_index).html ('Taproot Key Path');
				else
					$ ('#input-minimized-tx-type-' + data.Input_index).html (data.Output_type);
			}
			else
			{
				var p2sh_wrapped = (input_tx_type == 'P2SH-P2WPKH' || input_tx_type == 'P2SH-P2WSH') && data.Output_type == 'P2SH';
				var taproot_script_path = input_tx_type == 'Taproot Script Path' && data.Output_type == 'Taproot';
				if (!p2sh_wrapped && !taproot_script_path)
				{
					console.log (data.Output_type + ' incorrectly identified as ' + input_tx_type);
					$ ('#input-minimized-tx-type-' + data.Input_index).html (data.Output_type);
				}
			}

			// get the next one
			// an interval could be used as a timer in case some of the responses are never received
			pending_inputs.splice (0, 1);
			if (pending_inputs.length > 0)
				handle_pending_inputs ();
		}
//		complete: function (jqXHR, textStatus) {}
	});
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

	// set up the Enter key handler
	$ ('#query-box').on ('keypress', function (e) { if (e.which == 0x0d) get_transaction ($ ('#query-box').val ()); })
});

