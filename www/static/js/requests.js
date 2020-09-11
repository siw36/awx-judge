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

  $(obj).prop("disabled", true);
  // Fire off the request
  request = $.ajax({
    url: "/api/v1/requests/remove",
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
    $(obj).prop("disabled", false);
  });

};

function loadTable() {
  request = $.ajax({
    url: '/api/v1/requests/list',
    type: "GET",
    // beforeSend: function() {
    //     $('#current_page').append("loading..");
    //     },
    // success: finished //Change to this
  });

  // Callback handler that will be called on success
  request.done(function (response, textStatus, jqXHR){
    // Handler for no items in cart
    if (response.length <= 0) {
      $('#requests').html('There are no requests')
      return
    }
    // Update table
    $(function() {
      if ('content' in document.createElement('template')) {
        $.each(response, function(i, item) {
          var t = document.querySelector('#request_item_template'),
          // Setting all requeired elements
          request_item = t.content.querySelector('tr');
          request_icon = t.content.querySelector('#request_icon');
          request_name = t.content.querySelector('#request_name');
          request_reason = t.content.querySelector('#request_reason');
          request_state = t.content.querySelector('#request_state');
          request_judge_reason = t.content.querySelector('#request_judge_reason');
          request_button_reorder = t.content.querySelector('#request_button_reorder');
          request_button_delete = t.content.querySelector('#request_button_delete');

          request_item.id = item.id;
          if (item.icon != "") {
            request_icon.setAttribute('src', item.survey.icon);
          } else {
            request_icon.setAttribute('src', '/static/logo.png');
          }
          request_name.textContent = item.survey.name;
          request_reason.textContent = item.request_reason;
          request_state.textContent = item.state;
          request_judge_reason.textContent = item.reason;
          request_button_reorder.setAttribute('data-request_id', item.id);
          request_button_reorder.setAttribute('onclick', 'reorderItem(this)');
          request_button_delete.setAttribute('data-request_id', item.id);
          request_button_delete.setAttribute('onclick', 'deleteItem(this)');
          if (item.state != "pending"){
            request_button_delete.disabled = true;
          } else {
            request_button_delete.disabled = false;
          }

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
