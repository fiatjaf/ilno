export default {
  inline: `
#ilno-thread * {
    -webkit-box-sizing: border-box;
    -moz-box-sizing: border-box;
    box-sizing: border-box;
}
#ilno-thread .ilno-comment-header a {
    text-decoration: none;
    cursor: pointer;
}

#ilno-thread {
    padding: 0;
    margin: 0;
}
#ilno-thread > h4 {
    color: #555;
    font-weight: bold;
}
#ilno-thread > .ilno-feedlink {
    float: right;
    padding-left: 1em;
}
#ilno-thread > .ilno-feedlink > a {
    font-size: 0.8em;
    vertical-align: bottom;
}
#ilno-thread textarea {
    min-height: 58px;
    min-width: 100%;
    max-width: 100%;
    outline: 0;
}

#ilno-root .ilno-comment {
    max-width: 68em;
    padding-top: 0.95em;
    margin: 0.95em auto;
}
#ilno-root .ilno-comment:not(:first-of-type),
.ilno-follow-up .ilno-comment {
    border-top: 1px solid rgba(0, 0, 0, 0.1);
}
.ilno-comment > div.avatar {
    display: block;
    float: left;
    width: 7%;
    margin: 3px 15px 0 0;
}
.ilno-comment > div.avatar > svg {
    max-width: 48px;
    max-height: 48px;
    border: 1px solid rgba(0, 0, 0, 0.2);
    border-radius: 3px;
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
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
.ilno-comment > div.text-wrapper > .ilno-comment-header .spacer {
    padding: 0 6px;
}
.ilno-comment > div.text-wrapper > .ilno-comment-header .spacer,
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
    font-size: 130%;
    font-weight: bold;
}
.ilno-comment > div.text-wrapper > div.textarea-wrapper textarea,
.ilno-comment > div.text-wrapper > .ilno-comment-footer {
    font-size: 0.80em;
    color: gray !important;
    clear: left;
}
.ilno-feedlink,
.ilno-comment > div.text-wrapper > .ilno-comment-footer a {
    font-weight: bold;
    text-decoration: none;
    cursor: pointer;
}
.ilno-feedlink:hover,
.ilno-comment > div.text-wrapper > .ilno-comment-footer a:hover {
    color: #111111 !important;
    text-shadow: #aaaaaa 0 0 1px !important;
}
.ilno-comment > div.text-wrapper > .ilno-comment-footer > a {
    position: relative;
    top: .2em;
}
.ilno-comment > div.text-wrapper > .ilno-comment-footer > a + a {
    padding-left: 1em;
}
.ilno-comment > div.text-wrapper > .ilno-comment-footer .votes {
    color: gray;
}
.ilno-comment > div.text-wrapper > .ilno-comment-footer .upvote svg,
.ilno-comment > div.text-wrapper > .ilno-comment-footer .downvote svg {
    position: relative;
    top: .2em;
}
.ilno-comment .ilno-postbox {
    margin-top: 0.8em;
}
.ilno-comment.ilno-no-votes > * > .ilno-comment-footer span.votes {
    display: none;
}

.ilno-postbox {
    max-width: 68em;
    margin: 0 auto 2em;
    clear: right;
}
.ilno-postbox > form {
    display: block;
    padding: 0;
}
.ilno-postbox > form > .auth-section,
.ilno-postbox > form > .auth-section .post-action {
    display: block;
}
.ilno-postbox > form textarea,
.ilno-postbox > form input[type=checkbox] {
    vertical-align: middle;
    position: relative;
    bottom: 1px;
    margin-left: 0;
}
.ilno-postbox > form .notification-section {
    font-size: 0.90em;
    padding-top: .3em;
}
#ilno-thread textarea:focus,
#ilno-thread input:focus {
    border-color: rgba(0, 0, 0, 0.8);
}
.ilno-postbox > form > .auth-section .input-wrapper {
    display: inline-block;
    position: relative;
    max-width: 25%;
    margin: 0;
}
.ilno-postbox > form > .auth-section .input-wrapper input {
    padding: .3em 10px;
    max-width: 100%;
    border-radius: 3px;
    background-color: #fff;
    line-height: 1.4em;
    border: 1px solid rgba(0, 0, 0, 0.2);
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
}
.ilno-postbox > form > .auth-section .post-action {
    display: inline-block;
    float: right;
    margin: 0 0 0 5px;
}
.ilno-postbox > form > .auth-section .post-action > input {
    padding: calc(.3em - 1px);
    border-radius: 2px;
    border: 1px solid #CCC;
    background-color: #DDD;
    cursor: pointer;
    outline: 0;
    line-height: 1.4em;
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
}
.ilno-postbox > form > .auth-section .post-action > input:hover {
    background-color: #CCC;
}
.ilno-postbox > form > .auth-section .post-action > input:active {
    background-color: #BBB;
}
.ilno-postbox > form > .notification-section {
    display: none;
    padding-bottom: 10px;
}
@media screen and (max-width:600px) {
    .ilno-postbox > form > .auth-section .input-wrapper {
        display: block;
        max-width: 100%;
        margin: 0 0 .3em;
    }
    .ilno-postbox > form > .auth-section .input-wrapper input {
        width: 100%;
    }
}
    `
}
