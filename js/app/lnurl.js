import bech32 from 'bech32'

import utils from './utils'
import api from './api'

const letters = '0123456789abcdef'

var arr = []
for (let i = 0; i < 64; i++) {
  arr.push(letters[parseInt(Math.random() * 16)])
}
const k1 = arr.join('')

const authURL = api.endpoint + '/lnurlauth?tag=login&k1=' + k1

var sig
var key

var user = JSON.parse(
  utils.localStorage.getItem('stored-user') || '{"name": ""}'
)

export default {
  k1,
  sig,
  key,
  authURL,
  user,
  encode,
  listen
}

function encode(url) {
  return bech32.encode(
    'lnurl',
    bech32.toWords(url.split('').map(c => c.charCodeAt(0))),
    1500
  )
}

function listen(cb) {
  var es = new window.EventSource(api.endpoint + '/lnurlauth/stream?k1=' + k1)
  es.onerror = e => console.log('lnurl sse error', e.data)
  es.addEventListener('auth', e => {
    let data = JSON.parse(e.data)
    sig = data.sig
    key = data.key
    es.close()

    user.sig = sig
    user.key = key
    user.k1 = k1
    utils.localStorage.setItem('stored-user', JSON.stringify(user))

    cb(user)
  })
}
