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

    this.cloud_connector.getData(this.fetch_data_path)
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
          }
        );

        return html;
      });
    }
}