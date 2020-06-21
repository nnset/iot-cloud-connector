class TopNavigation {
  constructor() {
    const cloud_connector_url = window.localStorage.getItem('cloud_connector_url');

    if (cloud_connector_url) {
      const host = document.body.querySelector('.nav-wrapper .host');
      
      if (host) {
        host.innerHTML = cloud_connector_url;
      }
    }
  }
}
