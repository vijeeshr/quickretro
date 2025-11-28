export default {
    langName: 'English',
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
            lightTheme: 'Turn on light theme'
        },
        contentOverloadError: 'Content more than allowed limit.',
        contentStrippingError: 'Content more than allowed limit. Extra text is stripped from the end.',
        invalidColumnSelection: 'Please select column(s)'
    },
    join: {
        label: 'Join as guest',
        namePlaceholder: 'Type your name here!',
        nameRequired: 'Please enter your name',
        button: 'Join'
    },
    createBoard: {
        label: 'Create Board',
        namePlaceholder: 'Type board name here!',
        nameRequired: 'Please enter board name',
        teamNamePlaceholder: 'Type team name here!',
        button: 'Create',
        buttonProgress: 'Creating..',
        captchaInfo: 'Please complete the CAPTCHA to continue',
        boardCreationError: 'Error when creating board'
    },
    dashboard: {
        timer: {
            oneMinuteLeft: 'One minute left for countdown',
            timeCompleted: "Hey! You've run out of time",
            title: 'Start/Stop Timer',
            helpTip: 'Adjust minutes and seconds using the + and - controls, or the Up and Down arrows on keyboard. Max allowed is 1 hour.',
            invalid: 'Please enter valid minutes/seconds values. Allowed range is 1 second to 60 minutes.',
            tooltip: 'Countdown Timer'
        },
        share: {
            title: 'Copy and share below url to participants',
            linkCopied: 'Link copied!',
            linkCopyError: 'Failed to copy. Please copy directly.',
            toolTip: 'Share board with others'
        },
        mask: {
            maskTooltip: 'Mask messages',
            unmaskTooltip: 'Unmask messages'
        },
        lock: {
            lockTooltip: 'Lock board',
            unlockTooltip: 'Unlock board',
            message: 'Cannot add or update. Board is locked by owner.',
            discardChanges: 'Board locked! Unsaved messages discarded'
        },
        spotlight: {
            noCardsToFocus: 'There are no cards to focus',
            tooltip: 'Focus cards'
        },
        print: {
            tooltip: 'Print'
        },
        language: {
            tooltip: 'Change language'
        },
        delete: {
            title: 'Confirm deletion',
            text: 'Data cannot be recovered after its deleted. Are you sure to proceed?',
            tooltip: 'Delete this Board',
            continueDelete: 'Yes',
            cancelDelete: 'No'
        },
        columns: {
            col01: 'What went well',
            col02: 'Challenges',
            col03: 'Action Items',
            col04: 'Appreciations',
            col05: 'Improvements',
            cannotDisable: 'Cannot disable column(s) with cards',
            update: 'Update'
        },
        printFooter: 'Created with',
        offline: 'You seem to be offline.',
        notExists: 'Board was either auto-deleted, or manually deleted by its creator.',
        autoDeleteScheduleBase: 'This board will be cleaned up automatically on {date}',
        autoDeleteScheduleAddon: ', so you do not need to worry about deleting it manually.'
    }
}