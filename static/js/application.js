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
    // make code pretty
    window.prettyPrint && prettyPrint();

    $("[data-toggle=popover]").popover();

    $('.wmd-input').atwho({
        at: "@",
        data: 'http://www.golangtc.com/users.json'
    });

    setToTop();
});
