class LocaleLoader {
  constructor() {}

  texts() {
    const locale = window.localStorage.getItem('current_locale') || 'en';
    
    if (locale === 'es') {
        return i18n.create(i18n_es_texts)
    }
    
    return i18n.create(i18n_en_texts)
  }
}