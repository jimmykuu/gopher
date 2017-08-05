const converter = new showdown.Converter({
  simpleLineBreaks: true
});

class MarkdownEditor extends React.Component {
  constructor(props) {
    super(props);

    this.state = {
      isEditing: true,
      html: ''
    };
  }

  handleMarkdownChange = (e) => {
    let markdown = e.target.value;

    this.props.onChange(markdown);
  }

  onEditingTabClick = () => {
    this.setState({isEditing: true});
  }

  onPreviewTabClick = () => {
    this.setState({
      isEditing: false,
      html: converter.makeHtml(this.props.markdown)
    });
  }

  onUploadImage = () => {
    this.refs.file.click();
  }

  onChooesFile = (e) => {
    this.onUpload(e.target.files[0]);
  }

  onUpload = (file) => {
    let {markdown} = this.props;

    let formData = new FormData();
    formData.append('image', file);
    let xhr = new XMLHttpRequest();
    xhr.open("POST", "/api/upload/image", true);
    xhr.setRequestHeader("Authorization", 'Bearer ' + window.localStorage.getItem('token'));

    xhr.addEventListener('load', () => {
      const resp = JSON.parse(xhr.responseText);
      if (resp.status) {
        if (markdown) {
          markdown = markdown + "\n" + "![](" + resp.image_url + ")\n";
        } else {
          markdown = "![](" + resp.image_url + ")\n";
        }

        this.setState({
          showUploadDialog: false
        });

        this.props.onChange(markdown);
      }
    });

    xhr.send(formData);
  }

  render() {
    let editingTabClassName = "item";
    let previewTabClassName = "item";

    if (this.state.isEditing) {
      editingTabClassName += " active";
    } else {
      previewTabClassName += " active";
    }

    const inputFileStyle = {position: 'absolute', left: -1000000};

    return (<null>
      <input type="file" ref="file" style={inputFileStyle} onChange={this.onChooesFile} />
      <div className="ui top attached tabular menu">
        <a className={editingTabClassName} onClick={this.onEditingTabClick}>
          编辑
        </a>
        <a className={previewTabClassName} onClick={this.onPreviewTabClick}>
          预览
        </a>
        <div className="icon right menu">
          <a className="ui dropdown icon item">
            <i className="code icon"></i>
            <div className="menu">
              <div className="item">
                Open...
              </div>
              <div className="item">
                Save...
              </div>
            </div>
          </a>
          <a className="item" onClick={this.onUploadImage}>
            <i className="image icon"></i>
          </a>
        </div>
      </div>
      {
        this.state.isEditing ?
          <div className="field">
            <textarea value={this.props.markdown} onChange={this.handleMarkdownChange}></textarea>
          </div>
        :
          <div className="ui container" dangerouslySetInnerHTML={ {__html: this.state.html} }>
          </div>
      }
    </null>);
  }
}
