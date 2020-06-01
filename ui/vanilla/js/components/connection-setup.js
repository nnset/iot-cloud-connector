class ConnectionSetup {
  constructor(i18n) {
    this.i18n = i18n;

    const form = document.getElementById('connection_setup_form');
    form.addEventListener('submit', (submit_event) => {
        submit_event.stopPropagation();
        submit_event.preventDefault();

        this.__setup();
    });
  }

  __setup() {
    var cloud_connector_url = document.getElementById('cloud_connector_url');
    
    if (cloud_connector_url.value.length > 0) {
        window.localStorage.setItem('cloud_connector_url', cloud_connector_url.value);
    
        var cloud_connector_bearer = document.getElementById('cloud_connector_bearer');
        window.localStorage.setItem('cloud_connector_bearer', cloud_connector_bearer.value);

        window.location = 'dashboard.html';
    }
  }
}
