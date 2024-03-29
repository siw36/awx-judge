var request;

function loadTable() {
  if (request) {
    request.abort();
  }
  request = $.ajax({
    url: '/api/v1/requests/list',
    type: "GET",
  });

  // Callback handler that will be called on success
  request.done(function (response, textStatus, jqXHR){
    // Handler for no items in cart
    if (response == null || response.length <= 0) {
      $('#requests').html('There are no requests')
      $('#loader').hide('slow', function(){ $('#loader').remove(); });
      return
    }
    // Update table
    $(function() {
      if ('content' in document.createElement('template')) {
        $.each(response, function(i, item) {
          var t = document.querySelector('#request_item_template');
          var tc = document.importNode(t.content, true);
          // Setting all requeired elements
          tc.querySelector('tr').id = item.id;
          tc.querySelector('#request_name').textContent = item.template.name;
          tc.querySelector('#request_reason').textContent = item.request_reason;
          tc.querySelector('#request_state').textContent = item.state;
          tc.querySelector('#request_judge_reason').textContent = item.reason;
          tc.querySelector('#request_button_reorder').setAttribute('href', `/request?action=clone&source=requests&template_id=${item.template.id}&request_id=${item.id}`);
          tc.querySelector('#request_button_view').setAttribute('href', `/request?action=view&source=requests&template_id=${item.template.id}&request_id=${item.id}`);
          if (item.state != "pending") {
            tc.querySelector('#request_button_judge').remove();
          } else {
            tc.querySelector('#request_button_judge').removeAttribute('disabled');
            tc.querySelector('#request_button_judge').setAttribute('href', `/request?action=judge&source=requests&template_id=${item.template.id}&request_id=${item.id}`);
          }

          if (item.icon != "") {
            tc.querySelector('#request_icon').setAttribute('src', item.template.icon);
          } else {
            tc.querySelector('#request_icon').setAttribute('src', '/static/logo.png');
          }

          var tb = document.getElementsByTagName("tbody");
          tb[0].appendChild(tc);
        });
        $('#loader').hide('slow', function(){ $('#loader').remove(); });
      } else {
        alert("HTML5 templating does not work with your browser")
      }
    });
  });
}

// CONSTRUCT TABLE
$(document).ready(loadTable());
