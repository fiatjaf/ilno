var config = {
  lang: (navigator.language || navigator.userLanguage).split('-')[0]
}

var js = document.getElementsByTagName('script')

for (var i = 0; i < js.length; i++) {
  for (var j = 0; j < js[i].attributes.length; j++) {
    var attr = js[i].attributes[j]
    if (attr.name === 'data-ilno') {
      let endpoint = attr.value
      //  strip trailing slash
      if (endpoint[endpoint.length - 1] === '/') {
        endpoint = endpoint.substring(0, endpoint.length - 1)
      }
      config.endpoint = endpoint
    } else if (/^data-ilno-/.test(attr.name)) {
      try {
        config[attr.name.substring(10)] = JSON.parse(attr.value)
      } catch (ex) {
        config[attr.name.substring(10)] = attr.value
      }
    }
  }
}

export default config
