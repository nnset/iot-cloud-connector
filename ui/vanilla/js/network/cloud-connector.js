class CloudConnector {
  constructor(apiURL) {
    this.apiURL = apiURL;
  }

  async getData(path = '') {
    const response = await fetch(`${this.apiURL}/${path}`, {
      method: 'GET',
      mode: 'cors',
      headers: {
        'Content-Type': 'application/json',
        'Access-Control-Request-Method': 'GET'
      }
    });

    return response.json();
  }

  show_device_path(device_id) {
    return `devices/${device_id}/show`;
  }
}
