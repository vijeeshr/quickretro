# Dashboard
This guide gives a quick overview of the dashboard and all its features.

The left side-bar has all the action controls. The ***board creator a.k.a owner*** has more controls than a guest user.

<img src="/dashboard_owner.png" class="shadow-img" alt="Dashboard" width="640" loading="lazy">

The left side-bar for a guest user has fewer controls.

<img src="/dashboard_guest.png" class="shadow-img" alt="Dashboard" width="640" loading="lazy">

The right-sidebar shows a real-time display of all participants who are currently in the meeting.\
From <code>v1.2.0</code> onwards, each participant's message count is also displayed.

::: tip
Hover over the user's Avatar to know the full nickname.
:::

## Share Board

To make guests participate in the meeting, each board has a unique url that needs to be shared with them.\
Use the Share button <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5"
    stroke="currentColor" class="display-icon">
    <path stroke-linecap="round" stroke-linejoin="round"
        d="M7.217 10.907a2.25 2.25 0 1 0 0 2.186m0-2.186c.18.324.283.696.283 1.093s-.103.77-.283 1.093m0-2.186 9.566-5.314m-9.566 7.5 9.566 5.314m0 0a2.25 2.25 0 1 0 3.935 2.186 2.25 2.25 0 0 0-3.935-2.186Zm0-12.814a2.25 2.25 0 1 0 3.933-2.185 2.25 2.25 0 0 0-3.933 2.185Z" />
</svg> on the left-sidebar to share the board url.
::: tip
You can also share the url directly from the browser address-bar.
:::

## Add / Update a message

::: info NOTE
Board owner can update any message.
:::

::: danger BEHAVIOR CHANGE
From <code>v1.2.0</code>, each column has dedicated buttons to create messages. Columns names no longer act as buttons.\
To create a message, click the **+** button.
:::
<img src="/dashboard_add_cards.png" class="shadow-img" alt="Dashboard Add Cards" width="312" loading="lazy">

For all older versions, users can add messages by clicking on the Column name (*Yup! They are buttons*).\
Type in text and press Enter *or* click anywhere else on the page. The message is instantly sent to all.

To Update a message, click on the text and the card becomes updateable.\
Press Enter *or* click anywhere else on the page. The update is instantly sent to all.

### Quick video - For versions prior to <code>v1.2.0</code>
<video class="video-play" controls width="640">
  <source src="/videos/add-update-message.mp4" type="video/webm">
  Your browser does not support the video tag.
</video>

## Anonymous message
Available from <code>v1.2.0</code>

Use the 
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor"
    stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="display-icon">
    <path stroke="none" d="M0 0h24v24H0z" fill="none" />
    <path d="M3 11h18" />
    <path d="M5 11v-4a3 3 0 0 1 3 -3h8a3 3 0 0 1 3 3v4" />
    <path d="M7 17m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0" />
    <path d="M17 17m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0" />
    <path d="M10 17h4" />
</svg>
icon button below each column to create an anonymous message. Your name is not associated with an anonymous message.

## Delete message
Use the delete button 
<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5"
    stroke="currentColor" class="display-icon" >
    <path stroke-linecap="round" stroke-linejoin="round"
        d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" />
</svg>
on each card to delete a messsage

::: info NOTE
Board owner can delete any message.
:::

## Like message
Hit the like button
<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5"
    stroke="currentColor" class="w-6 h-6 display-icon">
    <path stroke-linecap="round" stroke-linejoin="round"
        d="M6.633 10.5c.806 0 1.533-.446 2.031-1.08a9.041 9.041 0 012.861-2.4c.723-.384 1.35-.956 1.653-1.715a4.498 4.498 0 00.322-1.672V3a.75.75 0 01.75-.75A2.25 2.25 0 0116.5 4.5c0 1.152-.26 2.243-.723 3.218-.266.558.107 1.282.725 1.282h3.126c1.026 0 1.945.694 2.054 1.715.045.422.068.85.068 1.285a11.95 11.95 0 01-2.649 7.521c-.388.482-.987.729-1.605.729H13.48c-.483 0-.964-.078-1.423-.23l-3.114-1.04a4.501 4.501 0 00-1.423-.23H5.904M14.25 9h2.25M5.904 18.75c.083.205.173.405.27.602.197.4-.078.898-.523.898h-.908c-.889 0-1.713-.518-1.972-1.368a12 12 0 01-.521-3.507c0-1.553.295-3.036.831-4.398C3.387 10.203 4.167 9.75 5 9.75h1.053c.472 0 .745.556.5.96a8.958 8.958 0 00-1.302 4.665c0 1.194.232 2.333.654 3.375z" />
</svg>
on each card to toggle like.

## Move message across columns
Available from <code>v1.2.0</code>

Use the move icon button
<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5"
    stroke="currentColor" class="display-icon">
    <path stroke-linecap="round" stroke-linejoin="round"
        d="M15.75 9V5.25A2.25 2.25 0 0 0 13.5 3h-6a2.25 2.25 0 0 0-2.25 2.25v13.5A2.25 2.25 0 0 0 7.5 21h6a2.25 2.25 0 0 0 2.25-2.25V15m3 0 3-3m0 0-3-3m3 3H9" />
</svg>
on each card to move cards across columns or categories. 

<img src="/dashboard_move.png" class="shadow-img" alt="Dashboard Move Card" width="312" loading="lazy">

A small menu opens up with the target columns.

## Start / Stop Timer
::: info NOTE
Only Board owner can Start or Stop the timer
:::
The Timer control appear view-only for guests.\
For a Board owner, it appears as a ***button and is clickable***. Enter the minutes or seconds in the popup and start it.\
The coundown stops when it runs its course.\
To Stop the timer before-hand, click the Timer control to open the popup again and Stop it.

### Quick video
<video class="video-play" controls width="640">
  <source src="/videos/start-stop-timer.mp4" type="video/webm">
  Your browser does not support the video tag.
</video>

## Mask Messages
::: info NOTE
Only available for Board owner
:::
Use the Mask <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5"
    stroke="currentColor" class="display-icon">
    <path stroke-linecap="round" stroke-linejoin="round"
        d="M3.98 8.223A10.477 10.477 0 0 0 1.934 12C3.226 16.338 7.244 19.5 12 19.5c.993 0 1.953-.138 2.863-.395M6.228 6.228A10.451 10.451 0 0 1 12 4.5c4.756 0 8.773 3.162 10.065 7.498a10.522 10.522 0 0 1-4.293 5.774M6.228 6.228 3 3m3.228 3.228 3.65 3.65m7.894 7.894L21 21m-3.228-3.228-3.65-3.65m0 0a3 3 0 1 0-4.243-4.243m4.242 4.242L9.88 9.88" />
</svg> Unmask 
<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5"
    stroke="currentColor" class="display-icon">
    <path stroke-linecap="round" stroke-linejoin="round"
        d="M2.036 12.322a1.012 1.012 0 0 1 0-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178Z" />
    <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" />
</svg>
buttons to hide or show messages of other users. The buttons act as a toggle.

## Lock Board
::: info NOTE
Only available for Board owner
:::
When a board is locked, no user will be able to Add new messages or Update existing messages.\
Use the Lock
<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5"
    stroke="currentColor" class="display-icon">
    <path stroke-linecap="round" stroke-linejoin="round"
        d="M16.5 10.5V6.75a4.5 4.5 0 1 0-9 0v3.75m-.75 11.25h10.5a2.25 2.25 0 0 0 2.25-2.25v-6.75a2.25 2.25 0 0 0-2.25-2.25H6.75a2.25 2.25 0 0 0-2.25 2.25v6.75a2.25 2.25 0 0 0 2.25 2.25Z" />
</svg>
Unlock
<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5"
    stroke="currentColor" class="display-icon">
    <path stroke-linecap="round" stroke-linejoin="round"
        d="M13.5 10.5V6.75a4.5 4.5 0 1 1 9 0v3.75M3.75 21.75h10.5a2.25 2.25 0 0 0 2.25-2.25v-6.75a2.25 2.25 0 0 0-2.25-2.25H3.75a2.25 2.25 0 0 0-2.25 2.25v6.75a2.25 2.25 0 0 0 2.25 2.25Z" />
</svg>
buttons to Lock or Unlock a board.

## Download as PDF
::: info NOTE
Only available for Board owner
:::
::: tip
If using the [live demo](https://demo.quickretro.app), remember to download as PDF. The data is deleted in 2 hours.
To retain data for more than 2 hours, it is recommended to self-host.
:::
Use the Download button 
<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5"
    stroke="currentColor" class="display-icon">
    <path stroke-linecap="round" stroke-linejoin="round"
        d="M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25m.75 12 3 3m0 0 3-3m-3 3v-6m-1.5-9H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z" />
</svg>
to download the messages as PDF.

## Dark Theme
Available from <code>v1.1.0</code>

Use the dark
<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" 
    stroke="currentColor" class="display-icon">
    <path stroke-linecap="round" stroke-linejoin="round"
        d="M21.752 15.002A9.72 9.72 0 0 1 18 15.75c-5.385 0-9.75-4.365-9.75-9.75 0-1.33.266-2.597.748-3.752A9.753 9.753 0 0 0 3 11.25C3 16.635 7.365 21 12.75 21a9.753 9.753 0 0 0 9.002-5.998Z" />
</svg> or light
<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" 
    stroke="currentColor" class="display-icon">
    <path stroke-linecap="round" stroke-linejoin="round"
        d="M12 3v2.25m6.364.386-1.591 1.591M21 12h-2.25m-.386 6.364-1.591-1.591M12 18.75V21m-4.773-4.227-1.591 1.591M5.25 12H3m4.227-4.773L5.636 5.636M15.75 12a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0Z" />
</svg> buttons to toggle between the two themes.

## Focussed View
Available from <code>v1.3.0</code>

Use the focus toggle button
<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"
    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
    class="display-icon">
    <circle cx="12" cy="12" r="3" />
    <path d="M3 7V5a2 2 0 0 1 2-2h2" />
    <path d="M17 3h2a2 2 0 0 1 2 2v2" />
    <path d="M21 17v2a2 2 0 0 1-2 2h-2" />
    <path d="M7 21H5a2 2 0 0 1-2-2v-2" />
</svg> to focus on cards by user.\
A small navigation panel opens at the top to help you quickly navigate by each user's cards.
<img src="/dashboard_focus_panel.png" class="shadow-img" alt="Dashboard Focus Panel" width="398" height="97" loading="lazy">
::: tip
Clicking on the Avatar displayed in each card also starts focussed view
:::

## Multi-Language support
Available from <code>v1.3.0</code>

Use the 
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor"
    stroke-width="2" stroke-linecap="round" stroke-linejoin="round" @click="openLanguageDialog"
    class="display-icon">
    <circle cx="12" cy="12" r="10" />
    <path d="M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20" />
    <path d="M2 12h20" />
</svg> button to change the current language.