export default {
    langName: '简体中文 (zh-CN)',  
    common: {
        anonymous: '匿名',
        minutes: '分钟',
        seconds: '秒',
        start: '开始',
        stop: '停止',
        copy: '复制',
        board: '看板',
        toolTips: {
            darkTheme: '启用深色主题',
            lightTheme: '启用浅色主题'
        },
        contentOverloadError: '内容超过允许限制',
        contentStrippingError: '内容超出限制，多余文字已被删除',
        invalidColumnSelection: '请选择列'
    },
    join: {
        label: '以访客加入',
        namePlaceholder: '在此输入姓名！',
        nameRequired: '请输入姓名',
        button: '加入'
    },
    createBoard: {
        label: '创建看板',
        namePlaceholder: '输入看板名称！',
        nameRequired: '请输入看板名称',
        teamNamePlaceholder: '输入团队名称！',
        invalidColumnSelection: '请选择列',
        button: '创建',
		buttonProgress: '创建中..',
        captchaInfo: '请完成验证码以继续',
        boardCreationError: '创建看板时出错'
    },
    dashboard: {
        timer: {
            oneMinuteLeft: '剩余一分钟',
            timeCompleted: '时间到！',
            title: '开始/停止计时器',
            helpTip: '使用+ -或方向键调整时间，最长1小时',
            invalid: '无效时间，允许范围：1秒至60分钟',
            tooltip: '倒计时器'
        },
        share: {
            title: '复制并分享链接',
            linkCopied: '链接已复制！',
            linkCopyError: '复制失败，请手动复制',
            toolTip: '分享看板'
        },
        mask: {
            maskTooltip: '隐藏消息',
            unmaskTooltip: '显示消息'
        },
        lock: {
            lockTooltip: '锁定看板',
            unlockTooltip: '解锁看板',
            message: '看板已被锁定',
            discardChanges: '看板已锁定！未保存的消息已丢弃'
        },
        spotlight: {
            noCardsToFocus: '没有可聚焦的卡片',
            tooltip: '聚焦卡片'
        },
        print: {
            tooltip: '打印'
        },
        language: {
            tooltip : '更改语言'
        },
        delete: {
            title: '确认删除',
            text: '数据删除后无法恢复。确定要继续吗？',
            tooltip: '删除此看板',
            continueDelete: '是',
            cancelDelete: '否'
        },
        columns: {
            col01: '做得好的',
            col02: '挑战',
            col03: '行动计划',
            col04: '感谢',
            col05: '改进建议',
            cannotDisable: "无法禁用包含卡片的列",
            update: "更新",
            discardNewMessages: '您的草稿已被丢弃，因为该列已被禁用。'
        },
        printFooter: '创建于',
        offline: '离线状态',
        notExists: '看板已被自动删除，或由创建者手动删除。',
        autoDeleteScheduleBase: '该看板将于 {date} 自动清理',
        autoDeleteScheduleAddon: '，因此您无需担心手动删除它。'
    }
}