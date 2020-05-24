class ComponentWithPreloader {
  constructor(container_selector, title) {
    this.container_selector = container_selector;
    this.title = title;
  }

  __render_preloader(container) {
    var container = document.body.querySelector(this.container_selector);

    if(!container) {
      return;
    }

    var html = `
      <h2>${this.title}</h2>
      <div class="row">
        <div class="progress">
          <div class="indeterminate"></div>
        </div>
      </div>
    `;

    container.innerHTML = '';
    container.insertAdjacentHTML('afterbegin', html);
  }

  __sleep(milliseconds) {
    return new Promise(resolve => setTimeout(resolve, milliseconds))
  }
}