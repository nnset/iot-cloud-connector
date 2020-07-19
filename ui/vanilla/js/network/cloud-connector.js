class CloudConnector {
  constructor(api_url) {
    this.api_url = api_url;
  }

  async get_data(path = '') {
    const response = await fetch(`${this.api_url}/${path}`, {
      method: 'GET',
      mode: 'cors',
      headers: {
        'Content-Type': 'application/json',
        'Access-Control-Request-Method': 'GET'
      }
    });

    return response.json();
  }

  async send_command(device_id, payload) {
    return this.__send_to_device(device_id, 'devices/command', payload);
  }

  async send_query(device_id, payload) {
    return this.__send_to_device(device_id, 'devices/query', payload);
  }

  async __send_to_device(device_id, path, payload) {
    const response = await fetch(`${this.api_url}/${path}/${device_id}`, {
        method: 'POST',
        mode: 'cors',
        headers: {
          'Content-Type': 'application/json',
          'Access-Control-Request-Method': 'POST'
        },
        body: JSON.stringify(payload),
      });
  
      return response.json();
  }

  show_device_path(device_id) {
    return `devices/${device_id}/show`;
  }

  subscribe_to_system_status_sse_url() {
    return `${this.api_url}/cloud-connector/status/stream`;
  }
}
