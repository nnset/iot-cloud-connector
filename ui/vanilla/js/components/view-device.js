class ViewDevice extends ComponentWithPreloader {
    constructor(device_id, cloud_connector, container_selector, title) {
      super(container_selector, title);

      this.device_id = device_id;
      this.cloud_connector = cloud_connector;
      this.container_selector = container_selector;
      this.already_rendered = false;
      this.refresh_handler_id = null;
      this.refresh_interval = 3000;
    }
  
    render() {
      var container = document.body.querySelector(this.container_selector);
  
      if(!container) {
        return '';
      }

      this.__render_preloader(container);

      this.cloud_connector.getData(this.cloud_connector.show_device_path(this.device_id))
        .then(data => {
          var html = `
              <h2>${this.title}</h2>
              <div class="status">
                  <p><span data-metric="uptime">${data['uptime']}</span> uptime seconds</p>
              </div>
              <div class="messages">
                  <p><span data-metric="received_messages">${data['received_messages']}</span> received messages</p>
                  <p><span data-metric="received_messages_per_second">${data['received_messages_per_second']}</span> received messages per second</p>
                  <p><span data-metric="sent_messages">${data['sent_messages']}</span> sent messages</p>
                  <p><span data-metric="sent_messages_per_second">${data['sent_messages_per_second']}</span> sent messages per second</p>
              </div>
          `;

          this.__sleep(500).then(() => {
            container.innerHTML = '';
            container.insertAdjacentHTML('afterbegin', html);

            this.already_rendered = true;
            this.__refresh(this.refresh_interval);
          });

          return html;
        });
      }

    __refresh(interval) {
      this.refresh_handler_id = setInterval(() => {
        this.cloud_connector.getData(this.cloud_connector.show_device_path(this.device_id))
        .then(data => {  
          for (var [metric_key, metric_value] of Object.entries(data)) {
            var metric_dom = document.body.querySelector(`${this.container_selector} [data-metric="${metric_key}"]`);
                metric_dom.innerHTML = metric_value;
          }
        });
      }, interval);
    }
  }