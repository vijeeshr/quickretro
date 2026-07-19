---
title: Dashboard
description: Learn how to use the QuickRetro dashboard to manage sprint retrospectives, moderate discussions, lock boards, and track team feedback in real time.
outline: deep
head:
  - - meta
    - name: keywords
      content: retrospective dashboard, remote team meeting, agile board features
---

# Dashboard

This guide gives a quick overview of the dashboard and all its features.

The left sidebar has all the action controls. The **_board creator (owner)_** has more controls than a guest user.

<img
  src="/dashboard_owner.png"
  class="shadow-img"
  alt="Dashboard (Owner)"
  width="640"
  loading="lazy"
/>

The left sidebar for a guest user has fewer controls.

<img
  src="/dashboard_guest.png"
  class="shadow-img"
  alt="Dashboard (Guest)"
  width="640"
  loading="lazy"
/>

The right sidebar shows a real-time display of all participants in the meeting.

From <Badge type="tip" text="v1.2.0" />, each participant’s message count is also displayed.

From <Badge type="tip" text="v1.6.3" />, each participant’s avatar displays a live pulse animation when the user is typing (for non-anonymous messages or comments).

From <Badge type="tip" text="v1.7.0" />, board owner's avatar will show a small star icon ⭐, and inactive participant's avatar will appear grayed out.

::: tip

Hover over a user’s avatar to see the full nickname.

:::

## Share Board

Each board has a unique URL that must be shared with guests to participate.

Use the **Share** button in the left sidebar to copy the board URL.

::: tip

You can also share the URL directly from the browser address bar.

:::

## Customize Columns

Available from <Badge type="tip" text="v1.6.0" />

<img
  src="/dashboard_edit_columns.png"
  class="shadow-img"
  alt="Edit Columns"
  width="312"
  loading="lazy"
/>

Board owners can click a column header to:

- Rename columns
- Enable or disable columns
- Reorder columns

::: info NOTE

Only available to the board owner.

:::

## Add / Update a Message

<img
  src="/dashboard_add_cards.png"
  class="shadow-img"
  alt="Add Cards"
  width="312"
  loading="lazy"
/>

Each column has buttons to create new messages

- Type text and press **Enter**, or click anywhere else
- Updates are broadcast instantly

To update an existing message, click its text, edit it, and press **Enter** or click outside.

## Anonymous Message

Available from <Badge type="tip" text="v1.2.0" />

Use the **Anonymous Message** button below each column.

Your name is not associated with anonymous messages.

## Delete Message

Use the **Delete** button on each card.

::: info NOTE

The board owner can delete any message.

:::

## Like Message

Use the **Like** button on a card to toggle likes.

### Offline Likes

Available from <Badge type="tip" text="v1.9.0" />

Offline Likes allow the board owner to record likes/votes on behalf of participants who are physically present in the room and not connected to the board.

<img
  src="/offline-likes-panel.png"
  class="shadow-img"
  alt="Offline likes panel"
  width="270"
  loading="lazy"
/>

This is not enabled by default. To turn it on, open the "Options" sub-menu in left sidebar and use the Offline likes toggle button.

<img
  src="/options-submenu.png"
  class="shadow-img"
  alt="Options submenu"
  width="200"
  loading="lazy"
/>

Once enabled, a small chevron icon appears below the "like" icon on each card. Use it to open the offline likes recording panel.

::: tip

Previously recorded **offline likes are retained even if the feature is later turned off**.

Update `panel_enabled` to `true` in config.toml to **enable offline likes by default**.
More details in [offline like configuration](configurations#offline-likes).

:::

::: info NOTE

Only board owner can update offline likes.

:::

## Pin Messages to top

Available from <Badge type="tip" text="v1.9.2" />

Board owner can pin messages to top of the board. The messages appear at the top of their columns for all users in the board.

Pinned messages **aways appear before other messages** even in Pdf prints, exported Json documents, and when [sorted by likes/comments](#sort-by-most-likes-comments).

<img
  src="/dashboard_pinned_message.png"
  class="shadow-img"
  alt="Pin Card"
  width="312"
  loading="lazy"
/>

::: info NOTE

Only board owner can pin/unpin messages.

:::

## Move Message Across Columns

Available from <Badge type="tip" text="v1.2.0" />

Use the **Move** button on a card to move it across columns or categories.

<img
  src="/dashboard_move.png"
  class="shadow-img"
  alt="Move Card"
  width="312"
  loading="lazy"
/>

A menu opens with the target columns.

## Comments

Available from <Badge type="tip" text="v1.5.4" />

<img
  src="/comments.png"
  class="shadow-img"
  alt="Comments"
  width="312"
  loading="lazy"
/>

Use the **Comments** button on a card to open or close the comments pane.

The comments pane is closed by default.

### Add / Edit / Delete Comments

- Type your comment and press **Enter** or click outside to save
- Click an existing comment to edit it
- Use the delete icon to remove a comment

::: info NOTE

The board owner can delete any comment.

:::

## Start / Stop Timer

::: info NOTE

Only the board owner can start or stop the timer.

:::

- Guests can only view the timer
- Owners can enter minutes or seconds and start it
- The countdown stops automatically
- To stop early, reopen the timer and click **Stop**

From <Badge type="tip" text="v1.8.0" />, a sub-menu with preset shortcuts has been introduced for quickly starting and stopping the timer.

<img
  src="/timer-submenu.png"
  class="shadow-img"
  alt="Timer presets submenu"
  width="312"
  loading="lazy"
/>

### Quick Video (older versions)

<video class="video-play" controls width="640">
  <source src="/videos/start-stop-timer.mp4" type="video/webm" />
  Your browser does not support the video tag.
</video>

## Mask Messages

::: info NOTE

Only available to the board owner.

:::

Use the **Mask / Unmask** buttons to hide or show other users’ messages.

These buttons act as a toggle.

## Lock Board

::: info NOTE

Only available to the board owner.

:::

When a board is locked:

- Users cannot add new messages
- Users cannot update existing messages

Use the **Lock / Unlock** buttons to control access.

::: warning

Any unsaved message edits will be lost when the board is locked.

:::

## Change Ownership

Available from <Badge type="tip" text="v1.7.0" />

The board owner can transfer ownership to any participant.

Click any participant in the right-sidebar, or the board ownership transfer icon in the left-sidebar (_only visible when there is atleast one other participant in the board_) to start the process.

The board owner **looses ownership** when the transfer is complete.

### Reclaim Ownership

Only the **original board creator** can reclaim ownership.

The board creator will be shown a "reclaim ownership" (_yellow colored_) icon in left sidebar to forcefully reclaim ownership.

## Save as PDF

::: info NOTE

Only available to board owner.

:::

::: tip

If using the live demo, remember to print as PDF. Demo data is deleted after **2 days**.

For long-term usage, consider self-hosting.

:::

To save messages as PDF, use the **Download** button to open the print dialog.

From <Badge type="tip" text="v1.8.0" />, the print button now features a sub-menu allowing you to optionally include **comments** and/or message author **names** in your printed output.

Click the **>>** symbol below the icon to open the sub-menu.

Names are never printed for anonymous cards.

<img
  src="/print-submenu.png"
  class="shadow-img"
  alt="Print options submenu"
  width="200"
  loading="lazy"
/>

## Save as JSON

::: info NOTE

Only available to board owner.

:::

Available from <Badge type="tip" text="v1.9.0" />

<img
  src="/options-submenu.png"
  class="shadow-img"
  alt="Options submenu"
  width="200"
  loading="lazy"
/>

Use the "Options" submenu to download the board as JSON.

## Sort by most Likes/Comments

Available from <Badge type="tip" text="v1.8.0" />

The 3-way toggle button in the left sidebar allows you to instantly sort cards to the top based on likes or comments locally.

<img
  src="/sort-likes-comments.png"
  class="shadow-img"
  alt="Sort by comments or likes"
  width="200"
  loading="lazy"
/>

## Dark Theme

Available from <Badge type="tip" text="v1.1.0" />

Use the **Dark / Light** toggle to switch themes.

## Focused View

Available from <Badge type="tip" text="v1.3.0" />

Use the **Focus** toggle to view cards by user.

A navigation panel appears at the top for quick switching.

<img
  src="/dashboard_focus_panel.png"
  class="shadow-img"
  alt="Focus Panel"
  width="398"
  height="97"
  loading="lazy"
/>

::: tip

Clicking a user’s avatar on a card also activates Focused View.

:::

## Multi-Language Support

Available from <Badge type="tip" text="v1.3.0" />

Use the **Language** button to change the current language.

## Delete Board Manually

Available from <Badge type="tip" text="v1.5.3" />

From <Badge type="warning" text="v1.9.0" />, this button has been moved to "Options" sub-menu.

Use the **Delete Board** button to permanently remove the board and all its data.

⚠️ This action cannot be undone.

::: info NOTE

Only available to the board owner.

:::

## Frequently Asked Questions

### Why are messages of other users' appearing blurred?

To prevent bias during brainstorming, other users' cards are blurred by default. Board owner can reveal or blur messages anytime.

### Who can lock the board or delete whole board?

Only the board owner can perform these actions.

### Can anonymous message creators be revealed later?

No. Anonymous message / card creators will always remain anonymous.

### Can I post anonymous comments?

No. Anonymous comments aren't yet supported. Only anonymous messages / cards can be added.

### How to enter likes/votes for physical participants?

Use the [Offline Likes](#offline-likes) feature to record likes/votes on behalf of participants who are physically present in the meeting and not connected to QuickRetro.
