class SystemMetric {
  constructor(metric_key, metric_value, metric_human_name, icon, metric_unit) {
    this.metric_key = metric_key;
    this.metric_value = metric_value;
    this.metric_human_name = metric_human_name;
    this.icon = icon;
    this.metric_unit = metric_unit || '';    
  }

  render(container = null) {
    var html = `
        <div class="metric system-metric col s5 m2" data-metric="${this.metric_key}">
            <div class="content">
                <div class="top-value">
                    <i class="material-icons">${this.icon}</i>${this.metric_human_name}
                </div>
                <div class="value">
                    ${this.metric_value}
                </div>
                <div class="unit">${this.metric_unit}</div>
            </div>
        </div>
    `;

    if (container) {
      container.insertAdjacentHTML('afterbegin', html);
    }

    return html;
  }
}
