class MarkdownEditor extends React.Component {
  constructor(props) {
    super(props);

    this.state = {
      editor: null
    };
  }

  getMarkdown() {
    return "markdown";
  }

  onChange = (event) => {
    const { editor } = this.state;

    this.props.onChange(editor.getMarkdown(), editor.getHtml());
  }

  componentDidMount() {
    let that = this;
    let editor = new tui.Editor({
      el: document.querySelector('#editSection'),
      initialEditType: 'markdown',
      previewStyle: 'vertical',
      height: '300px',
      language: 'zh_CN',
      hooks: {
        addImageBlobHook: function(file, callback, source) {
          console.log('addImageBlobHook');
          console.log(file);

          let formData = new FormData();
          formData.append('image', file);
          let xhr = new XMLHttpRequest();
          xhr.open("POST", "/api/upload/image", true);
          xhr.setRequestHeader("Authorization", 'Bearer ' + window.localStorage.getItem('token'));

          xhr.addEventListener('load', () => {
            const resp = JSON.parse(xhr.responseText);
            if (resp.status) {
              callback(resp.image_url, '');
            }
          });

          xhr.send(formData);
          return false;
        }
      },
      events: {
        change: this.onChange
      }
    });

    this.setState({editor});
  }

  render() {
    return <div id="editSection"></div>;
  }
}

class Toolbar extends React.Component {
  constructor(props) {
    super(props);

    this.state = {
      'canCollect': this.props.canCollect
    }
  }

  deleteTopic = () => {
    let { topicId } = this.props;
    if (confirm('确认删除该主题吗？')) {
      delete_('/api/topic/' + topicId).then((data) => {
        if (data.status) {
          window.location.href = '/';
        } else {
          alert(data.message);
        }
      });
    }
  }

  collectTopic = () => {
    get('/api/topic/' + this.props.topicId + '/collect').then((data) => {
      if (data.status) {
        this.setState({canCollect: 'false'});
      } else {
        alter(data.message);
      }
    });
  }

  cancelCollectTopic = () => {
    get('/api/topic/' + this.props.topicId + '/cancel_collect').then((data) => {
      if (data.status) {
        this.setState({canCollect: 'true'});
      } else {
        alter(data.message);
      }
    });
  }

  render() {
    let { canEdit, canDelete } = this.props;
    let { canCollect } = this.state;
    let editButton = null;
    if (canEdit == 'true') {
      editButton = (
        <a className="item" title="编辑" href={'/t/' + this.props.topicId + '/edit'}>
          <i className="edit icon"></i>
        </a>
      );
    }

    let deleteButton = null;
    if (canDelete == 'true') {
      deleteButton = (
        <a className="item" title="删除" onClick={this.deleteTopic}>
          <i className="remove icon"></i>
        </a>
      );
    }

    let collectButton = null;
    if (canCollect == 'true') {
      collectButton = (
        <a className="item" title="收藏" onClick={this.collectTopic}>
          <i className="outline star icon"></i>
        </a>
      );
    } else {
      collectButton = (
        <a className="item" title="取消收藏" onClick={this.cancelCollectTopic}>
          <i className="yellow star icon"></i>
        </a>
      );
    }

    return (
      <div className="ui compact mini right floated icon menu">
        {editButton}
        {deleteButton}
        {collectButton}
      </div>
    );
  }
}