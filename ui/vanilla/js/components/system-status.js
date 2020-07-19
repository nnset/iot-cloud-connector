class SystemStatus extends ComponentWithPreloader {
  constructor(fetch_data_path, cloud_connector, container_selector, title, i18n, icons) {
    super(container_selector, title);

    this.fetch_data_path = fetch_data_path;
    this.cloud_connector = cloud_connector;
    this.container_selector = container_selector;
    this.i18n = i18n;
    this.icons = icons;
  }

  render() {
    var container = document.body.querySelector(this.container_selector);

    if(!container) {
      return '';
    }

    this.__render_preloader(container);

    this.cloud_connector.get_data(this.fetch_data_path)
      .then(data => {
        var metrics = '';
        for (var [metric_key, metric_value] of Object.entries(data['metrics'])) {
          metrics += 
            (new SystemMetric(
                metric_key, 
                metric_value, 
                this.i18n(metric_key), 
                this.icons(metric_key), 
                this.i18n(data['units'][metric_key]))
            ).render();
        }

        var html = `
          <h2>${this.title}</h2>
          <div class="row">
            ${metrics}
          </div>
        `;

        this.__sleep(500).then(
          () => {
            container.innerHTML = '';
            container.insertAdjacentHTML('afterbegin', html);
            this.__subscribe_to_stream(this.container_selector)
          }
        );

        return html;
      });
    }

    __subscribe_to_stream(container_selector) {
      const event_source = new EventSource(this.cloud_connector.subscribe_to_system_status_sse_url());
 
      event_source.onerror = function(error) {
        console.error("EventSource failed: ", error);
      };

      event_source.addEventListener("system_status", function(e) {
        const data = JSON.parse(e.data);
        
        if (data.metric) {
          var metric_dom = document.body.querySelector(`${container_selector} [data-metric="${data.metric}"] .value`);
          metric_dom.innerHTML = data.value;
        }
      });
    }
}