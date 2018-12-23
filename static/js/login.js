function isEmail(email) {
    var regex = /^([a-zA-Z0-9_.+-])+\@(([a-zA-Z0-9-])+\.)+([a-zA-Z0-9]{2,4})+$/;
    return regex.test(email);
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

        $.ajax("/api/v0/login", {
            data : JSON.stringify({
                "name": name,
                "email": email,
            }),
            contentType : 'application/json',
            type : 'POST',
        });

		afterSubmitAlert.text('Es wurde eine E-Mail an "' + email + '" geschickt. Bitte klicke auf den darin enthaltenen Link.')
		afterSubmitAlert.show()
    });
});