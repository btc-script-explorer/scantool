function check_tx_hash_format (hash)
{
	if (hash.length != 64)
		return false;

	hash = hash.toLowerCase ();
	var allowed_chars = '0123456789abcdef';
	for (var i = 0; i < hash.length; i++)
	{
		if (allowed_chars.indexOf (hash.charAt (i)) == -1)
			return false;
	}

	return true;
}

function get_transaction (hash)
{
	if (!check_tx_hash_format (hash))
	{
		console.log (hash + ' is not a valid transaction id.');
		pending_inputs = [];
		tx_id = '';
		return;
	}

	$ ('#tx').html ('');

	$.ajax (
	{
		type: 'post',
		url: 'ajax',
		data: { method: 'gettx', hash: hash },
		dataType: 'json',
//		error: function (jqXHR, textStatus, errorThrown) {},
		success: function (data, textStatus, jqXHR)
		{
//console.log ('gettx response:'); console.log (data);

			tx_id = hash;
			$ ('#tx').html (data.Tx_html);
			pending_inputs = data.Pending_inputs;
			if (pending_inputs.length > 0)
				handle_pending_inputs ();
		}
//		complete: function (jqXHR, textStatus) {}
	});
}

var pending_inputs = [];
var tx_id = '';
function handle_pending_inputs ()
{
	if (pending_inputs.length == 0)
		return;

	var ajax_data = pending_inputs [0];
	ajax_data.tx_id = tx_id;
	ajax_data.method = 'getpreviousoutput';
//console.log ('getpreviousoutput request:'); console.log (ajax_data);

	$.ajax (
	{
		type: 'post',
		url: 'ajax',
		data: ajax_data,
		dataType: 'json',
//		error: function (jqXHR, textStatus, errorThrown) {},
		success: function (data, textStatus, jqXHR)
		{
//console.log ('getpreviousoutput response:'); console.log (data);

			$ ('#input-address-' + data.Input_index).html (data.Address);

			$ ('#input-value-' + data.Input_index).html (data.Value);
			var value_in = parseInt ($ ('#tx-value-in').html ()) + data.Value;
			$ ('#tx-value-in').html (value_in);

			var value_out = parseInt ($ ('#tx-value-out').html ());
			if (value_in >= value_out)
				$ ('#tx-fee').html (value_in - value_out);

			var input_tx_type = $ ('#input-tx-type-' + data.Input_index).html ();
			if (input_tx_type.length == 0)
			{
				if (data.Output_type == 'Taproot')
					$ ('#input-tx-type-' + data.Input_index).html ('Taproot Key Path');
				else
					$ ('#input-tx-type-' + data.Input_index).html (data.Output_type);
			}
			else
			{
				var p2sh_wrapped = (input_tx_type == 'P2SH-P2WPKH' || input_tx_type == 'P2SH-P2WSH') && data.Output_type == 'P2SH';
				var taproot_script_path = input_tx_type == 'Taproot Script Path' && data.Output_type == 'Taproot';
				if (!p2sh_wrapped && !taproot_script_path)
				{
					console.log (data.Output_type + ' incorrectly identified as ' + input_tx_type);
					$ ('#input-tx-type-' + data.Input_index).html (data.Output_type);
				}
			}

			// get the next one
			pending_inputs.splice (0, 1);
			if (pending_inputs.length > 0)
				handle_pending_inputs ();
		}
//		complete: function (jqXHR, textStatus) {}
	});
}

$ (document).ready (
function ()
{
	$ ('#h').on ('keypress', function (e) { if (e.which == 0x0d) get_transaction ($ ('#h').val ()); })
});

