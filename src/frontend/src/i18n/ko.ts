export default {
    langName: '한국어 (ko)',  
    common: {
        anonymous: '익명',
        minutes: '분',
        seconds: '초',
        start: '시작',
        stop: '중지',
        copy: '복사',
        board: '보드',
        toolTips: {
            darkTheme: '다크 모드 켜기',
            lightTheme: '라이트 모드 켜기'
        },
        contentOverloadError: '허용된 내용 초과',
        contentStrippingError: '초과된 텍스트가 삭제되었습니다',
        invalidColumnSelection: '열을 선택해 주세요'
    },
    join: {
        label: '게스트로 참여',
        namePlaceholder: '이름을 입력하세요!',
        nameRequired: '이름을 입력해 주세요',
        button: '참여'
    },
    createBoard: {
        label: '보드 생성',
        namePlaceholder: '보드 이름 입력!',
        nameRequired: '보드 이름을 입력해 주세요',
        teamNamePlaceholder: '팀 이름 입력!',
        button: '생성',
		buttonProgress: '생성 중..',
        captchaInfo: '계속하려면 CAPTCHA를 완료하세요',
        boardCreationError: '보드 생성 중 오류 발생'
    },
    dashboard: {
        timer: {
            oneMinuteLeft: '1분 남음',
            timeCompleted: '시간 종료!',
            title: '타이머 시작/중지',
            helpTip: '+ - 또는 방향키로 시간 조절. 최대 1시간.',
            invalid: '유효하지 않은 시간 (1초 ~ 60분)',
            tooltip: '카운트다운 타이머'
        },
        share: {
            title: '링크 공유',
            linkCopied: '링크 복사됨!',
            linkCopyError: '복사 실패. 직접 복사해 주세요.',
            toolTip: '보드 공유'
        },
        mask: {
            maskTooltip: '메시지 숨기기',
            unmaskTooltip: '메시지 표시'
        },
        lock: {
            lockTooltip: '보드 잠금',
            unlockTooltip: '잠금 해제',
            message: '보드가 잠겨 있습니다',
            discardChanges: '보드 잠김! 저장되지 않은 메시지가 삭제되었습니다'
        },
        spotlight: {
            noCardsToFocus: '포커스할 카드 없음',
            tooltip: '카드 강조'
        },
        print: {
            tooltip: '인쇄'
        },
        language: {
            tooltip : '언어 변경'
        },
        delete: {
            title: '삭제 확인',
            text: '삭제 후에는 데이터를 복구할 수 없습니다. 계속 진행하시겠습니까?',
            tooltip: '이 보드 삭제',
            continueDelete: '예',
            cancelDelete: '아니오'
        },
        columns: {
            col01: '잘된 점',
            col02: '어려운 점',
            col03: '액션 항목',
            col04: '감사한 점',
            col05: '개선점',
            cannotDisable: "카드가 있는 열은 비활성화할 수 없습니다",
            update: "업데이트",
            discardNewMessages: '열이 비활성화되어 임시 작성 내용이 삭제되었습니다.'
        },
        printFooter: '생성 도구',
        offline: '오프라인 상태',
        notExists: '보드는 자동으로 삭제되었거나 생성자가 수동으로 삭제했습니다.',
        autoDeleteScheduleBase: '{date}에 이 보드는 자동으로 정리됩니다',
        autoDeleteScheduleAddon: ', 따라서 직접 삭제할 필요가 없습니다.'
    }
}