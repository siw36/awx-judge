var request;
function deleteItem(obj){
  // Abort any pending request
  if (request) {
    request.abort();
  }
  // Setup variables
  var id = $(obj).data("request_id");
  var data = new Object;
  data["id"] = id
  var json_data = JSON.stringify(data)

  $(obj).disabled = true;
  // Fire off the request
  request = $.ajax({
    url: "/api/v1/cart/remove",
    type: "POST",
    contentType: "application/json; charset=utf-8",
    data: json_data
  });

  // Callback handler that will be called on success
  request.done(function (response, textStatus, jqXHR){
    // Log a message to the console
    console.log("Delete cart item successful");
    // Update table
    $('#' + id).hide('slow', function(){
      $('#' + id).remove();
      // Disable submit button when no items are in cart
      if ( $('#cart tr').length == 0 ){
        $("#submit_request").prop("disabled", true);
      };
    });
  });

  // Callback handler that will be called on failure
  request.fail(function (jqXHR, textStatus, errorThrown){
    // Log the error to the console
    console.error(
      "The following error occurred: "+
      textStatus, errorThrown
    );
    // Display alert
    alert("Something is wrong. Detailed information in console log.")
  });

  // Callback handler that will be called regardless
  // if the request failed or succeeded
  request.always(function () {
    // Reenable the inputs
    $(obj).disabled = false;
  });
};

function loadTable() {
  request = $.ajax({
    url: '/api/v1/cart/list',
    type: "GET",
  });

  // Callback handler that will be called on success
  request.done(function (response, textStatus, jqXHR){
    // Handler for no items in cart
    if (response.requests.length <= 0) {
      $('#loader').hide('slow', function(){ $('#loader').remove(); });
      $("#submit_request").prop("disabled", true);
      return
    }
    // Update table
    $(function() {
      if ('content' in document.createElement('template')) {
        $.each(response.requests, function(i, item) {
          var t = document.querySelector('#cart_item_template');
          var tc = document.importNode(t.content, true);

          // Setting all requeired elements
          tc.querySelector('tr').id = item.id;
          tc.querySelector('#cart_name').textContent = item.template.name;
          tc.querySelector('#cart_reason').textContent = item.request_reason;
          tc.querySelector('form').id = 'edit-' + item.id;
          tc.querySelector('#cart_template_id').setAttribute('value', item.template.id);
          tc.querySelector('#cart_request_id').setAttribute('value', item.id);
          tc.querySelector('#cart_button_edit').setAttribute('form', 'edit-' + item.id);
          tc.querySelector('#cart_button_delete').setAttribute('data-request_id', item.id);
          tc.querySelector('#cart_button_delete').setAttribute('onclick', 'deleteItem(this)');
          if (item.template.icon != "") {
            tc.querySelector('#cart_icon').src = item.template.icon;
          } else {
            tc.querySelector('#cart_icon').src = '/static/logo.png';
          }
          var tb = document.getElementsByTagName("tbody");
          tb[0].appendChild(tc);
        });
        // Hide loader
        $('#loader').hide('slow', function(){ $('#loader').remove(); });
        $("#submit_request").prop("disabled", false);
      } else {
        alert("HTML5 templating does not work with your browser")
      }
    });
  });
}

// CONSTRUCT TABLE
$(document).ready(loadTable());
