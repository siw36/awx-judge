function loadSurvey() {
  // Disable inputs
  $("#template :input").prop("readonly", true);
  // Setup variables
  var id = parseInt($('#template_id').val());
  var data = new Object;
  data["id"] = id
  var json_data = JSON.stringify(data)
  $.when(
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
    if (templateImported == undefined) {
      console.log("Template not found in DB");
      // Display error message
      return;
    } else {
      if ('content' in document.createElement('template')) {
        // Setting fixed vars
        $('#name').val(templateImported.name);
        $('#description').val(templateImported.description);
        $('#icon').attr('src', templateImported.icon);
        // For each variable select the right template and start to populate it
        $.each(templateImported.spec, function(i, survey) {
          var t = document.querySelector('#' + survey.type);
          var tc = document.importNode(t.content, true);
          // Setting variables that are present in every template
          tc.querySelector('#question_name').textContent = survey.question_name;
          tc.querySelector('#question_description').textContent = survey.question_description;

          // Setting variables that are individual to a template
          if( survey.type == "multiplechoice" ) {
            choices = survey.choices.split("\n");
            variable_choices = tc.querySelector('#choices');
            variable_choices.setAttribute('name', survey.variable);
            $.each(choices, function(i, choice) {
              var option = document.createElement("option");
              option.text = choice;
              option.value = choice;
              if(choice == survey.default){
                option.setAttribute('selected', true);
              }
              variable_choices.appendChild(option);
            })
          }
          if( survey.type == "multiselect" ) {
            choices = survey.choices.split("\n");
            variable_choices = tc.querySelector('#choices');
            variable_choices.setAttribute('name', survey.variable);
            $.each(choices, function(i, choice) {
              var wrapper = document.createElement("div");
              wrapper.setAttribute('class', 'form-check form-check-inline');
              var input = document.createElement("input");
              input.setAttribute('class', 'form-check-input');
              input.setAttribute('type', 'checkbox');
              input.setAttribute('id', choice);
              input.setAttribute('name', survey.variable);
              input.value = choice;
              var label = document.createElement("label");
              label.setAttribute('class', 'form-check-label');
              label.setAttribute('for', choice);
              label.innerHTML = choice;
              wrapper.appendChild(input);
              wrapper.appendChild(label);
              variable_choices.appendChild(wrapper);
            })
          }
          if( survey.type == "text" ) {
            variable_input = tc.querySelector("input");
            variable_regex_div = tc.querySelector(".input-group-append");
            variable_regex = tc.querySelector("span");

            variable_input.setAttribute('name', survey.variable);
            variable_input.required = survey.required;
            if(survey.default != ""){
              variable_input.value = survey.default;
            }
            variable_input.setAttribute('pattern', survey.regex);
            variable_regex.innerHTML = 'Pattern: ' + survey.regex;
          }

          var survey = document.querySelector('#survey_parameters');
          survey.appendChild(tc);
        });
      } else {
        alert("HTML5 templating does not work with your browser")
      }
      $('#request_reason').prop('readonly', false);
      $('#loader').hide('slow', function(){ $('#loader').remove(); });
    }
  });
};

// function loadData(){
//
//   // Get the request ID
//   var id = parseInt($('#request_id').val());
//
//   // Get the user cart
//   request = $.ajax({
//     url: '/api/v1/cart/list',
//     type: "GET",
//   });
//
//   // Callback handler that will be called on success
//   request.done(function (response, textStatus, jqXHR){
//     // Handler for no items in cart
//     if (response.requests.length <= 0) {
//       console.log("User cart is empty");
//       return
//     }
//     $(function() {
//       // Populate data
//       $.each(response.requests, function(i, request) {
//         if( request.id == request_id ){
//           $.each(request.template.spec, function(i, spec)) {
//             // text and number inputs
//             $('input[name ="' + spec + '"]').val(spec.value);
//             // multiplechoice and multiselect
//           };
//           // Break the loop
//           return false;
//         }
//       }
//       } else {
//         console.log("Error getting user cart");
//         return
//       }
//     });
//   });
// };

// CONSTRUCT TABLE
$(document).ready(loadSurvey());

$('#template_icon_link').change(function(){
  $('#preview_image').attr("src", $('#template_icon_link').val());
});
