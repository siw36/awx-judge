function loadSurvey() {
  // Disable inputs
  $("#template :input").prop("readonly", true);
  // Setup variables
  var id = parseInt($('#template_import_form_id').val());
  var data = new Object;
  data["id"] = id
  var json_data = JSON.stringify(data)
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
    if (surveyAvailable.variable == "") {
      console.log("Failed to get template survey from AWX")
    } else {
      if ('content' in document.createElement('template')) {
        $.each(surveyAvailable, function(i, available) {
          var t = document.querySelector('#import_variable_template'),
          // Setting all requeired elements
          variable_name = t.content.querySelector('#name');
          variable_question = t.content.querySelector('#question_name');
          variable_description = t.content.querySelector('#question_description');
          variable_default = t.content.querySelector('#default');
          variable_type = t.content.querySelector('#type');
          variable_choices = t.content.querySelector('#choices');
          variable_required = t.content.querySelector('#required');
          var variable_regex = t.content.querySelector('#regex');

          variable_name.textContent = available.variable;
          variable_question.textContent = available.question_name;
          variable_description.textContent = available.question_description;
          variable_default.textContent = available.default;
          variable_type.textContent = available.type;
          variable_choices.textContent = available.choices;
          variable_required.textContent = available.required;
          variable_regex.setAttribute("name", available.variable);

          if (templateImported.id == "") {
            console.log("Template not yet imported");
            $('#template_name').val(templateAvialable.name);
            $('#template_description').val(templateAvialable.description);
          } else {
            console.log("Template already imported. Using existing data.");
            console.log(templateImported.icon_link);
            if (templateImported.icon_link != "") {
              $('#template_icon_link').val(templateImported.icon_link);
              $('#preview_image').attr("src", templateImported.icon_link);
            }
            $('#template_name').val(templateImported.name);
            $('#template_description').val(templateImported.description);

            // Fill in the regex, if imported and available fields names are the same
            $.each(templateImported.spec, function(index, imported) {
              if (available.variable == imported.variable && available.type == imported.type) {
                variable_regex.setAttribute("value", imported.regex);
              }
            })
          }
          if (available.type == "multiplechoice" || available.type == "multiselect"){
            variable_regex.disabled = true;
          } else {
            variable_regex.disabled = false;
          }

          var tb = document.getElementsByTagName('tbody');
          var clone = document.importNode(t.content, true);
          tb[0].appendChild(clone);
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
var request;
// Bind to the submit event of our form
function importItem(){

  var form = $('#template');

  // Abort any pending request
  if (request) {
    request.abort();
  }
  // Setup variables
  var serialized_data = form.serialize();

  // Fire off the request
  request = $.ajax({
    url: "/api/v1/import/add",
    type: "POST",
    data: serialized_data
  });

  // Callback handler that will be called on success
  request.done(function (response, textStatus, jqXHR){
    // Log a message to the console
    console.log("Import template was successful");
    // Forward client to shop
    window.location.replace("/shop");
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
  // request.always(function () {
  //   // Reenable the inputs
  //   $(obj).disabled = false;
  // });

};

// CONSTRUCT TABLE
$(document).ready(loadSurvey());

$('#template_icon_link').change(function(){
  $('#preview_image').attr("src", $('#template_icon_link').val());
});
