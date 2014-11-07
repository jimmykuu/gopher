function setToTop() {
    $('body').append('<div id="toTop" title="回到顶部"><span class="glyphicon glyphicon-circle-arrow-up"></span></div>');
    $(window).scroll(function() {
        if ($(this).scrollTop()) {
            $('#toTop').fadeIn();
        } else {
            $('#toTop').fadeOut();
        }
    });

    $("#toTop").click(function () {
        //html works for FFX but not Chrome
        //body works for Chrome but not FFX
        //This strange selector seems to work universally
        $("html, body").animate({scrollTop: 0}, 200);
    });
}

$(document).ready(function(){
	$(".dropdown-toggle").on("mouseover", function() {
		if ($(this).parent().is(".open")) {
			return;
		}

		$(this).dropdown("toggle")
	});
	
    // make code pretty
    window.prettyPrint && prettyPrint();

    $("[data-toggle=popover]").popover();

    $('.wmd-input').atwho({
        at: "@",
        data: 'http://www.golangtc.com/users.json'
    });

    setToTop();
});
