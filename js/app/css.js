export default {
  inline: `
#ilno-thread * {
    -webkit-box-sizing: border-box;
    -moz-box-sizing: border-box;
    box-sizing: border-box;
}
#ilno-thread a {
    text-decoration: none;
    cursor: pointer;
}
#ilno-thread {
    max-width: 68em;
    padding: 0;
    margin: 0 auto;
}
#ilno-thread > h4 {
    color: #555;
    font-weight: bold;
}
#ilno-thread textarea {
    min-height: 58px;
    min-width: 100%;
    max-width: 100%;
    outline: 0;
}
#ilno-thread .lnurl {
    white-space: pre-wrap;
    font-family: monospace;
    word-break: break-all;
}
#ilno-root .ilno-comment {
    padding-top: 0.95em;
    margin: 0.95em auto;
}
#ilno-root .ilno-comment:not(:first-of-type),
.ilno-follow-up .ilno-comment {
    border-top: 1px solid rgba(0, 0, 0, 0.1);
}
.ilno-comment > div.text-wrapper {
    display: block;
}
.ilno-comment .ilno-follow-up {
    padding-left: calc(7% + 20px);
}
.ilno-comment > div.text-wrapper > .ilno-comment-header, .ilno-comment > div.text-wrapper > .ilno-comment-footer {
    font-size: 0.95em;
}
.ilno-comment > div.text-wrapper > .ilno-comment-header {
    font-size: 0.85em;
}
#ilno-thread .spacer {
    padding: 0 6px;
}
#ilno-thread .spacer,
.ilno-comment > div.text-wrapper > .ilno-comment-header a.permalink,
.ilno-comment > div.text-wrapper > .ilno-comment-header .note,
.ilno-comment > div.text-wrapper > .ilno-comment-header a.parent {
    color: gray !important;
    font-weight: normal;
    text-shadow: none !important;
}
.ilno-comment > div.text-wrapper > .ilno-comment-header .spacer:hover,
.ilno-comment > div.text-wrapper > .ilno-comment-header a.permalink:hover,
.ilno-comment > div.text-wrapper > .ilno-comment-header .note:hover,
.ilno-comment > div.text-wrapper > .ilno-comment-header a.parent:hover {
    color: #606060 !important;
}
.ilno-comment > div.text-wrapper > .ilno-comment-header .note {
    float: right;
}
.ilno-comment > div.text-wrapper > .ilno-comment-header .author.name {
    font-weight: bold;
    color: #777;
}
.ilno-comment > div.text-wrapper > .ilno-comment-header .author.key {
    font-weight: bold;
    color: #444;
    font-family: monospace;
}
.ilno-comment > div.text-wrapper > .textarea-wrapper textarea,
.ilno-comment > div.text-wrapper > div.text p {
    margin-top: 0.2em;
}
.ilno-comment > div.text-wrapper > div.text p:last-child {
    margin-bottom: 0.2em;
}
.ilno-comment > div.text-wrapper > div.text h1,
.ilno-comment > div.text-wrapper > div.text h2,
.ilno-comment > div.text-wrapper > div.text h3,
.ilno-comment > div.text-wrapper > div.text h4,
.ilno-comment > div.text-wrapper > div.text h5,
.ilno-comment > div.text-wrapper > div.text h6 {

}
.ilno-comment > div.text-wrapper > div.text hr {
    max-width: 100px;
    margin-left: 0;
    border-width: 1px 0 0 0;
}
#ilno-thread a.delete,
#ilno-thread a.reply,
#ilno-thread a.edit,
#ilno-thread a.logout,
#ilno-thread a.ban {
    font-size: 0.80em;
    color: gray !important;
    clear: left;
    padding-left: 12px;
    position: relative;
    font-weight: bold;
}
#ilno-thread a.delete:hover,
#ilno-thread a.reply:hover,
#ilno-thread a.edit:hover,
#ilno-thread a.logout:hover,
#ilno-thread a.ban:hover {
    color: #111111 !important;
    text-shadow: #aaaaaa 0 0 1px !important;
}
#ilno-thread .votes {
    color: gray;
}
#ilno-thread .upvote svg,
#ilno-thread .downvote svg {
    position: relative;
    top: .2em;
}
.ilno-comment .ilno-postbox {
    margin-top: 0.8em;
}
.ilno-postbox {
    margin: 0 auto 2em;
    clear: right;
}
.ilno-postbox form {
    display: block;
    padding: 0;
}
.ilno-postbox .actions-section {
    display: flex;
    justify-content: space-between;
    align-items: center;
}
.ilno-postbox .actions-section .input-wrapper .key {
    font-weight: bold;
}
.ilno-postbox form textarea {
    vertical-align: middle;
    position: relative;
    bottom: 1px;
    margin-left: 0;
}
#ilno-thread textarea:focus,
#ilno-thread input:focus {
    border-color: rgba(0, 0, 0, 0.8);
}
.ilno-postbox .actions-section .input-wrapper {
    display: inline-block;
    position: relative;
    max-width: 75%;
    margin: 0;
}
.ilno-comment div.text-wrapper .textarea-wrapper textarea,
.ilno-postbox .actions-section .input-wrapper input {
    padding: .3em 10px;
    max-width: 133px;
    border-radius: 3px;
    line-height: 1.4em;
    border: 1px solid rgba(0, 0, 0, 0.2);
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
}
ilno-postbox .actions-section .post-action {
    display: inline-block;
    float: right;
    margin: 0 0 0 5px;
}
.ilno-postbox .actions-section .post-action  button {
    padding: calc(.3em - 1px);
    border-radius: 2px;
    border: 1px solid #CCC;
    background-color: #DDD;
    cursor: pointer;
    outline: 0;
    line-height: 1.4em;
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
}
.ilno-postbox .actions-section .post-action button:hover {
    background-color: #CCC;
}
.ilno-postbox .actions-section .post-action button:active {
    background-color: #BBB;
}
@media screen and (max-width: 600px) {
    .ilno-postbox .actions-section .input-wrapper {
        display: block;
        max-width: 100%;
        margin: 0 0 .3em;
    }
    .ilno-postbox .actions-section .input-wrapper input {
        width: 100%;
    }
}
    `
}
