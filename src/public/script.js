// Initialize
let socket
let applyMasking = true
let isOwner = false

// Get board/group from url
const getBoardName = () => {
  const url = new URL(document.URL)
  const pathVars = url.pathname.split('/')
  return pathVars.length >= 2 && pathVars[pathVars.length - 2] == "board" ? pathVars[pathVars.length - 1] : ""
}

// Ensure a Userid is assigned
const ensureUser = () => {
  if (!localStorage.getItem("user")) {
    localStorage.setItem("user", crypto.randomUUID())
  }
  if (!localStorage.getItem("xid")) {
    localStorage.setItem("xid", crypto.randomUUID())
  }  
  if (!localStorage.getItem("nickname")) {
    window.location.replace(`/board/${getBoardName()}/join`)
  }    
  return { user: localStorage.getItem("user"), nickname: localStorage.getItem("nickname"), xid: localStorage.getItem("xid") }
}

const userDetails = ensureUser()
const user = userDetails.user
const nickname = userDetails.nickname
const externalId = userDetails.xid
const group = getBoardName()

if (!group) {
  console.error("Could not get board name")
}

class Event {
  constructor(type, payload) {
    this.typ = type
    this.pyl = payload
  }
}
class RegisterEvent {
  constructor(by, nickname, xid, group) {
    this.by = by
    this.nickname = nickname
    this.xid = xid
    this.grp = group
  }
}
class MaskEvent {
  constructor(by, group, mask) {
    this.by = by
    this.grp = group
    this.mask = mask
  }
}
class SaveMessageEvent {
  constructor(id, by, nickname, group, content, category) {
    this.id = id
    this.by = by
    this.nickname = nickname
    this.grp = group
    this.msg = content
    this.cat = category
  }
}
class LikeMessageEvent {
  constructor(messageId, by, like) {
    this.msgId = messageId
    this.by = by
    this.like = like
  }
}
class DeleteMessageEvent {
  constructor(messageId, by, group) {
    this.msgId = messageId
    this.by = by
    this.grp = group
  }
}
// Response models
class RegisterResponse {
  constructor(typ, boardName, boardTeam, boardStatus, boardMasking, isBoardOwner, mine, users) {
    this.typ = typ
    this.boardName = boardName
    this.boardTeam = boardTeam
    this.boardStatus = boardStatus
    this.boardMasking = boardMasking
    this.isBoardOwner = isBoardOwner
    this.mine = mine
    this.users = users // fields- nickname, xid.
  }
}
class UserClosingResponse {
  constructor(typ, users) {
    this.typ = typ
    this.users = users // fields- nickname, xid.
  }
}
class MessageResponse {
    constructor(typ, id, nickname, msg, cat, likes, liked, mine) {
        this.typ = typ
        this.id = id
        this.nickname = nickname
        this.msg = msg
        this.cat = cat
        this.likes = likes
        this.liked = liked
        this.mine = mine
    }
}
class DeleteResponse {
    constructor(typ, id) {
        this.typ = typ
        this.id = id
    }
}
class LikeResponse {
    constructor(typ, id, likes, liked) {
        this.typ = typ
        this.id = id
        this.likes = likes
        this.liked = liked
    }
}
class MaskResponse {
  constructor(typ, mask) {
      this.typ = typ
      this.mask = mask
  }
}

const dispatchEvent = (eventType, eventPayload) => {
  let event = new Event(eventType, eventPayload)
  console.log("dispatch", event)
  if (socket.readyState == 1) {
    socket.send(JSON.stringify(event)); // Can throw error if socket object is "connecting". Check the docs.
  } else {
    console.log('Socket not ready for send operation')
  }
}

// Connect
const connect = () => {
  if (!window["WebSocket"]) {
    console.log("Websocket not supported. Cannot continue.")
    return
  }

  const url = `ws://${document.location.host}/ws/board/${group}/user/${user}/meet`
  // const url = `ws://localhost:8080/ws/board/${group}/user/${user}/meet`
  socket = new WebSocket(url)

  // Socket event handlers
  socket.onopen = (event) => {
    console.log("[open] Connection established", event);
    // Dispatching user's RegisterEvent immediately to server.
    dispatchEvent("reg", new RegisterEvent(user, nickname, externalId, group))
  }

  socket.onmessage = (event) => {
    const message = JSON.parse(event.data)
    let receivedMessage = {}

    switch (message.typ) {
      case "reg":
        receivedMessage = Object.assign(new RegisterResponse, message)
        break         
      case "mask":
        receivedMessage = Object.assign(new MaskResponse, message)
        break;         
      case "msg":
        receivedMessage = Object.assign(new MessageResponse, message)
        break
      case "del":
        receivedMessage = Object.assign(new DeleteResponse, message)
        break
      case "like":
        receivedMessage = Object.assign(new LikeResponse, message)
        break;
      case "closing":
        receivedMessage = Object.assign(new UserClosingResponse, message)
        break;                
      default:
        console.log("Unknown response type")
        break;
    }
    console.log(message, receivedMessage)
    
    // Update dom
    if (receivedMessage instanceof RegisterResponse) {
      applyMasking = receivedMessage.boardMasking
      isOwner = receivedMessage.isBoardOwner
      setMaskIcon()
      renderOnlinePresenceSidebar(receivedMessage.users)
      if (receivedMessage.mine) {
        // Try reloading messages. Useful if user reloads the page, which triggers new socket connection.
        reload() // TODO: Maybe this can be conditional i.e. if the RegisterResponse has some prop that says reload isn't required (e.g. board has no msgs or board hasn't started etc). RegisterResponse.skipReload.
      }
    }

    if (receivedMessage instanceof UserClosingResponse) {
      renderOnlinePresenceSidebar(receivedMessage.users)
    }

    if (receivedMessage instanceof MaskResponse) {
      applyMasking = receivedMessage.mask
      setMaskIcon()
      EnforceMaskingForExistingCards()
    }

    if (receivedMessage instanceof MessageResponse) {
      appendMessageToDOM(receivedMessage)
    }

    if (receivedMessage instanceof LikeResponse) {
      const card = document.querySelector(`[data-card="${receivedMessage.id}"]`)
      if (!card) { return }
      const likesEl = card.querySelector(`[data-likes]`)
      if (likesEl) {
        likesEl.innerText = receivedMessage.likes
        likesEl.dataset.likes = receivedMessage.likes
      }
      const likedEl = card.querySelector(`[data-liked]`)
      if (likedEl) {
        likedEl.dataset.liked = receivedMessage.liked
      }
    }

    if (receivedMessage instanceof DeleteResponse) {
      const card = document.querySelector(`[data-card="${receivedMessage.id}"]`)
      if (!card) { return }
      card.remove()
    }

  }

  socket.onclose = (event) => {
    console.log("Close received", event)
    // event.code === 1000
    // event.reason === "Work complete"
    // event.wasClean === true (clean close)
  }

  socket.onerror = (error) => {
    console.log("Error", error);
  };
}
connect()


const EnforceMaskingForExistingCards = () => {
  // Mask/Unmask exisitng cards in DOM
  // If selector has no space, it returns elem with both atrributes. With space means child.
  maskableCards = document.querySelectorAll(`[data-card][data-mine='false']`)
  for (const card of maskableCards)
  {
    const blurEl = card.querySelector(`[data-mine='false']`)
    const contentEl = card.querySelector(`[data-content-id="${card.dataset.card}"]`)
    if (contentEl) {
      if (applyMasking) {
        // mask
        contentEl.innerText = bytesToBase64(new TextEncoder().encode(contentEl.innerText))
        blurEl.classList.add(`data-[mine=false]:blur-sm`)
      } else {
        // unmask
        contentEl.innerText = new TextDecoder().decode(base64ToBytes(contentEl.innerText))
        blurEl.classList.remove(`data-[mine=false]:blur-sm`)
      }
    }
  }
}

const newCard = (message, isFirstDraft) => {
  const avatarText = createAvatarText(message.nickname)
  const avatarColor = createAvatarColor(message.nickname)
  let maskMessage = applyMasking == true && !message.mine ? bytesToBase64(new TextEncoder().encode(message.msg)) : message.msg
  let maskClass = applyMasking == true && !message.mine ? `data-[mine=false]:blur-sm` : ``
  return `
  <div class="bg-white rounded-lg p-3 mb-2 shadow-xl" data-card="${message.id}" data-mine="${message.mine}"> 
    <div class="text-gray-500 pb-2 ${maskClass}" data-mine="${message.mine}">
      <article class="min-h-4 cursor-default text-center break-words focus:outline-none" 
        onclick="edit('${message.id}')" onkeydown="saveOnEnter(event)" onblur="save(event)"
        data-content-id="${message.id}">${maskMessage}</article>
    </div>
    <div class="flex items-center text-gray-500 pt-2 data-[firstdraft=true]:invisible" data-firstdraft="${isFirstDraft}">
      <div class="flex mr-2">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6 cursor-pointer data-[liked=true]:text-blue-500" onclick="likeItem(this)" data-liked="${message.liked}"><path stroke-linecap="round" stroke-linejoin="round" d="M6.633 10.5c.806 0 1.533-.446 2.031-1.08a9.041 9.041 0 012.861-2.4c.723-.384 1.35-.956 1.653-1.715a4.498 4.498 0 00.322-1.672V3a.75.75 0 01.75-.75A2.25 2.25 0 0116.5 4.5c0 1.152-.26 2.243-.723 3.218-.266.558.107 1.282.725 1.282h3.126c1.026 0 1.945.694 2.054 1.715.045.422.068.85.068 1.285a11.95 11.95 0 01-2.649 7.521c-.388.482-.987.729-1.605.729H13.48c-.483 0-.964-.078-1.423-.23l-3.114-1.04a4.501 4.501 0 00-1.423-.23H5.904M14.25 9h2.25M5.904 18.75c.083.205.173.405.27.602.197.4-.078.898-.523.898h-.908c-.889 0-1.713-.518-1.972-1.368a12 12 0 01-.521-3.507c0-1.553.295-3.036.831-4.398C3.387 10.203 4.167 9.75 5 9.75h1.053c.472 0 .745.556.5.96a8.958 8.958 0 00-1.302 4.665c0 1.194.232 2.333.654 3.375z" /></svg>
        <span class="data-[likes='0']:invisible cursor-default" data-likes="${message.likes}">${message.likes}</span>
      </div>
      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6 cursor-pointer data-[mine=false]:invisible" onclick="deleteItem(this)" data-mine="${message.mine}"><path stroke-linecap="round" stroke-linejoin="round" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" /></svg>
      <div class="inline-flex items-center justify-center w-6 h-6 overflow-hidden rounded-full ml-auto" style="background-color:${avatarColor};" data-avatar>
        <span class="font-medium text-xs cursor-default text-white" data-avatar-text>${avatarText}</span>
      </div>
    </div>
  </div>
  `
}

const updateCard = (cardEl, receivedMessage) => {
    // Card already present in DOM. Update parts of the card.
    const a = cardEl.querySelector(`[data-content-id="${receivedMessage.id}"]`)
    a.innerText = applyMasking == true && !receivedMessage.mine ? bytesToBase64(new TextEncoder().encode(receivedMessage.msg)) : receivedMessage.msg
    // Reset "firstdraft". "firstdraft" is only set when new message is added in DOM(before sending to server). 
    // At this point, response is already back from server. Its no more a "firstdraft" here.
    document.querySelector(`[data-card="${receivedMessage.id}"] [data-firstdraft]`).dataset.firstdraft = false
}

const deleteItem = (itemEl) => {
  const el = itemEl.closest('[data-card]')
  if (el) {
    dispatchEvent("del", new DeleteMessageEvent(el.dataset.card, user, group))
  }
}

const likeItem = (itemEl) => {
  const el = itemEl.closest(`[data-card]`)
  let like = el.querySelector(`[data-liked]`).dataset.liked === "true" ? false : true // Make "like" behave like a toggle
  if (el) {
    dispatchEvent("like", new LikeMessageEvent(el.dataset.card, user, like))
  }
}

const mask = () => {
  dispatchEvent("mask", new MaskEvent(user, group, true))
}

const unmask = () => {
  dispatchEvent("mask", new MaskEvent(user, group, false))
}

const add = (el) => {
  const categoryEl = el.closest('[data-category]')
  const id = crypto.randomUUID()
  categoryEl.insertAdjacentHTML("beforeend", newCard({ id: id, nickname: nickname, msg: "", mine: true, likes: "0", liked: false}, true))
  edit(id)
}

const edit = (id) => {
  const card = document.querySelector(`[data-card="${id}"]`)
  const p = document.querySelector(`[data-content-id="${id}"]`)
  // User can only edit own message
  if (card.dataset.mine === "true" && p.contentEditable !== "true") {
    p.contentEditable = "true"
    p.classList.replace("cursor-default", "cursor-auto")
    card.classList.add("border", "border-sky-400")
    p.focus()
  }
}

const saveOnEnter = (e) => {
  if (e.which === 13 && !e.shiftKey) {
    save(e)
    e.preventDefault() // In a textarea, this stops addition of a new line. Just adding comment for information.
  }
}

const save = (e) => {
  const p = e.target
  if (p.contentEditable === "true") {
    // Make section read-only and remove border highlighting
    p.contentEditable = "false"
    p.classList.replace("cursor-auto", "cursor-default")
    const card = p.closest("[data-card]")
    card.classList.remove("border", "border-sky-400")
    // Attempt sending data
    const category = p.closest('[data-category]').dataset.category
    dispatchEvent("msg", new SaveMessageEvent(p.dataset.contentId, user, nickname, group, p.innerText, category)) // Called function can throw exception when connection is not in ready state. e.g. Try send a large message that errors. 
    //card.querySelector(`[data-firstdraft]`).dataset.firstdraft = false // setting to false since save attempt has already been made.
  }
}

const reload = () => {
  // Specify the API endpoint for user data
  // const apiUrl = `http://localhost:8080/api/board/${group}/user/${user}/refresh`
  const apiUrl = `http://${document.location.host}/api/board/${group}/user/${user}/refresh`

  // Make a GET request using the Fetch API
  fetch(apiUrl)
    .then(response => {
      if (!response.ok) {
        throw new Error('Network response was not ok');
      }
      return response.json();
    })
    .then(data => {
      let receivedMessages =  data.map(m => Object.assign(new MessageResponse, m))
      console.log('Refreshed data:', receivedMessages);
      // ToDo: Delete orphaned messages in DOM?
      for (let rm of receivedMessages) {
        appendMessageToDOM(rm)
      }
    })
    .catch(error => {
      console.error('Error:', error);
    });
}

const appendMessageToDOM = (receivedMessage) => {
  const card = document.querySelector(`[data-card="${receivedMessage.id}"]`)
  if (card) {
    // Card already present in DOM. Update parts of the card.
    updateCard(card, receivedMessage)
  } else {
    // Add new card to DOM
    const el = document.querySelector(`[data-category="${receivedMessage.cat}"]`)
    el.insertAdjacentHTML("beforeend", newCard(receivedMessage), false) // Todo: Escape user input .i.e the arg to itemTemplate()
  }
}

function createAvatarText(str) {
  if (str) {
    let n = str.trim().split(/\s+/)
    if (n && n.length >= 2) return `${n[0][0].toUpperCase()}${n[n.length-1][0].toUpperCase()}`
    if (n && n.length == 1) return `${n[0][0].toUpperCase()}`
  }
  return ''
}

function createAvatarColor(str) {
  if (str) {
    let hash = 0
    const saturation = 50, lightness = 60
    for (let i = 0; i < str.length; i++) {
      hash = str.charCodeAt(i) + ((hash << 5) - hash)
    }
    let h = hash % 360
    return 'hsl('+h+','+saturation+'%,'+lightness+'%)'
  }
  return 'hsl(0,0,100)'
}

function base64ToBytes(base64) {
  const binString = atob(base64);
  return Uint8Array.from(binString, (m) => m.codePointAt(0));
}
function bytesToBase64(bytes) {
  const binString = String.fromCodePoint(...bytes);
  return btoa(binString);
}
// bytesToBase64(new TextEncoder().encode("a Ā 𐀀 文 🦄")); // "YSDEgCDwkICAIOaWhyDwn6aE"
// new TextDecoder().decode(base64ToBytes("YSDEgCDwkICAIOaWhyDwn6aE")); // "a Ā 𐀀 文 🦄"

const setMaskIcon = () => {
  if (!isOwner) {
    return
  }

  let maskIcon = document.querySelector(`[data-icon-mask]`)
  let unmaskIcon = document.querySelector(`[data-icon-unmask]`)

  if (applyMasking) {
    maskIcon.classList.add('hidden')
    unmaskIcon.classList.remove('hidden')
  } else {
    maskIcon.classList.remove('hidden')
    unmaskIcon.classList.add('hidden')
  }
}

function renderOnlinePresenceSidebar(users) {
  let rightSb = document.querySelector(`[data-right-sidebar]`)
  // clear all existing avatars and re-render.
  while (rightSb.hasChildNodes()) {
    rightSb.removeChild(rightSb.firstChild)
  }
  // rerender all
  if (users && users.length > 0) {
    for (let u of users) {
      const avatarText = createAvatarText(u.nickname)
      const avatarColor = createAvatarColor(u.nickname)
      let newPresenceAvatar = 
      `
      <div class="inline-flex items-center justify-center w-8 h-8 overflow-hidden rounded-full ml-auto mx-auto mb-4" title="${u.nickname}" style="background-color:${avatarColor};" data-presence-for="${u.xid}">
        <span class="font-medium text-xs cursor-default text-white" data-presence-for-text>${avatarText}</span>
      </div>
      `
      rightSb.insertAdjacentHTML("beforeend", newPresenceAvatar)
    }
  } 
}

// Note: This can be moved to setAvatar() and call on RegisterResponse, like setMaskIcon. Just for consistency. But it will be executed late. ?
document.addEventListener("DOMContentLoaded", (event) => {
  // Prepare sidebar Avatar
  let avatar = document.querySelector(`[data-icon-avatar]`)
  avatar.style.backgroundColor = createAvatarColor(nickname)
  avatar.querySelector(`[data-icon-avatar-text]`).innerText = createAvatarText(nickname)
})