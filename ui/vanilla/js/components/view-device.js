class ViewDevice extends ComponentWithPreloader {
    constructor(device_id, cloud_connector, container_selector, title, i18n, icons) {
      super(container_selector, title);

      this.device_id = device_id;
      this.cloud_connector = cloud_connector;
      this.container_selector = container_selector;
      this.already_rendered = false;
      this.refresh_handler_id = null;
      this.refresh_interval = 3000;
      this.i18n = i18n;
      this.icons = icons;
    }
  
    render() {
      var container = document.body.querySelector(this.container_selector);
  
      if(!container) {
        return '';
      }

      this.__render_preloader(container);

      this.cloud_connector.get_data(this.cloud_connector.show_device_path(this.device_id))
        .then(data => {

          var metrics = '';

          for (var [metric_key, metric_value] of Object.entries(data['metrics'])) {
            metrics += new SystemMetric(
              metric_key, 
              metric_value, 
              this.i18n(metric_key), 
              this.icons(metric_key), 
              this.i18n(data['units'][metric_key])
            ).render();
          }

          var html = `
            <section class="view-device">
              <h2>${this.title} <small>(${this.device_id})</small></h2>
              <div class="metrics row">
                ${metrics}
              <div>
              
              <div class="messages row">
              <div>
            </section>
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
        this.cloud_connector.get_data(this.cloud_connector.show_device_path(this.device_id))
        .then(data => {  

          for (var [metric_key, metric_value] of Object.entries(data['metrics'])) {
            var metric_dom = document.body.querySelector(`${this.container_selector} [data-metric="${metric_key}"] .value`);
                metric_dom.innerHTML = metric_value;
          }

        });
      }, interval);
    }
  }