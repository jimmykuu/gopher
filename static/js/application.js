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

function createEditorMd(divId, submitId, markdown) {
    var editor = editormd(divId, {
        height: 400,
		markdown: markdown,
	    autoFocus: false,
        path: "http://gopher.qiniudn.com/static/lib/editor.md-1.5.0/lib/",
	    placeholder: "Mardkown，提交前请查看预览格式是否正确",
        toolbarIcons: function() {
          return ["undo", "redo", "|", "bold", "italic", "quote", "|", "h1", "h2", "h3", "h4", "h5", "h6", "|", "list-ul", "list-ol", "hr", "|", "link", "reference-link", "image", "code", "preformatted-text", "code-block", "|", "goto-line", "watch", "preview", "fullscreen", "|", "help", "info"]
        },
        saveHTMLToTextarea: true,
        imageUpload: true,
        imageFormats: ["jpg", "jpeg", "gif", "png"],
        imageUploadURL: "/upload/image",
	    onchange: function() {
	      $(submitId).attr('disabled', this.getMarkdown().trim() == "");
	    }
      });

	return editor;
}

$(document).ready(function(){
	editormd.urls.atLinkBase = "/member/";

    $("[data-toggle=popover]").popover();

    $('.wmd-input').atwho({
        at: "@",
        data: 'http://www.golangtc.com/users.json'
    });

    setToTop();
    
    $('pre').addClass("prettyprint linenums");
    prettyPrint();
});
