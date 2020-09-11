// DELETE CART ITEM
var request;
// Bind to the submit event of our form
function deleteItem(obj){

  // Abort any pending request
  if (request) {
    request.abort();
  }
  // Setup variables
  var id = $(obj).data("request_id")
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
    $('#' + id).hide('slow', function(){ $('#' + id).remove(); });
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
    // beforeSend: function() {
    //     $('#current_page').append("loading..");
    //     },
    // success: finished //Change to this
  });

  // Callback handler that will be called on success
  request.done(function (response, textStatus, jqXHR){
    // Handler for no items in cart
    if (response.requests.length <= 0) {
      $("#submit_request").prop("disabled", true);
      return
    }
    // Update table
    $(function() {
      if ('content' in document.createElement('template')) {
        $.each(response.requests, function(i, item) {
          var t = document.querySelector('#cart_item_template'),
          // Setting all requeired elements
          cart_item = t.content.querySelector('tr');
          cart_icon = t.content.querySelector('#cart_icon');
          cart_name = t.content.querySelector('#cart_name');
          cart_reason = t.content.querySelector('#cart_reason');
          cart_button_edit = t.content.querySelector('#cart_button_edit');
          cart_button_delete = t.content.querySelector('#cart_button_delete');

          cart_item.id = item.id;
          if (item.survey.icon != "") {
            cart_icon.src = item.survey.icon;
          } else {
            cart_icon.src = '/static/logo.png';
          }
          cart_name.textContent = item.survey.name;
          cart_reason.textContent = item.request_reason;
          cart_button_edit.setAttribute('data-request_id', item.id);
          cart_button_edit.setAttribute ('onclick', 'editItem(this)');
          cart_button_delete.setAttribute('data-request_id', item.id);
          cart_button_delete.setAttribute ('onclick', 'deleteItem(this)');

          var tb = document.getElementsByTagName("tbody");
          var clone = document.importNode(t.content, true);
          tb[0].appendChild(clone);
        });
        $("#submit_request").prop("disabled", false);
      } else {
        alert("HTML5 templating does not work with your browser")
      }
    });
  });
}

// CONSTRUCT TABLE
$(document).ready(loadTable());
