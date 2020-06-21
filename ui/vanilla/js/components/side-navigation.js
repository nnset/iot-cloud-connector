class SideNavigation {
  constructor() {
    const locale_selectors = document.body.querySelectorAll('.sidenav .locale a');

    locale_selectors.forEach(link => {
      const current_locale = window.localStorage.getItem('current_locale') || 'en';
      const locale = link.dataset.locale || 'en';

      if (current_locale === locale){
        link.classList.add('selected');
      }

      link.addEventListener('click', function (event) {
        event.stopPropagation();
        event.preventDefault();

        if (current_locale !== locale) {
          window.localStorage.setItem('current_locale', locale);
          location.reload();
        }
      });
    });
  }
}
