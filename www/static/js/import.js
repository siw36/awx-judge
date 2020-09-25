function loadTable() {
  $('#loader').show('slow');
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
          var t = document.querySelector('#import_item_template');
          var tc = document.importNode(t.content, true);
          // Setting all requeired elements
          tc.querySelector('#import_id').textContent = item.id;
          tc.querySelector('#import_name').textContent = item.name;
          tc.querySelector('#import_description').textContent = item.description;
          tc.querySelector('#import_button').setAttribute('href', `/template-import-form?template_id=${item.id}`);

          if (templatesImported != undefined) {
            if (templatesImported.filter(function(e) { return e.id == item.id; }).length > 0) {
              tc.querySelector('#import_button_delete').disabled = false;
              tc.querySelector('#import_button_delete').hidden = false;
              tc.querySelector('#import_button_delete').setAttribute('onclick', `templateDelete('${item.id}')`);
              tc.querySelector('#import_button').className = "btn btn-warning";
              tc.querySelector('#import_button').setAttribute('title', "Re-import");
              tc.querySelector('#import_button_icon').className = "fas fa-redo-alt";
            } else {
              tc.querySelector('#import_button_delete').disabled = true;
              tc.querySelector('#import_button_delete').hidden = true;
              tc.querySelector('#import_button').className = "btn btn-primary";
              tc.querySelector('#import_button').setAttribute('title', "Import");
              tc.querySelector('#import_button_icon').className = "fas fa-file-import";
            }
          }

          var tb = document.getElementsByTagName('tbody');
          tb[0].appendChild(tc);
        });
      } else {
        alert("HTML5 templating does not work with your browser")
      }
    });
    $('#loader').hide('slow');
  });
};

function templateDelete(template_id){
  $('#loader').show('slow');
  // Setup variables
  var data = new Object;
  data["id"] = parseInt(template_id);
  var json_data = JSON.stringify(data);

  request = $.ajax({
    url: "/api/v1/templates/remove",
    type: "POST",
    contentType: "application/json; charset=utf-8",
    data: json_data
  });

  request.done(function (response, textStatus, jqXHR){
    // Log a message to the console
    console.log("Delete job template successful");
    // Update table
    loadTable();
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
    $('#loader').hide('slow');
  });

};

// CONSTRUCT TABLE
$(document).ready(loadTable());
