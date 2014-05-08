$(document).ready(function(){
    // make code pretty
    window.prettyPrint && prettyPrint();

    $("[data-toggle=popover]").popover();

    $('.wmd-input').atwho({
        at: "@",
        data: 'http://www.golangtc.com/users.json'
    });
});
