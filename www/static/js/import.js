function loadTable() {
  $.when(
    $.ajax({
      url: '/api/v1/import/list',
      type: "GET",
      contentType: "application/json; charset=utf-8",
      success: function(response){
        templatesAvialable = response;
      }
    }),
    $.ajax({
      url: '/api/v1/templates/list',
      type: "GET",
      contentType: "application/json; charset=utf-8",
      success: function(response){
        templatesImported = response;
      }
    })
  ).then(function() {
    // Handler for no items in cart
    if (templatesAvialable == undefined) {
      console.log("No templates found");
      // Display error
      return;
    }
    // Update table
    $(function() {
      var tbody = document.getElementsByTagName("tbody")[0].innerHTML = "";
      if ('content' in document.createElement('template')) {
        $.each(templatesAvialable, function(i, item) {
          var t = document.querySelector('#import_item_template'),
          // Setting all requeired elements
          import_id = t.content.querySelector('#import_id');
          import_name = t.content.querySelector('#import_name');
          import_description = t.content.querySelector('#import_description');
          import_button = t.content.querySelector('#import_button');
          import_button_delete = t.content.querySelector('#import_button_delete');
          import_form = t.content.querySelector('form');
          import_form_id = t.content.querySelector('#import_form_id');
          import_button_icon = t.content.querySelector('#import_button_icon')

          if (templatesImported != undefined) {
            if (templatesImported.filter(function(e) { return e.id == item.id; }).length > 0) {
              import_button_delete.disabled = false;
              import_button_delete.hidden = false;
              import_button_delete.setAttribute ('onclick', 'deleteItem(this)');
              import_button.className = "btn btn-warning";
              import_button.setAttribute('title', "Re-import");
              import_button_icon.className = "fas fa-redo-alt";
            } else {
              import_button_delete.disabled = true;
              import_button_delete.hidden = true;
              import_button.className = "btn btn-primary";
              import_button.setAttribute('title', "Import");
              import_button_icon.className = "fas fa-file-import";
            }
          }
          import_id.textContent = item.id;
          import_name.textContent = item.name;
          import_description.textContent = item.description;
          import_button.setAttribute('form', item.id);
          import_button_delete.setAttribute('data-template_id', item.id);

          import_form.id = item.id;
          import_form_id.setAttribute('value', item.id);

          var tb = document.getElementsByTagName('tbody');
          var clone = document.importNode(t.content, true);
          tb[0].appendChild(clone);
        });
      } else {
        alert("HTML5 templating does not work with your browser")
      }
    });
    $('#loader').hide('slow', function(){ $('#loader').remove(); });
  });
};

var request;
function deleteItem(obj){
  // Abort any pending request
  if (request) {
    request.abort();
  }
  // Setup variables
  var data = new Object;
  data["id"] = $(obj).data("template_id");
  var json_data = JSON.stringify(data);

  $(obj).disabled = true;
  // Fire off the request
  request = $.ajax({
    url: "/api/v1/templates/remove",
    type: "POST",
    contentType: "application/json; charset=utf-8",
    data: json_data
  });

  // Callback handler that will be called on success
  request.done(function (response, textStatus, jqXHR){
    // Log a message to the console
    console.log("Delete job template successful");
    // Update table
    loadTable();
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

// CONSTRUCT TABLE
$(document).ready(loadTable());
