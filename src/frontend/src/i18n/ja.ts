export default {
    langName: '日本語 (ja)',   
    common: {
        anonymous: '匿名',
        minutes: '分',
        seconds: '秒',
        start: '開始',
        stop: '停止',
        copy: 'コピー',
        board: 'ボード',
        toolTips: {
            darkTheme: 'ダークテーマ',
            lightTheme: 'ライトテーマ'
        },
        contentOverloadError: 'コンテンツ制限超過',
        contentStrippingError: '末尾のテキストが削除されました',
        invalidColumnSelection: '列を選択してください'
    },
    join: {
        label: 'ゲスト参加',
        namePlaceholder: '名前を入力してください！',
        nameRequired: '名前を入力してください',
        button: '参加'
    },
    createBoard: {
        label: 'ボード作成',
        namePlaceholder: 'ボード名を入力！',
        nameRequired: 'ボード名を入力してください',
        teamNamePlaceholder: 'チーム名を入力！',
        button: '作成',
		buttonProgress: '作成中..',
        captchaInfo: 'CAPTCHAを完了してください',
        boardCreationError: 'ボードの作成中にエラーが発生しました'
    },
    dashboard: {
        timer: {
            oneMinuteLeft: '残り1分',
            timeCompleted: '時間切れです！',
            title: 'タイマー開始/停止',
            helpTip: '+/-または矢印キーで調整 最大1時間',
            invalid: '無効な時間です（1秒～60分）',
            tooltip: 'カウントダウンタイマー'
        },
        share: {
            title: 'URLを共有',
            linkCopied: 'コピーしました！',
            linkCopyError: 'コピー失敗 手動でコピーしてください',
            toolTip: 'ボードを共有'
        },
        mask: {
            maskTooltip: 'メッセージを非表示',
            unmaskTooltip: 'メッセージを表示'
        },
        lock: {
            lockTooltip: 'ボードをロック',
            unlockTooltip: 'ロック解除',
            message: 'ボードがロックされています',
            discardChanges: 'ボードがロックされました！保存されていないメッセージは破棄されました'
        },
        spotlight: {
            noCardsToFocus: 'カードがありません',
            tooltip: 'カードをフォーカス'
        },
        print: {
            tooltip: '印刷'
        },
        language: {
            tooltip : '言語を変更'
        },
        delete: {
            title: '削除の確認',
            text: '削除後はデータを復元できません。続行してもよろしいですか？',
            tooltip: 'このボードを削除',
            continueDelete: 'はい',
            cancelDelete: 'いいえ'
        },
        columns: {
            col01: '良かった点',
            col02: '課題',
            col03: 'アクション項目',
            col04: '感謝',
            col05: '改善点',
            cannotDisable: "カードがある列は無効にできません",
            update: "更新"
        },
        printFooter: '作成者',
        offline: 'オフライン',
        notExists: 'ボードは自動的に削除されたか、作成者によって手動で削除されました。',
        autoDeleteScheduleBase: 'このボードは {date} に自動的にクリーンアップされます',
        autoDeleteScheduleAddon: 'ので、手動で削除する必要はありません。'
    }
}