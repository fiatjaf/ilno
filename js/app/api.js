import lnurl from './lnurl'
import config from './config'

const endpoint = config.endpoint
const salt = 'Eech7co8Ohloopo9Ol6baimi'
function location() {
  return window.location.pathname
}

function curl(method, url, data) {
  return window
    .fetch(url, {
      method,
      body: method === 'GET' ? undefined : JSON.stringify(data),
      headers: {
        'lnurl-auth-k1': lnurl.user.k1,
        'lnurl-auth-key': lnurl.user.key,
        'lnurl-auth-sig': lnurl.user.sig
      }
    })
    .then(r => {
      if (r.ok) return r.json()
      return r.text().then(text => {
        throw new Error(text)
      })
    })
}

var qs = function (params) {
  var rv = ''
  for (let key in params) {
    if (params[key]) {
      rv += key + '=' + encodeURIComponent(params[key]) + '&'
    }
  }

  return rv.substring(0, rv.length - 1) // chop off trailing "&"
}

var create = function (data) {
  let rootElement = document.getElementById('ilno-thread')
  let tid = rootElement.dataset.ilnoId
  let title = rootElement.dataset.ilnoTitle
  data.title = title

  return curl('POST', endpoint + '/new?' + qs({uri: tid || location()}), data)
}

var modify = function (id, data) {
  return curl('PUT', endpoint + '/id/' + id, data)
}

var remove = function (id) {
  return curl('DELETE', endpoint + '/id/' + id, null).then(
    body => body.id === 0
  )
}

var fetch = function (parent, lastcreated) {
  let rootElement = document.getElementById('ilno-thread')
  let tid = rootElement.dataset.ilnoId

  if (typeof parent === 'undefined') {
    parent = null
  }

  var query_dict = {uri: tid || location(), after: lastcreated, parent: parent}

  return curl('GET', endpoint + '/?' + qs(query_dict), null).catch(() => ({
    total_replies: 0
  }))
}

var count = function (urls) {
  return curl('POST', endpoint + '/count', urls)
}

var like = function (id) {
  return curl('POST', endpoint + '/id/' + id + '/like', null)
}

var dislike = function (id) {
  return curl('POST', endpoint + '/id/' + id + '/dislike', null)
}

var getConfig = function () {
  return curl('GET', endpoint + '/config', null)
}

var banned = function () {
  return curl('GET', endpoint + '/banned', null).then(r => r || [])
}

var ban = function (key) {
  return curl('POST', endpoint + '/ban/' + key, null)
}

var unban = function (key) {
  return curl('POST', endpoint + '/unban/' + key, null)
}

export default {
  endpoint: endpoint,
  salt: salt,

  create: create,
  modify: modify,
  remove: remove,
  fetch: fetch,
  count: count,
  like: like,
  dislike: dislike,

  banned,
  ban,
  unban,
  getConfig
}
