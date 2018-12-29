function isURL(str) {
  var pattern = new RegExp('^(https?:\\/\\/)?'+ // protocol
    '((([a-z\\d]([a-z\\d-]*[a-z\\d])*)\\.?)+[a-z]{2,}|'+ // domain name
        '((\\d{1,3}\\.){3}\\d{1,3}))'+ // OR ip (v4) address
    '(\\:\\d+)?(\\/[-a-z\\d%_.~+]*)*'+ // port and path
    '(\\?[;&a-z\\d%_.~+=-]*)?'+ // query string
    '(\\#[-a-z\\d_]*)?$','i'); // fragment locator
  return pattern.test(str);
}

function getCookie(cname) {
  var name = cname + "=";
  var decodedCookie = decodeURIComponent(document.cookie);
  var ca = decodedCookie.split(';');
  for(var i = 0; i <ca.length; i++) {
      var c = ca[i];
      while (c.charAt(0) == ' ') {
            c = c.substring(1);
          }
      if (c.indexOf(name) == 0) {
            return c.substring(name.length, c.length);
          }
    }
  return "";
}

function eraseCookie(name) {
    document.cookie = name+'=; Max-Age=-99999999;';
}

function updateList() {
    $('#item-list-body > tr').remove();
    $.get("/api/v0/list", function(data) {
        $.each(data["items"], function(index, item) {
            var btn = null

            if(!item.is_reserved) {
                btn = $('<button class="btn btn-sm btn-outline-success">').text("Reservieren")
                btn.on('click', function(){
                    $.ajax("/api/v0/reserve", {
                        data : JSON.stringify({
                            "item_id": item.id,
                            "do_reserve": true,
                        }),
                        contentType : 'application/json',
                        type : 'POST',
                    });
                })
            } else if(item.is_reserved_by_us) {
                btn = $('<button class="btn btn-sm btn-outline-danger">').text("Reservierung aufheben")
                btn.on('click', function(){
                    $.ajax("/api/v0/reserve", {
                        data : JSON.stringify({
                            "item_id": item.id,
                            "do_reserve": false,
                        }),
                        contentType : 'application/json',
                        type : 'POST',
                    });
                });
            } else if(item.is_reserved && !item.is_reserved_by_us) {
                btn = $('<button class="btn btn-sm btn-outline-dark disabled">').text("Bereits reserviert")
            }

            var closer = $('<a>')
                .attr('href', '#')
                .attr('class', 'close alert-closer')
                .attr('data-dismiss', 'alert')
                .attr('aria-label', 'close')
                .html('&times;')

            closer.on('click', function(){
                    $.ajax("/api/v0/delete", {
                        data : JSON.stringify({
                            "itemid": item.id,
                        }),
                        contentType : 'application/json',
                        type : 'POST',
                    });
            });

            var linkText = item.name
            if(item.link != undefined && item.link == "") {
                linkText = '<a href="'+item.link+'">'+item.name+'</a>'
            }

            var row = $("<tr>")
                .append($('<th scope="row">').text(index + 1))
                .append($('<td>').append(linkText))
                .append($('<td>').append(btn))

            if(item.is_own) {
                row.append($('<td>').append(closer))
            } else {
                row.append($('<td>'))
            }

            $("#item-list-body").append(row)

            var logout = $('<a href="#">Logout</a>')
            logout.on('click', function() {
                $.ajax("/api/v0/logout", {
                    type : 'GET',
                });

                eraseCookie("user_name");
                eraseCookie("user_email");
                eraseCookie("session_id");
                window.location.replace("/login.html")
            })

            $("#logged-in-as").html(
                'Eingeloggt als ' + getCookie('user_name') + ' (' + getCookie('user_email') + ')' +
                ' | <a href="https://github.com/sahib/wishlist">Quelltext dieser Seite</a>' +
                ' | '
            ).append(logout);
        });
    }).fail(function() {
        $("#alert-add-item").show()
        $("#alert-add-item-span").html("Es scheint du bist nicht eingeloggt. Möglicherweise ist deine Sitzung abgelaufen. Du wirst in 5 Sekunden auf die <a href=\"/login.html\">Login-Seite umgeleitet.</a>");
        $("#login-welcome").hide();
        setTimeout(function() { window.location.replace("/login.html"); }, 5000);

    })
}

function pollServer(delay) {
    window.setTimeout(function() {
        $.ajax({
            timeout: 60000,
            url: "/api/v0/events?timeout=60&category=list-change",
            type: "GET",
            dataType: "json",
            success: function(result) {
                updateList();
                pollServer(100);
            },
            error: function(data, e, m) {
                pollServer(1000);
            }});
    }, delay);
}

$(document).ready(function(){
    // Hide all alerts by default:
    $(".alert").hide()
    $(".alert-info").show()
    $(".alert-closer").click(function() {
        $(this).parent().toggle()
    })

    updateList();
    pollServer(0);

    $("#btn-add-item").on('click', function(){
        var name = $("#inputItemName").val()
        var link = $("#inputItemLink").val()
        if(name.length < 1) {
            $("#alert-add-item").show()
            $("#alert-add-item-span").text("Bitte gib ein Geschenknamen ein.");
            return
        }

        if(link != "") {
            if(!isURL(link)) {
                $("#alert-add-item").show()
                $("#alert-add-item-span").text("Bitte gib eine valide URL ein.");
                return
            }
        }

        $.ajax("/api/v0/add", {
            data : JSON.stringify({
                "name": name,
                "link": link,
            }),
            contentType : 'application/json',
            type : 'POST',
        });
    });
});
