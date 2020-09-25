var request;
function requestDelete(request_id){
  // Abort any pending request
  if (request) {
    request.abort();
  }
  // Setup variables
  var data = new Object;
  data["id"] = request_id;
  var json_data = JSON.stringify(data);

  request = $.ajax({
    url: "/api/v1/cart/remove",
    type: "POST",
    contentType: "application/json; charset=utf-8",
    data: json_data
  });

  request.done(function (response, textStatus, jqXHR){
    // Log a message to the console
    console.log("Delete cart item successful");
    // Update table
    $('#' + request_id).hide('slow', function(){
      $('#' + request_id).remove();
      // Disable submit button when no items are in cart
      if ( $('#cart tr').length == 0 ){
        $("#submit_request").prop("disabled", true);
      };
    });
  });

  request.fail(function (jqXHR, textStatus, errorThrown){
    // Log the error to the console
    console.error(
      "The following error occurred: "+
      textStatus, errorThrown
    );
    // Display alert
    alert("Something is wrong. Detailed information in console log.")
  });

  request.always(function () {
  });
};

function loadTable() {
  if (request) {
    request.abort();
  }
  request = $.ajax({
    url: '/api/v1/cart/list',
    type: "GET",
  });

  // Callback handler that will be called on success
  request.done(function (response, textStatus, jqXHR){
    // Handler for no items in cart
    if (response.requests == null || response.requests.length <= 0) {
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

          tc.querySelector('#cart_button_view').setAttribute('href', `/request?action=view&source=cart&template_id=${item.template.id}&request_id=${item.id}`);
          tc.querySelector('#cart_button_clone').setAttribute('href', `/request?action=clone&source=cart&template_id=${item.template.id}&request_id=${item.id}`);
          tc.querySelector('#cart_button_edit').setAttribute('href', `/request?action=edit&source=cart&template_id=${item.template.id}&request_id=${item.id}`);
          tc.querySelector('#cart_button_delete').setAttribute('onclick', `requestDelete('${item.id}')`);
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

function submitRequest() {
  // Abort any pending request
  if (request) {
    request.abort();
  }

  $(':button').prop('disabled', true);
  // Fire off the request
  request = $.ajax({
    url: "/api/v1/cart/execute",
    type: "POST",
    contentType: "application/json; charset=utf-8"
  });

  // Callback handler that will be called on success
  request.done(function (response, textStatus, jqXHR){
    // Log a message to the console
    console.log("Created requests based on cart intems");
    window.location.href="/requests";
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
    $(':button').prop('disabled', false);
  });
};

// CONSTRUCT TABLE
$(document).ready(loadTable());
