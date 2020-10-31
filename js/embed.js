import * as Vue from 'vue/dist/vue.esm-bundler.js'
import QRCode from 'qrcode'
import hashbow from 'hashbow'
import marked from 'marked'

import domready from './app/lib/ready'
import globals from './app/globals'
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
  const root = document.getElementById('ilno-thread')
  if (!root) {
    return console.log('abort, #ilno-thread is missing')
  }

  root.appendChild(document.createElement('root'))

  const style = document.createElement('style')
  style.id = 'ilno-style'
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
  props: ['parent', 'user', 'cancellable', 'autofocus'],
  data() {
    return {
      lnurlauth: lnurl.encode(lnurl.authURL),
      text: ''
    }
  },
  template: `
<div class="ilno-postbox">
  <form @submit="postComment">
    <div class="textarea-wrapper">
      <textarea
        :placeholder="f.translate('postbox-text')"
        v-model="text"
        :autofocus="autofocus"
    ></textarea>
    </div>
    <section class="auth-section">
      <p class="input-wrapper">
        <input type="text" name="author" :placeholder="f.translate('postbox-author')" v-model="user.name" />
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
          this.$emit('posted', comment)

          utils.localStorage.setItem('stored-user', JSON.stringify(this.user))
        })
    }
  }
})

app.component('thread', {
  props: ['user', 'comments', 'id'],
  data() {
    return {}
  },
  computed: {
    count() {
      return this.comments.length
    }
  },
  template: `
<comment
  v-for="comment in comments"
  v-bind="comment"
  :authorKey="comment.key"
  :user="user"
  :key="comment.id"
/>
  `
})

app.component('comment', {
  props: [
    'user',
    'id',
    'parent',
    'created',
    'authorKey',
    'author',
    'text',
    'replies',
    'mode',
    'likes',
    'dislikes',
    'total_replies',
    'hidden_replies'
  ],
  data() {
    return {
      votes: this.likes - this.dislikes,
      createdHumanized: '',
      deleting: false,
      editing: false,
      replying: false,
      editedText: '',
      newText: null,
      newReplies: [],
      newMode: null,
      fullyErased: false
    }
  },
  computed: {
    actualText() {
      return this.newText || this.text
    },
    formattedText() {
      return marked(this.actualText)
    },
    actualReplies() {
      return this.newReplies.concat(this.replies || [])
    },
    actualMode() {
      return this.newMode || this.mode
    },
    createdReadable() {
      return new Date(parseInt(this.created, 10) * 1000).toString()
    },
    createdISO() {
      return new Date(parseInt(this.created, 10) * 1000).toISOString()
    },
    canEdit() {
      let isRecent =
        new Date().getTime() - parseInt(this.created) * 1000 <
        1000 * 60 * 60 * 6 // 6 hours
      return this.authorKey === this.user.key && isRecent
    },
    canReply() {
      // the server is dumb and only allows one level of comment nesting, so we
      // rather not show the reply button for the third level since comments will
      // be migrated to the second level anyway
      return !this.parent
    },
    keyLastDigits() {
      return (this.authorKey || '').slice(-5)
    },
    keyColor() {
      return hashbow(this.authorKey)
    }
  },
  template: `
<div v-if="fullyErased"></div>
<div v-else class="ilno-comment" :id="'ilno-' + id" ref="this">
  <div class="text-wrapper">
    <div class="ilno-comment-header" role="meta">
      <span v-if="actualMode === 4">
        <span class="spacer">×</span>
      </span>
      <span v-else>
        <span class="author name">{{ author }}</span>
        <span class="spacer">•</span>
        <span class="author key" :style="{color: keyColor}">{{ keyLastDigits }}</span>
        <span class="spacer" v-if="keyLastDigits.length">•</span>
      </span>
      <a class="permalink" href="#ilno-1">
        <time :title="createdReadable" :datetime="createdISO">
          {{ createdHumanized }}
        </time>
      </a>
      <span v-if="actualMode === 2" class="note">{{ f.translate('comment-queued') }}</span>
      <span v-if="actualMode === 4" class="note">{{ f.translate('comment-deleted') }}</span>
    </div>
    <div :class="{'textarea-wrapper': editing, 'text': true}">
      <div v-if="actualMode === 4"><!--- deleted ---></div>
      <textarea v-else-if="editing" v-model="editedText"></textarea>
      <div v-else v-html="formattedText" />
    </div>
    <div class="ilno-comment-footer">
      <span class="votes" v-if="votes !== 0">{{ votes }}</span>
      <a class="upvote" @click="upvote"><arrow-up /></a>
      <span class="spacer">|</span>
      <a class="downvote" @click="downvote"><arrow-down /></a>
      <span class="spacer">&nbsp;</span>
      <span v-if="deleting">
        <a class="delete" @click="deleteConfirm">{{ f.translate('comment-confirm') }}</a>
        <span class="spacer">&nbsp;</span>
        <a class="delete" @click="deleteCancel">{{ f.translate('comment-cancel') }}</a>
      </span>
      <span v-else-if="editing">
        <a class="delete" @click="deletePrepare">{{ f.translate('comment-delete') }}</a>
        <span class="spacer">&nbsp;</span>
        <a class="edit" @click="editSave">{{ f.translate('comment-save') }}</a>
        <span class="spacer">&nbsp;</span>
        <a class="edit" @click="editCancel">{{ f.translate('comment-cancel') }}</a>
      </span>
      <span v-else-if="replying">
        <a class="reply" @click="replyCancel">{{ f.translate('comment-close') }}</a>
      </span>
      <span v-else>
        <a v-if="canReply" class="reply" @click="reply">{{ f.translate('comment-reply') }}</a>
        <span class="spacer">&nbsp;</span>
        <a v-if="canEdit" class="edit" @click="editStart">{{ f.translate('comment-edit') }}</a>
        <span class="spacer">&nbsp;</span>
        <a v-if="canEdit" class="delete" @click="deletePrepare">{{ f.translate('comment-delete') }}</a>
      </span>
    </div>
    <postbox :user="user" v-if="replying" :parent="id" autofocus @posted="handleNewReply" />
    <div v-if="actualReplies.length" class="ilno-follow-up">
      <thread :comments="actualReplies" :id="id" :user="user" />
    </div>
  </div>
</div>
  `,
  methods: {
    handleNewReply(reply) {
      this.replying = false
      this.newReplies = [reply, ...this.newReplies]
    },
    upvote() {
      api.like(this.id).then(r => {
        this.votes = r.likes - r.dislikes
      })
    },
    downvote() {
      api.dislike(this.id).then(r => {
        this.votes = r.likes - r.dislikes
      })
    },
    editStart() {
      this.editedText = this.actualText
      this.editing = true
    },
    editSave() {
      this.editing = false
      api
        .modify(this.id, {
          key: this.user.key,
          k1: this.user.k1,
          sig: this.user.sig,
          text: this.editedText
        })
        .then(r => {
          this.newText = r.text
        })
    },
    editCancel() {
      this.editing = false
    },
    deletePrepare() {
      this.deleting = true
    },
    deleteCancel() {
      this.deleting = false
    },
    deleteConfirm() {
      this.deleting = false
      api
        .remove(this.id, {
          key: this.user.key,
          k1: this.user.k1,
          sig: this.user.sig
        })
        .then(eraseTotally => {
          if (eraseTotally) {
            this.fullyErased = true
          } else {
            this.newMode = 4
          }
        })
    },
    reply() {
      this.replying = true
    },
    replyCancel() {
      this.replying = false
    },
    updateHumanizedDate() {
      this.createdHumanized = utils.ago(
        globals.offset.localTime(),
        new Date(parseInt(this.created, 10) * 1000)
      )
    }
  },
  mounted() {
    // update datetime every 60 seconds
    setInterval(this.updateHumanizedDate, 60000)
    this.updateHumanizedDate()

    // scroll into view
    if (
      window.location.hash.length > 0 &&
      window.location.hash.match('^#ilno-[0-9]+$')
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
  <postbox v-else :user="user" @posted="handleNewComment" />
  <div id="ilno-root">
    <thread :comments="comments" :id="0" :user="user" />
  </div>
</div>
  `,
  methods: {
    handleNewComment(x) {
      this.fetchComments()
    },
    fetchComments() {
      api.fetch().then(
        resp => {
          this.count = resp.total_replies
          this.comments = (resp.replies || []).sort(
            (a, b) => b.created - a.created
          )

          if (resp.hidden_replies > 0) {
            // TODO
          }
        },
        err => {
          console.log(err)
        }
      )
    },
    authListen() {
      if (!this.user.key) {
        // wait for login from wallet
        lnurl.listen(user => {
          this.user = {...user}
        })
      }
    }
  },
  mounted() {
    this.authListen()
    this.fetchComments()
  }
})

window.Ilno = {
  init: init
}
