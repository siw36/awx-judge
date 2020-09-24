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
          switch (survey.type) {
            case "multiplechoice":
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
              break;
            case "multiselect":
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
                input.value = choice + '\\n';
                var label = document.createElement("label");
                label.setAttribute('class', 'form-check-label');
                label.setAttribute('for', choice);
                label.innerHTML = choice;
                wrapper.appendChild(input);
                wrapper.appendChild(label);
                variable_choices.appendChild(wrapper);
              })
              break;
            case "text":
            case "password":
              variable_input = tc.querySelector("input");
              variable_regex_div = tc.querySelector(".input-group-append");
              variable_regex = tc.querySelector("span");

              variable_input.setAttribute('name', survey.variable);
              if(survey.required == true){
                variable_input.required = true;
              }
              if(survey.default != ""){
                variable_input.value = survey.default;
              }
              variable_input.setAttribute('pattern', survey.regex);
              variable_regex.innerHTML = 'Pattern: ' + survey.regex;
              break;
            case "textarea":
              variable_input = tc.querySelector("textarea");
              variable_regex_div = tc.querySelector(".input-group-append");
              variable_regex = tc.querySelector("span");

              variable_input.setAttribute('name', survey.variable);
              if(survey.required == true){
                variable_input.required = true;
              }
              if(survey.default != ""){
                variable_input.value = survey.default;
              }
              variable_input.setAttribute('pattern', survey.regex);
              variable_regex.innerHTML = 'Pattern: ' + survey.regex;
              break;
            case "integer":
            case "float":
              variable_input = tc.querySelector("input");
              variable_regex_div = tc.querySelector(".input-group-append");

              variable_input.setAttribute('name', survey.variable);
              if(survey.required == true){
                variable_input.required = true;
              } else {
                variable_input.required = false;
              }
              if(survey.default != ""){
                variable_input.value = survey.default;
              }
              variable_input.setAttribute('min', survey.min);
              variable_input.setAttribute('max', survey.max);
              break;
          }

          var survey = document.querySelector('#survey_parameters');
          survey.appendChild(tc);
        });
      } else {
        alert("HTML5 templating does not work with your browser")
      }
      // Load data into inputs
      // Get the request ID
      var request_id = $('#request_id').val();
      if (request_id == "00000000-0000-0000-0000-000000000000") {
        return
      }
      // Get the action type
      const params = new URLSearchParams(window.location.search);
      var action = 'none';
      if (params.has('action')) {
        action = params.get('action');
      }

      // Get the user cart
      request = $.ajax({
        url: '/api/v1/cart/list',
        type: "GET",
      });

      // Callback handler that will be called on success
      request.done(function (response, textStatus, jqXHR){
        if (response.requests.length <= 0) {
          console.log("User cart is empty");
          return
        }
        $.each(response.requests, function(i, response_request) {
          if( response_request.id == request_id ) {
            // Set the request reason
            $('#request_reason').val(response_request.request_reason);
            $.each(response_request.template.spec, function(a, spec) {
              switch (spec.type) {
                case "textarea":
                  $(`textarea[name="${spec.variable}"]`).val(spec.value);
                  break;
                case "multiplechoice":
                  $(`option[value="${spec.value}"]`).prop('selected', true);
                  break;
                case "multiselect":
                  choices = spec.value.split("\n");
                  $.each(choices, function(i, choice) {
                    if(choice != ""){
                      $(`input[name=${spec.variable}][value*=${choice}]`).prop('checked', true);
                    }
                  })
                  break;
                default:
                  $(`input[name="${spec.variable}"]`).val(spec.value);
                  break;
              }
            })
            return false;
          }
        })
        switch (action) {
          case "view":
            $('h1').text('View request');
            $('#template :input').prop("disabled", true);
            $('#request_submit').text('Back to cart');
            $('#request_submit').prop('type', 'button');
            $('#request_submit').attr('onclick', 'window.location.href="/cart"');
            break;
          case "edit":
            $('h1').text('Edit request');
            $('#request_submit').text('Update');
            $('#request_submit').attr('onclick', 'edit()');
            $('#request_id').prop('name', 'request_id');
            break;
          case "clone":
            $('h1').text('Clone request');
            $('#request_submit').text('Add clone to cart');
            $('#template').prop('action', '/api/v1/cart/add');
            break;
          default:
            break;
        }
      });
    }
  });
  $('#request_reason').prop('readonly', false);
  $('#loader').hide('slow', function(){ $('#loader').remove(); });
};

function edit(){
  event.preventDefault();

  var serializedData = $('#template').serialize();
  $("#template :input").prop("readonly", true);

  console.log(serializedData);

  request = $.ajax({
    url: "/api/v1/cart/edit",
    type: "POST",
    data: serializedData
  });

  request.done(function (response, textStatus, jqXHR){
    console.log("Edit successful");
    window.location.href = "/cart";
  });

  request.fail(function (jqXHR, textStatus, errorThrown){
    console.error(
      "The following error occurred: "+
      textStatus, errorThrown
    );
  });
};

// CONSTRUCT TABLE
$(document).ready(function() {
  loadSurvey();
});

$('#template_icon_link').change(function(){
  $('#preview_image').attr("src", $('#template_icon_link').val());
});
