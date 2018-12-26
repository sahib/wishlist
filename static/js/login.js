function isEmail(email) {
    var regex = /^([a-zA-Z0-9_.+-])+\@(([a-zA-Z0-9-])+\.)+([a-zA-Z0-9]{2,4})+$/;
    return regex.test(email);
}

function pollServerForLogin(delay) {
    // try to list items until we succeed (i.e. we logged in)
    window.setTimeout(function() {
        $.ajax({
            timeout: 60000,
            url: "/api/v0/list",
            type: "GET",
            dataType: "json",
            success: function(result) {
                window.location.replace("/list.html");
            },
            error: function(data, e, m) {
                pollServerForLogin(1000);
            }});
    }, delay);
}

$(document).ready(function(){
    $(".alert-closer").click(function() {
        $(this).parent().toggle()
    })

    var afterSubmitAlert = $("#after-submit-alert")
    afterSubmitAlert.hide();

    var errorBox = $("#error-box")
    errorBox.hide();

    $("#btn-login").on('click', function(){
        var name = $("#inputLoginName").val()
        var email = $("#inputLoginEmail").val()

        if(name.length <= 2) {
            errorBox.text("Bitte gib einen Namen mit mindestens 3 Zeichen ein.")
            errorBox.show()
            return
        }

        if(!isEmail(email)) {
            errorBox.text('"' + email + '" is keine valide E-Mail Adresse.')
            errorBox.show()
            return
        }

        errorBox.hide()

        $.ajax({
            url: "/api/v0/login",
            dataType: "json",
            type: 'POST',
            contentType: 'application/json',
            processData: false,
            data: JSON.stringify({
                "name": name,
                "email": email,
            }),
            success: function(data, textStatus, jQxhr) {
                afterSubmitAlert.show()
                console.log("DATA " + data, data.Success, data.IsAlreadyLoggedIn)
                if(data.IsAlreadyLoggedIn) {
                    afterSubmitAlert.text('Bereits eingeloggt. Leite weiter.');
                    window.location.replace("/list.html")
                } else {
                    afterSubmitAlert.text(
                        'Es wurde eine E-Mail an "' + email + '" geschickt. Bitte klicke auf den darin enthaltenen Link.'
                    );
                    pollServerForLogin(1000)
                }
            },
        });
    });
});
