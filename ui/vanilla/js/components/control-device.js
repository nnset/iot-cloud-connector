class ControlDevice extends ComponentWithSleep {
  constructor(container_selector, device_id, cloud_connector, title, i18n, icons) {
    super();
    this.device_id = device_id;
    this.cloud_connector = cloud_connector;
    this.i18n = i18n;
    this.icons = icons;
    this.container_selector = container_selector;
    this.title = title;
  }

  render() {
    var container = document.body.querySelector(this.container_selector);

    if(!container) {
      return '';
    }

    var html = `
      <h2>
        ${this.title}
        <div class="request-spinner preloader-wrapper small active">
                <div class="spinner-layer spinner-blue-only">
                    <div class="circle-clipper left">
                        <div class="circle"></div>
                    </div>
                    <div class="gap-patch">
                        <div class="circle"></div>
                    </div>
                    <div class="circle-clipper right">
                        <div class="circle"></div>
                    </div>
                </div>
            </div>
        </div>
      </h2>

      <div class="row">
        <div class="payload-form col s12 m12 l6">
            <form id="payload-form" class="col s12">
                <div class="input-field col s12">
                    <textarea id="payload" class="materialize-textarea"></textarea>
                    <label for="payload">${this.i18n('payload_to_send')}</label>
                </div>

                <div class="col s12 form-buttons ">
                    <div class="submit-command col s6">
                        <button class="btn waves-effect waves-light" type="submit" name="action">${this.i18n('send_as_command')}
                            <i class="material-icons right">${this.icons("send_command")}</i>
                        </button>
                    </div>
                    <div class="submit-query col s6">
                        <button class="btn waves-effect waves-light" type="submit" name="action">${this.i18n('send_as_query')}
                            <i class="material-icons right">${this.icons("send_query")}</i>
                        </button>
                    </div>
                </div>
            </form>
        </div>

        <div class="responses col s12 m12 l6">
            <div class="input-field col s12">
                <textarea id="responses" class="materialize-textarea" readonly></textarea>
                <label for="responses">${this.i18n('responses')}</label>
            </div>
        </div>
      </div>
    `;

    container.innerHTML = '';
    container.insertAdjacentHTML('afterbegin', html);

    const form = document.getElementById('payload-form');
    form.addEventListener('submit', (submit_event) => {
        this.__render_spinner();

        submit_event.stopPropagation();
        submit_event.preventDefault();

        this.__payload_submit(this.cloud_connector, this.device_id)
    });

    return html;
  }

  __payload_submit(cloud_connector, device_id) {
    var form_data = new FormData(document.getElementById('payload-form'));

    cloud_connector.send_command(device_id, form_data).then(data => {
        var responses = document.getElementById('responses');

        responses.innerHTML = responses.innerHTML + JSON.stringify(data) + "\n";
        M.updateTextFields();
        M.textareaAutoResize(responses);

        this.__sleep(300).then(() => {
            var spinner = document.body.querySelector(`${this.container_selector} .request-spinner`);
            spinner.style.opacity = '0';
        });
    });
  }

  __render_spinner() {
    var spinner = document.body.querySelector(`${this.container_selector} .request-spinner`);
    spinner.style.opacity = '1';
  }
}