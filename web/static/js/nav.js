// Heighligt current path
$(function(){
    var current = window.location.pathname;
    $('#nav a').each(function(){
        var $this = $(this);
        // if the current path is like this link, make it active
        if($this.attr('href').indexOf(current) !== -1){
            $this.addClass('active');
        }
    })
})

// Cart: display the amount of items

// Requests: display the amount of pending, approved, denied requests

// Judge: Display the amount of pending requests
