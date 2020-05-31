class ComponentWithSleep {
  
  __sleep(milliseconds) {
      return new Promise(resolve => setTimeout(resolve, milliseconds))
  }

}