$(document).ready(function(){
    // make code pretty
    window.prettyPrint && prettyPrint();

    $("[data-toggle=popover]").popover();

    $('.wmd-input').atwho({
        at: "@",
        data: 'http://localhost:8888/users.json'
    });
});
