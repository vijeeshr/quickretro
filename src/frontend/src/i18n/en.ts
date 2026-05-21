export default {
  common: {
    anonymous: 'Anonymous',
    minutes: 'Minutes',
    seconds: 'Seconds',
    start: 'Start',
    stop: 'Stop',
    copy: 'Copy',
    board: 'Board',
    toolTips: {
      darkTheme: 'Turn on dark theme',
      lightTheme: 'Turn on light theme',
    },
    contentOverloadError: 'Content more than allowed limit.',
    contentStrippingError: 'Content more than allowed limit. Extra text is stripped from the end.',
    invalidColumnSelection: 'Please select column(s)',
    typing: '{name} is typing',
    share: {
      linkCopied: 'Link copied!',
      linkCopyError: 'Failed to copy. Please copy directly.',
    },
    customColumnSetup: {
      shareLabel: 'Share column setup',
      shareHelp:
        'Copy the link below to share your custom column setup with others, or bookmark it for later.',
      applied: 'Custom column setup applied',
    },
  },
  join: {
    label: 'Join as guest',
    namePlaceholder: 'Type your name here!',
    nameRequired: 'Please enter your name',
    button: 'Join',
  },
  createBoard: {
    label: 'Create Board',
    namePlaceholder: 'Type board name here!',
    nameRequired: 'Please enter board name',
    teamNamePlaceholder: 'Type team name here!',
    button: 'Create',
    buttonProgress: 'Creating..',
    captchaInfo: 'Please complete the CAPTCHA to continue',
    boardCreationError: 'Error when creating board',
    columns: 'Columns',
  },
  dashboard: {
    timer: {
      oneMinuteLeft: 'One minute left for countdown',
      timeCompleted: "Hey! You've run out of time",
      title: 'Start/Stop Timer',
      helpTip:
        'Adjust minutes and seconds using the + and - controls, or the Up and Down arrows on keyboard. Max allowed is 1 hour.',
      invalid:
        'Please enter valid minutes/seconds values. Allowed range is 1 second to 60 minutes.',
      tooltip: 'Countdown Timer',
      presets: 'Presets',
    },
    share: {
      title: 'Copy and share below url to participants',
      toolTip: 'Share board with others',
    },
    mask: {
      maskTooltip: 'Mask messages',
      unmaskTooltip: 'Unmask messages',
    },
    lock: {
      lockTooltip: 'Lock board',
      unlockTooltip: 'Unlock board',
      message: 'Board locked. No changes allowed.',
      discardChanges: 'Board locked! Unsaved messages discarded',
      unlockButton: 'Unlock',
    },
    spotlight: {
      noCardsToFocus: 'There are no cards to focus',
      tooltip: 'Focus cards',
    },
    print: {
      tooltip: 'Print',
    },
    language: {
      tooltip: 'Change language',
    },
    delete: {
      title: 'Confirm deletion',
      text: 'Data cannot be recovered after its deleted. Are you sure to proceed?',
      tooltip: 'Delete this Board',
      continueDelete: 'Yes',
      cancelDelete: 'No',
    },
    columns: {
      col01: 'What went well',
      col02: 'Challenges',
      col03: 'Action Items',
      col04: 'Appreciations',
      col05: 'Improvements',
      cannotDisable: 'Cannot disable column(s) with cards',
      update: 'Update',
      discardNewMessages: 'Your draft was discarded because the column was disabled.',
    },
    printFooter: 'Created with',
    offline: 'You seem to be offline.',
    autoDeleteSchedule: 'This board will be cleaned up automatically on {date}.',
    welcome: {
      title: 'Welcome!',
      maskInfo:
        'Messages are masked by default. Only the author can see their own messages until you unmask.',
      maskOnLabel: 'Card Masking is ON',
      maskOffLabel: 'Card Masking is OFF',
      ok: 'Got it!',
    },
    notFound: {
      title: 'Board not found',
      text: 'The board you are looking for was either auto-deleted, or manually deleted by its owner.',
      createNewBoard: 'Create new board',
      supportText: 'Support us with a star on GitHub! ⭐',
    },
  },
  transferOwnership: {
    tooltip: 'Transfer board ownership',
    promotedNotification: 'You are now the board owner. You can manage settings and participants.',
    demotedNotification: 'Ownership transferred. You are no longer the board owner.',
    title: 'Transfer Ownership',
    selectLabel: 'Select New Owner',
    selectPlaceholder: 'Select User',
    cancel: 'Cancel',
    confirm: 'Transfer',
    reclaim: {
      tooltip: 'Reclaim board ownership',
      title: 'Reclaim Ownership',
      text: 'This action will make you board owner again.',
      confirm: 'Reclaim',
    },
  },
}
