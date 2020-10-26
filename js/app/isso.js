/* Isso – Ich schrei sonst!
 */

import QRCode from 'qrcode'

import $ from './dom'
import utils from './utils'
import config from './config'
import lnurl from './lnurl'
import api from './api'
import jade from './jade'
import i18n from './i18n'
import identicons from './lib/identicons'
import globals from './globals'

var Postbox = function (parent) {
  function render() {
    return jade.render('postbox', {
      lnurlauth: lnurl.encode(lnurl.authURL),
      authed,
      author: JSON.parse(localStorage.getItem('author')),
      email: JSON.parse(localStorage.getItem('email')),
      website: JSON.parse(localStorage.getItem('website'))
    })
  }

  function postRender() {
    // only display notification checkbox if email is filled in
    var email_edit = function () {
      if (
        config['reply-notifications'] &&
        $("[name='email']", el).value.length > 0
      ) {
        $('.notification-section', el).show()
      } else {
        $('.notification-section', el).hide()
      }
    }
    $("[name='email']", el).on('input', email_edit)
    email_edit()

    // email is not optional if this config parameter is set
    if (config['require-email']) {
      $("[name='email']", el).setAttribute(
        'placeholder',
        $("[name='email']", el)
          .getAttribute('placeholder')
          .replace(/ \(.*\)/, '')
      )
    }

    // author is not optional if this config parameter is set
    if (config['require-author']) {
      $("[name='author']", el).placeholder = $(
        "[name='author']",
        el
      ).placeholder.replace(/ \(.*\)/, '')
    }

    // submit form, initialize optional fields with `null` and reset form.
    // If replied to a comment, remove form completely.
    $('form', el).on('submit', function () {
      if (!el.validate()) {
        return
      }

      var author = $('[name=author]', el).value || null,
        email = $('[name=email]', el).value || null,
        website = $('[name=website]', el).value || null

      localStorage.setItem('author', JSON.stringify(author))
      localStorage.setItem('email', JSON.stringify(email))
      localStorage.setItem('website', JSON.stringify(website))

      api
        .create($('#isso-thread').getAttribute('data-isso-id'), {
          author: author,
          email: email,
          website: website,
          text: $('textarea', el).value,
          parent: parent || null,
          title: $('#isso-thread').getAttribute('data-title') || null,
          notification: $('[name=notification]', el).checked() ? 1 : 0
        })
        .then(function (comment) {
          $('textarea', el).value = ''
          $('textarea', el).blur()
          insert(comment, true)

          if (parent !== null) {
            el.onsuccess()
          }
        })
    })
  }

  var authed = false
  var localStorage = utils.localStorageImpl
  var el = $.htmlify(render())

  // render QR code
  let canvas = el.obj.querySelector('canvas')
  QRCode.toCanvas(canvas, canvas.dataset.qr.toUpperCase())

  // wait for login from wallet
  lnurl.listen(() => {
    authed = true
    el.innerHTML = render()
    postRender()
  })

  el.validate = function () {
    if ($('textarea', this).value.length < 3) {
      $('textarea', this).focus()
      return false
    }
    if (
      config['require-email'] &&
      $("[name='email']", this).value.length <= 0
    ) {
      $("[name='email']", this).focus()
      return false
    }
    if (
      config['require-author'] &&
      $("[name='author']", this).value.length <= 0
    ) {
      $("[name='author']", this).focus()
      return false
    }
    return true
  }

  return el
}

var insert_loader = function (comment, lastcreated) {
  var entrypoint
  if (comment.id === null) {
    entrypoint = $('#isso-root')
    comment.name = 'null'
  } else {
    entrypoint = $('#isso-' + comment.id + ' > .text-wrapper > .isso-follow-up')
    comment.name = comment.id
  }
  var el = $.htmlify(jade.render('comment-loader', {comment: comment}))

  entrypoint.append(el)

  $('a.load_hidden', el).on('click', function () {
    el.remove()
    api
      .fetch(
        $('#isso-thread').getAttribute('data-isso-id'),
        config['reveal-on-click'],
        config['max-comments-nested'],
        comment.id,
        lastcreated
      )
      .then(
        function (rv) {
          if (rv.total_replies === 0) {
            return
          }

          var lastcreated = 0
          rv.replies.forEach(function (commentObject) {
            insert(commentObject, false)
            if (commentObject.created > lastcreated) {
              lastcreated = commentObject.created
            }
          })

          if (rv.hidden_replies > 0) {
            insert_loader(rv, lastcreated)
          }
        },
        function (err) {
          console.log(err)
        }
      )
  })
}

var insert = function (comment, scrollIntoView) {
  var el = $.htmlify(jade.render('comment', {comment: comment}))

  // update datetime every 60 seconds
  var refresh = function () {
    $('.permalink > time', el).textContent = utils.ago(
      globals.offset.localTime(),
      new Date(parseInt(comment.created, 10) * 1000)
    )
    setTimeout(refresh, 60 * 1000)
  }

  // run once to activate
  refresh()

  if (config['avatar']) {
    $('div.avatar > svg', el).replace(identicons.generate(comment.hash, 4, 48))
  }

  var entrypoint
  if (comment.parent === null) {
    entrypoint = $('#isso-root')
  } else {
    entrypoint = $(
      '#isso-' + comment.parent + ' > .text-wrapper > .isso-follow-up'
    )
  }

  entrypoint.append(el)

  if (scrollIntoView) {
    el.scrollIntoView()
  }

  var footer = $(
      '#isso-' + comment.id + ' > .text-wrapper > .isso-comment-footer'
    ),
    header = $(
      '#isso-' + comment.id + ' > .text-wrapper > .isso-comment-header'
    ),
    text = $('#isso-' + comment.id + ' > .text-wrapper > .text')

  var form = null // XXX: probably a good place for a closure
  $('a.reply', footer).toggle(
    'click',
    function (toggler) {
      form = footer.insertAfter(
        new Postbox(comment.parent === null ? comment.id : comment.parent)
      )
      form.onsuccess = function () {
        toggler.next()
      }
      $('textarea', form).focus()
      $('a.reply', footer).textContent = i18n.translate('comment-close')
    },
    function () {
      form.remove()
      $('a.reply', footer).textContent = i18n.translate('comment-reply')
    }
  )

  if (config.vote) {
    var voteLevels = config['vote-levels']
    if (typeof voteLevels === 'string') {
      // Eg. -5,5,15
      voteLevels = voteLevels.split(',')
    }

    // update vote counter
    var votes = function (value) {
      var span = $('span.votes', footer)
      if (span === null) {
        footer.prepend($.new('span.votes', value))
      } else {
        span.textContent = value
      }
      if (value) {
        el.classList.remove('isso-no-votes')
      } else {
        el.classList.add('isso-no-votes')
      }
      if (voteLevels) {
        var before = true
        for (var index = 0; index <= voteLevels.length; index++) {
          if (
            before &&
            (index >= voteLevels.length || value < voteLevels[index])
          ) {
            el.classList.add('isso-vote-level-' + index)
            before = false
          } else {
            el.classList.remove('isso-vote-level-' + index)
          }
        }
      }
    }

    $('a.upvote', footer).on('click', function () {
      api.like(comment.id).then(function (rv) {
        votes(rv.likes - rv.dislikes)
      })
    })

    $('a.downvote', footer).on('click', function () {
      api.dislike(comment.id).then(function (rv) {
        votes(rv.likes - rv.dislikes)
      })
    })

    votes(comment.likes - comment.dislikes)
  }

  $('a.edit', footer).toggle(
    'click',
    function (toggler) {
      var edit = $('a.edit', footer)
      var avatar =
        config['avatar'] || config['gravatar']
          ? $('.avatar', el, false)[0]
          : null

      edit.textContent = i18n.translate('comment-save')
      edit
        .insertAfter($.new('a.cancel', i18n.translate('comment-cancel')))
        .on('click', function () {
          toggler.canceled = true
          toggler.next()
        })

      toggler.canceled = false
      api.view(comment.id, 1).then(function (rv) {
        var textarea = $.new('textarea')

        textarea.value = rv.text
        textarea.focus()

        text.classList.remove('text')
        text.classList.add('textarea-wrapper')

        text.textContent = ''
        text.append(textarea)
      })

      if (avatar !== null) {
        avatar.hide()
      }
    },
    function (toggler) {
      var textarea = $('textarea', text)
      var avatar =
        config['avatar'] || config['gravatar']
          ? $('.avatar', el, false)[0]
          : null

      if (!toggler.canceled && textarea !== null) {
        if (textarea.value.length < 3) {
          textarea.focus()
          toggler.wait()
          return
        } else {
          api.modify(comment.id, {text: textarea.value}).then(function (rv) {
            text.innerHTML = rv.text
            comment.text = rv.text
          })
        }
      } else {
        text.innerHTML = comment.text
      }

      text.classList.remove('textarea-wrapper')
      text.classList.add('text')

      if (avatar !== null) {
        avatar.show()
      }

      $('a.cancel', footer).remove()
      $('a.edit', footer).textContent = i18n.translate('comment-edit')
    }
  )

  $('a.delete', footer).toggle(
    'click',
    function (toggler) {
      var del = $('a.delete', footer)
      var state = !toggler.state

      del.textContent = i18n.translate('comment-confirm')
      del.on('mouseout', function () {
        del.textContent = i18n.translate('comment-delete')
        toggler.state = state
        del.onmouseout = null
      })
    },
    function () {
      var del = $('a.delete', footer)
      api.remove(comment.id).then(function (rv) {
        if (rv) {
          el.remove()
        } else {
          $('span.note', header).textContent = i18n.translate('comment-deleted')
          text.innerHTML = '<p>&nbsp;</p>'
          $('a.edit', footer).remove()
          $('a.delete', footer).remove()
        }
        del.textContent = i18n.translate('comment-delete')
      })
    }
  )

  // remove edit and delete buttons when cookie is gone
  var clear = function (button) {
    if (!utils.cookie('isso-' + comment.id)) {
      if ($(button, footer) !== null) {
        $(button, footer).remove()
      }
    } else {
      setTimeout(function () {
        clear(button)
      }, 15 * 1000)
    }
  }

  clear('a.edit')
  clear('a.delete')

  // show direct reply to own comment when cookie is max aged
  var show = function (el) {
    if (utils.cookie('isso-' + comment.id)) {
      setTimeout(function () {
        show(el)
      }, 15 * 1000)
    } else {
      footer.append(el)
    }
  }

  if (!config['reply-to-self'] && utils.cookie('isso-' + comment.id)) {
    show($('a.reply', footer).detach())
  }

  if (comment.hasOwnProperty('replies')) {
    var lastcreated = 0
    comment.replies.forEach(function (replyObject) {
      insert(replyObject, false)
      if (replyObject.created > lastcreated) {
        lastcreated = replyObject.created
      }
    })
    if (comment.hidden_replies > 0) {
      insert_loader(comment, lastcreated)
    }
  }
}

export default {
  insert: insert,
  insert_loader: insert_loader,
  Postbox: Postbox
}
