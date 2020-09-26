function loadSurvey() {
  // Disable inputs
  $("#template :input").prop("readonly", true);
  // Setup variables
  const params = new URLSearchParams(window.location.search);
  if (params.has('template_id')) {
    template_id = parseInt(params.get('template_id'));
    $('#template_id').val(template_id);
  } else {
    console.log("Query varibale template_id is not defined or malformed");
    return;
  }
  var data = new Object;
  data["id"] = template_id;
  var json_data = JSON.stringify(data);
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
              if(survey.regex != ""){
                variable_input.setAttribute('pattern', survey.regex);
                variable_regex.innerHTML = 'Pattern: ' + survey.regex;
                variable_regex_div.style.display = '';
              }
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
              if(survey.regex != ""){
                variable_input.setAttribute('pattern', survey.regex);
                variable_regex.innerHTML = 'Pattern: ' + survey.regex;
                variable_regex_div.style.display = '';
              }
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
              if(survey.regex != ""){
                variable_input.setAttribute('pattern', survey.regex);
                variable_regex.innerHTML = 'Pattern: ' + survey.regex;
                variable_regex_div.style.display = '';
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
      switchSourceAction();
    }
  });
  $('#request_reason').prop('readonly', false);
  $('#loader').hide('slow', function(){ $('#loader').remove(); });
};

function loadData(source){
  // Load data into inputs
  // Get the request ID
  const params = new URLSearchParams(window.location.search);
  if (params.has('request_id')) {
    request_id = params.get('request_id');
    $('#request_id').val(request_id);
  } else {
    return;
  }

  // Get the user cart
  request = $.ajax({
    url: `/api/v1/${source}/list`,
    source: "GET",
  });

  // Format data based on source
  var data
  request.done(function (response, textStatus, jqXHR){
    if (response == null || response.length <= 0) {
      console.log("No requests found");
      return
    }
    switch (source) {
      case "cart":
        data = response.requests;
        break;
      case "requests":
        data = response;
        break;
      default:
        break;
    }
    $.each(data, function(i, response_request) {
      if( response_request.id == request_id ) {
        // Set the request reason
        $('#request_reason').val(response_request.request_reason);
        $('#reason').val(response_request.reason);
        $('#state').val(response_request.state);

        $('#user_id').val(response_request.user_id);
        $('#judge_id').val(response_request.judge_id);
        $('#last_message').val(response_request.last_message);
        $('#updated_at').val(response_request.updated_at);

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
  });
}

function switchSourceAction(){
  // Get the action and source
  const params = new URLSearchParams(window.location.search);
  var action = 'none';
  if (params.has('action')) {
    action = params.get('action');
  }
  var source = 'none';
  if (params.has('source')) {
    source = params.get('source');
  }
  switch (true) {
    // shop
    case (source == "shop" && action == "create"):
      $('h1').text('Create request');
      $('#request_submit')
        .text('Add request to cart')
        .attr('onclick', 'requestCreate()');
      break;

    // cart
    case (source == "cart" && action == "view"):
      $('h1').text('View request');
      loadData(source);
      $('#template :input').prop("disabled", true);
      $('#request_submit')
        .text('Back to cart')
        .prop('type', 'button')
        .attr('onclick', 'window.location.href="/cart"');
      break;
    case (source == "cart" && action == "edit"):
      $('h1').text('Edit request');
      loadData(source);
      $('#request_submit')
        .text('Update')
        .attr('onclick', 'requestEdit()');
      $('#request_id').prop('name', 'request_id');
      break;
    case (source == "cart" && action == "clone"):
      $('h1').text('Clone request');
      loadData(source);
      $('#request_submit')
        .text('Add clone to cart')
        .attr('onclick', 'requestClone()');
      break;

    // reuqests
    case (source == "requests" && action == "view"):
      $('h1').text('View request');
      loadData(source);
      $('#template :input').prop("disabled", true);
      $('#request_submit')
        .text('Back to requests')
        .prop('type', 'button')
        .attr('onclick', 'window.location.href="/requests"');
      break;
    case (source == "requests" && action == "clone"):
      $('h1').text('Clone request');
      loadData(source);
      $('#request_submit')
        .text('Add clone to cart')
        .prop('type', 'button')
        .attr('onclick', 'requestClone()');
      break;
    case (source == "requests" && action == "judge"):
      // check if user is allowed to judge
      $('h1').text('Judge request');
      loadData(source);
      $('#template :input').prop("disabled", true);
      $('#judge_actions').show();
      $('#reason')
        .attr('readonly', false)
        .attr('disabled', false)
        .attr('required', true);
      $('#button_approve').attr('disabled', false);
      $('#button_deny').attr('disabled', false);
      $('#request_submit')
        .text('Back to requests')
        .prop('type', 'button')
        .attr('onclick', 'window.location.href="/requests"');
      break;
    default:
      break;
  }
}

function requestCreate(){
  event.preventDefault();

  var serializedData = $('#template').serialize();
  $("#template :input").prop("readonly", true);

  request = $.ajax({
    url: "/api/v1/cart/add",
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

function requestEdit(){
  event.preventDefault();

  var serializedData = $('#template').serialize();
  $("#template :input").prop("readonly", true);

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

function requestClone(){
  event.preventDefault();

  var serializedData = $('#template').serialize();
  $("#template :input").prop("readonly", true);

  request = $.ajax({
    url: "/api/v1/cart/add",
    type: "POST",
    data: serializedData
  });

  request.done(function (response, textStatus, jqXHR){
    console.log("Clone successful");
    window.location.href = "/cart";
  });

  request.fail(function (jqXHR, textStatus, errorThrown){
    console.error(
      "The following error occurred: "+
      textStatus, errorThrown
    );
  });
};

function requestApprove(){
  event.preventDefault();

  $('#button_approve').prop('disabled', true);
  $('#button_deny').prop('disabled', true);
  $("#template :input").prop("readonly", true);

  var data = new Object;
  data['id'] = $('#request_id').val();
  data['reason'] = $('#reason').val();
  var json_data = JSON.stringify(data);

  request = $.ajax({
    url: "/api/v1/requests/approve",
    type: "POST",
    contentType: "application/json; charset=utf-8",
    data: json_data
  });

  request.done(function (response, textStatus, jqXHR){
    console.log("Judging successful");
    window.location.href = "/requests";
  });

  request.fail(function (jqXHR, textStatus, errorThrown){
    console.error(
      "The following error occurred: "+
      textStatus, errorThrown
    );
  });
};

function requestDeny(){
  event.preventDefault();

  $('#button_approve').prop('disabled', true);
  $('#button_deny').prop('disabled', true);
  $("#template :input").prop("readonly", true);

  var data = new Object;
  data['id'] = $('#request_id').val();
  data['reason'] = $('#reason').val();
  var json_data = JSON.stringify(data);

  request = $.ajax({
    url: "/api/v1/requests/deny",
    type: "POST",
    contentType: "application/json; charset=utf-8",
    data: json_data
  });

  request.done(function (response, textStatus, jqXHR){
    console.log("Judging successful");
    window.location.href = "/requests";
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
