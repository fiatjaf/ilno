import * as Vue from 'vue/dist/vue.esm-bundler.js'
import QRCode from 'qrcode'

import domready from './app/lib/ready'
import globals from './app/globals'
import config from './app/config'
import i18n from './app/i18n'
import api from './app/api'
import lnurl from './app/lnurl'
import utils from './app/utils'
// import count from './app/count'
import css from './app/css'

const app = Vue.createApp({})
app.config.globalProperties.f = {
  translate: i18n.translate,
  pluralize: i18n.pluralize,
  log: console.log
}

domready(() => {
  init()
})

function init() {
  const root = document.getElementById('isso-thread')
  if (!root) {
    return console.log('abort, #isso-thread is missing')
  }

  root.appendChild(document.createElement('root'))

  const style = document.createElement('style')
  style.id = 'isso-style'
  style.type = 'text/css'
  style.textContent = css.inline
  document.head.appendChild(style)

  // count()

  app.mount(root)
}

app.component('qrcode', {
  props: ['value'],
  template: '<canvas ref="canvas"></canvas>',
  mounted() {
    QRCode.toCanvas(this.$refs.canvas, this.value)
  }
})

app.component('arrow-down', {
  template:
    '<svg width="16" height="16" viewBox="0 0 32 32" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" fill="gray"><g><path d="M 24.773,13.701c-0.651,0.669-7.512,7.205-7.512,7.205C 16.912,21.262, 16.456,21.44, 16,21.44c-0.458,0-0.914-0.178-1.261-0.534 c0,0-6.861-6.536-7.514-7.205c-0.651-0.669-0.696-1.87,0-2.586c 0.698-0.714, 1.669-0.77, 2.522,0L 16,17.112l 6.251-5.995 c 0.854-0.77, 1.827-0.714, 2.522,0C 25.47,11.83, 25.427,13.034, 24.773,13.701z"></path></g></svg>'
})

app.component('arrow-up', {
  template:
    '<svg width="16" height="16" viewBox="0 0 32 32" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" fill="gray"><g><path d="M 24.773,18.299c-0.651-0.669-7.512-7.203-7.512-7.203C 16.912,10.739, 16.456,10.56, 16,10.56c-0.458,0-0.914,0.179-1.261,0.536 c0,0-6.861,6.534-7.514,7.203c-0.651,0.669-0.696,1.872,0,2.586c 0.698,0.712, 1.669,0.77, 2.522,0L 16,14.89l 6.251,5.995 c 0.854,0.77, 1.827,0.712, 2.522,0C 25.47,20.17, 25.427,18.966, 24.773,18.299z"></path></g></svg>'
})

app.component('postbox', {
  props: ['parent', 'onsuccess', 'user'],
  data() {
    return {
      lnurlauth: lnurl.encode(lnurl.authURL),
      text: ''
    }
  },
  template: `
<div class="isso-postbox">
  <form @submit="postComment">
    <div class="textarea-wrapper">
      <textarea :placeholder="f.translate('postbox-text')" :value="text"></textarea>
    </div>
    <section class="auth-section">
      <p class="input-wrapper">
        <input type="text" name="author" :placeholder="f.translate('postbox-author')" :value="user.name" />
      </p>
      <p class="post-action">
        <button type="submit">{{ f.translate('postbox-submit') }}</button>
      </p>
    </section>
  </form>
</div>
  `,
  methods: {
    postComment(e) {
      e.preventDefault()

      api
        .create({
          author: this.user.name,
          sig: this.user.sig,
          key: this.user.key,
          k1: this.user.k1,
          text: this.text,
          parent: this.parent || null
        })
        .then(comment => {
          this.text = ''

          if (this.parent !== null) {
            this.onsuccess()
          }
        })
    }
  }
})

app.component('thread', {
  props: ['comments', 'id'],
  data() {
    return {}
  },
  computed: {
    count() {
      return this.comments.length
    }
  },
  template: `
<comment v-for="comment in comments" v-bind="comment" :key="comment.id" />
  `
})

app.component('comment', {
  props: [
    'id',
    'created',
    'author',
    'text',
    'votes',
    'replies',
    'mode',
    'likes',
    'dislikes',
    'total_replies',
    'hidden_replies'
  ],
  data() {
    return {
      createdHumanized: ''
    }
  },
  computed: {
    createdReadable() {
      return new Date(parseInt(this.created, 10) * 1000).toString()
    },
    createdISO() {
      return new Date(parseInt(this.created, 10) * 1000).toISOString()
    },
    canEdit() {
      return true
    },
    canDelete() {
      return false
    }
  },
  template: `
<div class="isso-comment isso-no-votes" :id="'isso-' + id" ref="this">
  <div class="text-wrapper">
    <div class="isso-comment-header" role="meta">
      <span class="author">{{ author }}</span>
      <span class="spacer">•</span>
      <a class="permalink" href="#isso-1">
        <time :title="createdReadable" :datetime="createdISO">
          {{ createdHumanized }}
        </time>
      </a>
      <span v-if="mode === 2" class="note">{{ f.translate('comment-queued') }}</span>
      <span v-if="mode === 4" class="note">{{ f.translate('comment-deleted') }}</span>
    </div>
    <div class="text"><p v-if="mode !== 4">{{ text }}</p></div>
    <div class="isso-comment-footer">
      <span class="votes">{{ votes }}</span>
      <a class="upvote"><arrow-up /></a>
      <span class="spacer">|</span>
      <a class="downvote"><arrow-down /></a>
      <a class="reply">{{ f.translate('comment-reply') }}</a>
      <a v-if="canEdit" class="edit">{{ f.translate('comment-edit') }}</a>
      <a v-if="canDelete" class="delete">{{ f.translate('comment-delete') }}</a>
    </div>
    <div v-if="replies" class="isso-follow-up">
      <thread :comments="replies" :id="id" />
    </div>
  </div>
</div>
  `,
  mounted() {
    // update datetime every 60 seconds
    setInterval(() => {
      this.createdHumanized = utils.ago(
        globals.offset.localTime(),
        new Date(parseInt(this.created, 10) * 1000)
      )
    }, 60000)

    // scroll into view
    if (
      window.location.hash.length > 0 &&
      window.location.hash.match('^#isso-[0-9]+$')
    ) {
      this.$refs.this.scrollIntoView()
    }
  }
})

app.component('root', {
  data() {
    return {
      count: null,
      lnurlauth: lnurl.encode(lnurl.authURL),
      user: lnurl.user,
      comments: []
    }
  },
  computed: {
    heading() {
      if (this.count) {
        return i18n.pluralize('num-comments', this.count)
      }
      return i18n.translate('no-comments')
    }
  },
  template: `
<div>
  <h4>{{ heading }}</h4>
  <div v-if="!user.key">
    <qrcode :value="lnurlauth" />
    <p style="white-space: pre-wrap; font-family: monospace; word-break: break-all">
      {{ lnurlauth }}
    </p>
  </div>
  <postbox v-else :user="user" />
  <div id="isso-root">
    <thread :comments="comments" :id="0" />
  </div>
</div>
  `,
  mounted() {
    // wait for login from wallet
    lnurl.listen(user => {
      this.user = {...user}
    })

    // fetch comments
    api.fetch(config['max-comments-top'], config['max-comments-nested']).then(
      resp => {
        this.count = resp.total_replies
        this.comments = resp.replies.sort((a, b) => b.created - a.created)

        if (resp.hidden_replies > 0) {
          // TODO
        }
      },
      err => {
        console.log(err)
      }
    )
  }
})

window.Isso = {
  init: init
}
