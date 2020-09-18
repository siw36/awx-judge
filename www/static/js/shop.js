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
    if (response == null) {
      $("#shop").html("<p>No services available. If you are an admin, you can import some in the import tab.</p>");
      $('#loader').hide('slow', function(){ $('#loader').remove(); });
      return
    }
    console.log("Loading shop items")
    // Update table
    $(function() {
      if ('content' in document.createElement('template')) {
        $.each(response, function(i, item) {
          var t = document.querySelector('#shop_card_template'),
          // Setting all requeired elements
          card_icon = t.content.querySelector('#card_icon');
          card_name = t.content.querySelector('#card_name');
          card_description = t.content.querySelector('#card_description');
          template_id = t.content.querySelector('#template_id');

          if (item.icon != "") {
            card_icon.setAttribute('src', item.icon);
          } else {
            card_icon.setAttribute('src', '/static/logo.png');
          }
          card_name.textContent = item.name;
          card_description.textContent = item.description;
          template_id.value = item.id;

          var shop = $('#shop')
          var clone = document.importNode(t.content, true);
          shop[0].appendChild(clone);
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
