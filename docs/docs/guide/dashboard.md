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

The right sidebar shows a real-time display of all participants currently in the meeting.

From <Badge type="tip" text="v1.2.0" />, each participant’s message count is also displayed.

From <Badge type="tip" text="v1.6.3" />, each participant’s avatar displays a live pulse animation when the user is typing (for non-anonymous messages or comments).

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

### Quick Video

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

## Save as PDF

::: info NOTE

Only available to board owner.

:::

::: tip

If using the live demo, remember to print as PDF. Demo data is deleted after **2 days**.

For long-term usage, consider self-hosting.

:::

To save messages as PDF, use the **Download** button to open the print dialog.

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

::: info NOTE

Only available to the board owner.

:::

Use the **Delete Board** button to permanently remove the board and all its data.

⚠️ This action cannot be undone.
