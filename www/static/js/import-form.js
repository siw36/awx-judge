function loadSurvey() {
  // Disable inputs
  $("#template :input").prop("readonly", true);
  // Setup variables
  const params = new URLSearchParams(window.location.search);
  if (params.has('template_id')) {
    var template_id = parseInt(params.get('template_id'));
    $('#template_id').val(template_id);
  } else {
    console.log("Query varibale template_id is not defined or malformed");
    return;
  }
  var data = new Object;
  data["id"] = template_id;
  var json_data = JSON.stringify(data);
  $.when(
    // Template from AWX
    $.ajax({
      url: '/api/v1/import/get',
      type: "POST",
      contentType: "application/json; charset=utf-8",
      data: json_data,
      success: function(response){
        templateAvialable = response;
      }
    }),
    // Survey spec from AWX
    $.ajax({
      url: "/api/v1/import/survey/get",
      type: "POST",
      contentType: "application/json; charset=utf-8",
      data: json_data,
      success: function(response){
        surveyAvailable = response;
      }
    }),
    // Template from DB, also containing survey spec
    $.ajax({
      url: '/api/v1/templates/get',
      type: "POST",
      contentType: "application/json; charset=utf-8",
      data: json_data,
      success: function(response){
        templateImported = response;
      }
    })
  ).then(function() {
    if (templateAvialable == undefined) {
      console.log("Template not found on AWX");
      // Display error message
      return;
    }
    if (surveyAvailable == null) {
      console.log("Failed to get template survey from AWX or template has no survey variables")
      $('#template_import_form_id').val(template_id);
      $('#template_name').val(templateAvialable.name);
      $('#template_description').val(templateAvialable.description);
      // Enabling inputs after data is filled
      $("#template_name").prop("readonly", false);
      $("#template_description").prop("readonly", false);
      $("#template_icon_link").prop("readonly", false);
      $('#survey').hide();
    } else {
      if ('content' in document.createElement('template')) {
        $.each(surveyAvailable, function(i, available) {
          var t = document.querySelector('#import_variable_template');
          var tc = document.importNode(t.content, true);
          // Setting all requeired elements
          tc.querySelector('#name').textContent = available.variable;
          tc.querySelector('#question_name').textContent = available.question_name;
          tc.querySelector('#question_description').textContent = available.question_description;
          tc.querySelector('#default').textContent = available.default;
          tc.querySelector('#type').textContent = available.type;
          tc.querySelector('#choices').textContent = available.choices;
          tc.querySelector('#required').textContent = available.required;
          tc.querySelector('#regex').setAttribute("name", available.variable);

          if (templateImported.id == "") {
            console.log("Template not yet imported");
            $('#template_import_form_id').val(template_id);
            $('#template_name').val(templateAvialable.name);
            $('#template_description').val(templateAvialable.description);
          } else {
            console.log("Template already imported. Using existing data.");
            console.log(templateImported.icon_link);
            if (templateImported.icon_link != "") {
              $('#template_icon_link').val(templateImported.icon_link);
              $('#preview_image').attr("src", templateImported.icon_link);
            }
            $('#template_import_form_id').val(template_id);
            $('#template_name').val(templateImported.name);
            $('#template_description').val(templateImported.description);

            // Fill in the regex, if imported and available fields names are the same
            $.each(templateImported.spec, function(index, imported) {
              if (available.variable == imported.variable && available.type == imported.type) {
                tc.querySelector('#regex').setAttribute("value", imported.regex);
              }
            })
          }
          if (available.type == "multiplechoice" || available.type == "multiselect"){
            tc.querySelector('#regex').disabled = true;
          } else {
            tc.querySelector('#regex').disabled = false;
          }

          var tb = document.getElementsByTagName('tbody');
          tb[0].appendChild(tc);
        });
        // Enabling inputs after data is filled
        $("#template_name").prop("readonly", false);
        $("#template_description").prop("readonly", false);
        $("#template_icon_link").prop("readonly", false);
      } else {
        alert("HTML5 templating does not work with your browser")
      }
    }
      $('#loader').hide('slow', function(){ $('#loader').remove(); });
  });
};

// Import template
function importItem(){
  var form = $('#template');
  if (request) {
    request.abort();
  }
  // Setup variables
  var serialized_data = form.serialize();

  request = $.ajax({
    url: "/api/v1/import/add",
    type: "POST",
    data: serialized_data
  });

  request.done(function (response, textStatus, jqXHR){
    // Log a message to the console
    console.log("Import template was successful");
    // Forward client to shop
    window.location.replace("/shop");
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

};

// CONSTRUCT TABLE
$(document).ready(loadSurvey());

$('#template_icon_link').change(function(){
  $('#preview_image').attr("src", $('#template_icon_link').val());
});
