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
        path: "http://77fkk5.com1.z0.glb.clouddn.com/static/lib/editor.md-1.5.0/lib/",
	    placeholder: "Markdown，提交前请查看预览格式是否正确",
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

function aliyunA(data) {
    var banner = eval(data['data'][0]['content'])[0];
    $('#aliyun-a').replaceWith('<a href="' + banner['url'] + '" target="_blank"><img src="' + banner['img'] + '" style="max-width: 100%;"></a>');
}

function aliyunC(data) {
    var banner = eval(data['data'][0]['content'])[0];
    $('#aliyun-c').replaceWith('<a href="' + banner['url'] + '" target="_blank"><img src="' + banner['img'] + '" width="100%"></a>');
}

function aliyunD(data) {
    var banner = eval(data['data'][0]['content'])[0];
    $('#aliyun-d').replaceWith('<a href="' + banner['url'] + '" target="_blank"><img src="' + banner['img'] + '"></a>');
}

$(document).ready(function(){
	editormd.urls.atLinkBase = "/member/";

    $("[data-toggle=popover]").popover();

    setToTop();

    $('.editormd-preview-container pre').addClass("prettyprint linenums");
    prettyPrint();

    $('.content .body a').attr('target', '_blank');
    $('.reply-content a ').attr('target', '_blank');
});
