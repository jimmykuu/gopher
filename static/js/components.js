class MarkdownEditor extends React.Component {
  constructor(props) {
    super(props);

    this.state = {
      isEditing: true,
      markdown: '',
      html: ''
    };
  }

  handleMarkdownChange = (e) => {
    let markdown = e.target.value;
    this.setState({
      markdown: markdown
    });

    this.props.onChange(markdown);
  }

  onEditingTabClick = () => {
    this.setState({isEditing: true});
  }

  onPreviewTabClick = () => {
    this.setState({
      isEditing: false,
      html: converter.makeHtml(this.state.markdown)
    });
  }

  render() {
    let editingTabClassName = "item";
    let previewTabClassName = "item";

    if (this.state.isEditing) {
      editingTabClassName += " active";
    } else {
      previewTabClassName += " active";
    }

    return (<null>
      <div className="ui top attached tabular menu">
        <a className={editingTabClassName} onClick={this.onEditingTabClick}>
          编辑
        </a>
        <a className={previewTabClassName} onClick={this.onPreviewTabClick}>
          预览
        </a>
      </div>
      {
        this.state.isEditing ?
          <div className="field">
            <textarea value={this.state.markdown} onChange={this.handleMarkdownChange}></textarea>
          </div>
        :
          <div className="ui container" dangerouslySetInnerHTML={ {__html: this.state.html} }>
          </div>
      }
    </null>);
  }
}
