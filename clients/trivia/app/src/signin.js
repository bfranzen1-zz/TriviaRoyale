/**
 *********************************
TODO:

    1. Send post request to users
        a. On success, store auth key and possibly user struct in local storage, redirect to lobbies
        b. On fail, show alert with error message
    2. Send get request to users (or whatever validation is)
        a. On success, store auth key and possibly user struct in local storage, redirect to lobbies
        b. On fail, show alert with error message

 *********************************
 */

const url = "placeholder.com"


$('.form').find('input, textarea').on('keyup blur focus', function (e) {

    var $this = $(this),
        label = $this.prev('label');

    if (e.type === 'keyup') {
        if ($this.val() === '') {
            label.removeClass('active highlight');
        } else {
            label.addClass('active highlight');
        }
    } else if (e.type === 'blur') {
        if ($this.val() === '') {
            label.removeClass('active highlight');
        } else {
            label.removeClass('highlight');
        }
    } else if (e.type === 'focus') {

        if ($this.val() === '') {
            label.removeClass('highlight');
        }
        else if ($this.val() !== '') {
            label.addClass('highlight');
        }
    }

});

$('.tab a').on('click', function (e) {
    e.preventDefault();

    $(this).parent().addClass('active');
    $(this).parent().siblings().removeClass('active');

    target = $(this).attr('href');

    $('.tab-content > div').not(target).hide();

    $(target).fadeIn(600);

});

// Logic to send new user or returning user data to server

$('#new-user-form').submit(function(e) {
    e.preventDefault();
    var formInputs = $('#new-user-form :input');

    var values = {};
    formInputs.each(function() {
        values[this.name] = $(this).val();
    });
    var valJson = JSON.stringify(values);

    $.ajax({
        type: "POST",
        url: url,
        contentType: 'application/json',
        data: valJson,
        success: function( data, textStatus, response) {
            var auth = response.getResponseHeader('Authorization');
            localStorage.setItem('auth', auth)
            window.location.replace("../public/lobby.html");
        },
        error: function(jqXhr, textStatus, errorThrown) {
            alert(errorThrown);
        }
    })

});

$('#user-form').submit(function(e) {
    e.preventDefault();
    window.location.replace("../public/lobby.html");
});

