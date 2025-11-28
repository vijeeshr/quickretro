export default {
    langName: 'Русский (ru)',  
    common: {
        anonymous: 'Аноним',
        minutes: 'Минуты',
        seconds: 'Секунды',
        start: 'Старт',
        stop: 'Стоп',
        copy: 'Копировать',
        board: 'Доска',
        toolTips: {
            darkTheme: 'Тёмная тема',
            lightTheme: 'Светлая тема'
        },
        contentOverloadError: 'Превышен лимит содержимого',
        contentStrippingError: 'Лишний текст удалён',
        invalidColumnSelection: 'Выберите столбцы'
    },
    join: {
        label: 'Войти как гость',
        namePlaceholder: 'Введите имя здесь!',
        nameRequired: 'Введите имя',
        button: 'Присоединиться'
    },
    createBoard: {
        label: 'Создать доску',
        namePlaceholder: 'Название доски здесь!',
        nameRequired: 'Введите название доски',
        teamNamePlaceholder: 'Название команды здесь!',
        button: 'Создать',
		buttonProgress: 'Создание..',
        captchaInfo: 'Пройдите CAPTCHA для продолжения',
        boardCreationError: 'Ошибка при создании доски'
    },
    dashboard: {
        timer: {
            oneMinuteLeft: 'Осталась минута',
            timeCompleted: 'Время вышло!',
            title: 'Старт/Стоп таймер',
            helpTip: 'Используйте +/- или стрелки. Макс. 1 час.',
            invalid: 'Недопустимое время (1 сек - 60 мин)',
            tooltip: 'Таймер обратного отсчёта'
        },
        share: {
            title: 'Скопируйте и поделитесь ссылкой',
            linkCopied: 'Ссылка скопирована!',
            linkCopyError: 'Ошибка копирования',
            toolTip: 'Поделиться доской'
        },
        mask: {
            maskTooltip: 'Скрыть сообщения',
            unmaskTooltip: 'Показать сообщения'
        },
        lock: {
            lockTooltip: 'Заблокировать доску',
            unlockTooltip: 'Разблокировать доску',
            message: 'Доска заблокирована',
            discardChanges: 'Доска заблокирована! Несохранённые сообщения удалены'
        },
        spotlight: {
            noCardsToFocus: 'Нет карточек',
            tooltip: 'Выделить карточки'
        },
        print: {
            tooltip: 'Печать'
        },
        language: {
            tooltip : 'Сменить язык'
        },
        delete: {
            title: 'Подтвердите удаление',
            text: 'После удаления данные невозможно восстановить. Вы уверены, что хотите продолжить?',
            tooltip: 'Удалить эту доску',
            continueDelete: 'Да',
            cancelDelete: 'Нет'
        },
        columns: {
            col01: 'Что прошло хорошо',
            col02: 'Сложности',
            col03: 'Действия',
            col04: 'Благодарности',
            col05: 'Улучшения',
            cannotDisable: "Нельзя отключить колонку, в которой есть карточки",
            update: "Обновить"
        },
        printFooter: 'Создано с',
        offline: 'Офлайн',
        notExists: 'Доска была удалена автоматически или вручную её создателем.',
        autoDeleteScheduleBase: 'Эта доска будет автоматически очищена {date}',
        autoDeleteScheduleAddon: ', поэтому вам не нужно беспокоиться о её ручном удалении.'
    }
}