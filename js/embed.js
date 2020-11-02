import * as Vue from 'vue/dist/vue.esm-bundler.js'
import QRCode from 'qrcode'
import hashbow from 'hashbow'
import marked from 'marked'

import domready from './app/lib/ready'
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
  hashbow: hashbow,
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
  props: ['parent', 'cancellable', 'autofocus', 'currentUser', 'adminKey'],
  data() {
    return {
      lnurlauth: lnurl.encode(lnurl.authURL),
      text: '',
      submitting: false,
      fullKey: false
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
        :disabled="submitting"
      ></textarea>
    </div>
    <section class="actions-section">
      <div class="input-wrapper">
        <input
          type="text"
          name="author"
          :placeholder="f.translate('postbox-author')"
          v-model="currentUser.name"
          :disabled="submitting"
        />
        <span class="spacer">&nbsp;</span>
        <span
          class="key"
          :style="{color: f.hashbow(currentUser.key)}"
          @dblclick="fullKey = !fullKey"
        >
          {{ fullKey ? currentUser.key : currentUser.key.slice(-5) }}
        </span>
        <a v-if="!parent" class="logout" @click="logout">
          {{ f.translate("auth-logout") }}
        </a>
      </div>
      <div class="post-action">
        <button type="submit" :disabled="submitting">{{ f.translate('postbox-submit') }}</button>
      </div>
    </section>
  </form>
</div>
  `,
  methods: {
    postComment(e) {
      this.submitting = true
      e.preventDefault()

      api
        .create({
          author: this.currentUser.name,
          text: this.text,
          parent: this.parent || null
        })
        .then(comment => {
          this.text = ''
          this.$emit('posted', comment)

          utils.localStorage.setItem(
            'stored-user',
            JSON.stringify(this.currentUser)
          )

          this.submitting = false
        })
    },
    logout() {
      this.$emit('logout')
    }
  }
})

app.component('thread', {
  props: ['comments', 'id', 'currentUser', 'adminKey'],
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
  :currentUser="currentUser"
  :adminKey="adminKey"
  :key="comment.id"
/>
  `
})

app.component('comment', {
  props: [
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
    'hidden_replies',
    'currentUser',
    'adminKey'
  ],
  data() {
    return {
      votes: this.likes - this.dislikes,
      createdHumanized: '',
      deleting: false,
      editing: false,
      replying: false,
      banning: false,
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
      return this.authorKey === this.currentUser.key && isRecent
    },
    canDelete() {
      return (
        this.mode !== 4 &&
        (this.canEdit || this.currentUser.key === this.adminKey)
      )
    },
    canBan() {
      return (
        this.currentUser.key === this.adminKey &&
        this.adminKey !== this.authorKey
      )
    },
    canReply() {
      // the server is dumb and only allows one level of comment nesting, so we
      // rather not show the reply button for the third level since comments will
      // be migrated to the second level anyway
      return !this.parent && this.currentUser.key
    },
    keyLastDigits() {
      return (this.authorKey || '').slice(-5)
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
        <span v-if="keyLastDigits.length">
          <span class="spacer">•</span>
          <span class="author key" :style="{color: f.hashbow(authorKey)}">{{ keyLastDigits }}</span>
          <span v-if="banning">
            <a class="ban" @click="banConfirm">{{ f.translate("comment-confirm") }}</a>
            <a class="ban" @click="banning = false">{{ f.translate("comment-cancel") }}</a>
          </span>
          <span v-else-if="canBan">
            <a class="ban" @click="banning = true">{{ f.translate("admin-ban") }}</a>
          </span>
        </span>
        <span class="spacer">•</span>
      </span>
      <a class="permalink" :href="'#ilno-' + id">
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
      <span v-if="deleting">
        <a class="delete" @click="deleteConfirm">{{ f.translate('comment-confirm') }}</a>
        <a class="delete" @click="deleting = false">{{ f.translate('comment-cancel') }}</a>
      </span>
      <span v-else-if="editing">
        <a class="delete" @click="deleting = true">{{ f.translate('comment-delete') }}</a>
        <a class="edit" @click="editSave">{{ f.translate('comment-save') }}</a>
        <a class="edit" @click="editing = false">{{ f.translate('comment-cancel') }}</a>
      </span>
      <span v-else-if="replying">
        <a class="reply" @click="replying = false">{{ f.translate('comment-close') }}</a>
      </span>
      <span v-else>
        <a v-if="canReply" class="reply" @click="replying = true">{{ f.translate('comment-reply') }}</a>
        <a v-if="canEdit" class="edit" @click="editStart">{{ f.translate('comment-edit') }}</a>
        <a v-if="canDelete" class="delete" @click="deleting = true">{{ f.translate('comment-delete') }}</a>
      </span>
    </div>
    <postbox
      v-if="replying"
      :parent="id"
      autofocus
      @posted="handleNewReply"
      :currentUser="currentUser"
      :adminKey="adminKey"
    />
    <div v-if="actualReplies.length" class="ilno-follow-up">
    <thread
      :comments="actualReplies"
      :id="id"
      :currentUser="currentUser"
      :adminKey="adminKey"
    />
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
          text: this.editedText
        })
        .then(r => {
          this.newText = r.text
        })
    },
    deleteConfirm() {
      this.deleting = false
      api.remove(this.id).then(eraseTotally => {
        if (eraseTotally) {
          this.fullyErased = true
        } else {
          this.newMode = 4
        }
      })
    },
    banConfirm() {
      api.ban(this.authorKey)
    },
    updateHumanizedDate() {
      this.createdHumanized = utils.ago(
        new Date(),
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
      comments: [],
      adminKey: null,
      showingBanList: false
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
  <h4>
    {{ heading }}
    <span v-if="user.key === adminKey">
      <a class="ban" v-if="showingBanList" @click="showingBanList = false">{{ f.translate('hide-banned') }}</a>
      <a class="ban" v-else @click="showBanList">{{ f.translate('show-banned') }}</a>
    </span>
  </h4>
  <ul v-if="showingBanList">
    <li v-for="b in showingBanList">
      <span :style="{color: f.hashbow(b.key)}" title="b.banned_at">{{ b.key.slice(-5) }}</span>
      <a class="ban" @click="unban(b.key)">{{ f.translate("admin-unban") }}</a>
    </li>
  </ul>
  <div class="lnurl" v-if="!user.key">
    <p>{{ f.translate("auth-login") }}</p>
    <a :href="'lightning:' + lnurlauth">
      <qrcode :value="lnurlauth" />
    </a>
    <p>
      {{ lnurlauth }}
    </p>
  </div>
  <postbox
    v-else
    @posted="handleNewComment"
    @logout="logout"
    :currentUser="user"
    :adminKey="adminKey"
  />
  <div id="ilno-root">
    <thread :comments="comments" :id="0" :currentUser="user" :adminKey="adminKey" />
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
    },
    logout() {
      lnurl.logout()
      this.user = {...lnurl.user}
      this.authListen()
    },
    showBanList() {
      api.banned().then(banned => {
        this.showingBanList = banned
      })
    },
    unban(key) {
      api.unban(key).then(this.showBanList)
    }
  },
  mounted() {
    this.authListen()
    this.fetchComments()

    // get admin key to check if we are it
    api.getConfig().then(config => {
      this.adminKey = config.admin
    })
  }
})

window.Ilno = {
  init: init
}
