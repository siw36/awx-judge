// LOAD AVAILIABLE ITEMS
var request;
function loadShopItems() {
  request = $.ajax({
    url: '/api/v1/templates/list',
    type: "GET",
  });

  // Callback handler that will be called on success
  request.done(function (response, textStatus, jqXHR){
    // Handler for no items in cart
    if (response == null || response.length <= 0) {
      $("#shop").html("<p>No services available. If you are an admin, you can import some in the import tab.</p>");
      $('#loader').hide('slow', function(){ $('#loader').remove(); });
      return
    }
    console.log("Loading shop items")
    // Update table
    $(function() {
      if ('content' in document.createElement('template')) {
        $.each(response, function(i, item) {
          var t = document.querySelector('#shop_card_template');
          var tc = document.importNode(t.content, true);
          // Setting all requeired elements
          tc.querySelector('#card_name').textContent = item.name;
          tc.querySelector('#card_description').textContent = item.description;
          tc.querySelector('#template_id').value = item.id;

          if (item.icon != "") {
            tc.querySelector('#card_icon').setAttribute('src', item.icon);
          } else {
            tc.querySelector('#card_icon').setAttribute('src', '/static/logo.png');
          }

          var shop = $('#shop')
          shop[0].appendChild(tc);
        });
      } else {
        alert("HTML5 templating does not work with your browser");
      }
    });
    $('#loader').hide('slow', function(){ $('#loader').remove(); });
  });
}

// CONSTRUCT TABLE
$(document).ready(loadShopItems());
