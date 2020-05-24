class ConnectedDevices extends ComponentWithPreloader {
  constructor(fetch_data_path, cloud_connector, container_selector, title) {
    super(container_selector, title);

    this.fetch_data_path = fetch_data_path;
    this.cloud_connector = cloud_connector;
    this.container_selector = container_selector;
  }

  render() {
    var container = document.body.querySelector(this.container_selector);

    if (!container) {
      return '';
    }
    
    this.__render_preloader(container);

    this.cloud_connector.getData(this.fetch_data_path)
      .then(data => {     
        var tableData = this.__connectedDevicesToTableData(data);

        var html = `
          <h2>${this.title}</h2>
          <div class="row">
            ${(new Table(tableData.columns, tableData.rows)).render()}
          </div>
        `;

        this.__sleep(700).then(
          () => {
            container.innerHTML = '';
            container.insertAdjacentHTML('afterbegin', html);
          }
        );

        return html;
      });
  }

  __connectedDevicesToTableData(data) {
    var columns, rows = [];

    columns = [
      'ID', 'Last connection', 'Actions'
    ];
  
    for (var [index, device_id] of Object.entries(data['devices'])) {
      rows.push([device_id, 'N/A', this.__show_connected_device_link(device_id)]);
    }

    return {
      columns: columns,
      rows: rows,
    }
  }

  __show_connected_device_link(device_id) {
    var html = `
      <a href = "device.html?id=${device_id}" class="waves-effect waves-light btn-small">
        <i class="material-icons left">cloud</i>View
      </a>
    `;

    return html;
  }
}