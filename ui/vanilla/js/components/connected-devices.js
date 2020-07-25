class ConnectedDevices extends ComponentWithPreloader {
  constructor(fetch_data_path, cloud_connector, container_selector, i18n, icons) {
    super(container_selector, i18n('connected_devices'));

    this.fetch_data_path = fetch_data_path;
    this.cloud_connector = cloud_connector;
    this.container_selector = container_selector;
    this.i18n = i18n;
    this.icons = icons;
  }

  render() {
    var container = document.body.querySelector(this.container_selector);

    if (!container) {
      return '';
    }

    this.__render_preloader(container);

    this.cloud_connector.get_data(this.fetch_data_path)
      .then(data => {
        var tableData = this.__connectedDevicesToTableData(data);

        var html = `
          <h2>${this.title}</h2>
          <div class="content row">
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
      'ID', this.i18n('name'), this.i18n('last_connection'), this.i18n('actions')
    ];
  
    for (var [index, device] of Object.entries(data['devices'])) {
      const lastConnection = (new Date(device['last_received_message'] * 1000));
      rows.push([
        device['device_id'],
        device['device_name'],
        lastConnection.toLocaleDateString() + ' ' + lastConnection.toLocaleTimeString(),
        this.__show_connected_device_link(device['device_id'])]
      );
    }

    return {
      columns: columns,
      rows: rows,
    }
  }

  __show_connected_device_link(device_id) {
    var html = `
      <a href = "device.html?id=${device_id}" class="list-action waves-effect waves-light">
        <i class="material-icons left">${this.icons('view_device')}</i>${this.i18n('view_device')}
      </a>
    `;

    return html;
  }
}