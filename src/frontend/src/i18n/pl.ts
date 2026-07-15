export default {
  common: {
    anonymous: 'Anonimowy',
    minutes: 'Minuty',
    seconds: 'Sekundy',
    start: 'Start',
    stop: 'Stop',
    copy: 'Kopiuj',
    board: 'Tablica',
    toolTips: {
      darkTheme: 'Włącz tryb ciemny',
      lightTheme: 'Włącz tryb jasny',
    },
    contentOverloadError: 'Treść przekracza dozwolony limit.',
    contentStrippingError:
      'Treść przekracza dozwolony limit. Nadmiarowy tekst został obcięty z końca.',
    invalidColumnSelection: 'Proszę wybrać kolumnę/kolumny',
    typing: '{name} pisze',
    share: {
      linkCopied: 'Link skopiowany!',
      linkCopyError: 'Nie udało się skopiować. Proszę skopiować ręcznie.',
    },
    customColumnSetup: {
      shareLabel: 'Udostępnij konfigurację kolumn',
      shareHelp:
        'Skopiuj poniższy link, aby udostępnić swoją niestandardową konfigurację kolumn innym lub zapisać go na później.',
      applied: 'Zastosowano niestandardową konfigurację kolumn',
    },
  },
  join: {
    label: 'Dołącz jako gość',
    namePlaceholder: 'Wpisz tutaj swoje imię!',
    nameRequired: 'Proszę podać swoje imię',
    button: 'Dołącz',
  },
  createBoard: {
    label: 'Utwórz tablicę',
    namePlaceholder: 'Wpisz nazwę tablicy!',
    nameRequired: 'Proszę podać nazwę tablicy',
    teamNamePlaceholder: 'Wpisz nazwę zespołu!',
    button: 'Utwórz',
    buttonProgress: 'Tworzenie...',
    captchaInfo: 'Proszę ukończyć CAPTCHA, aby kontynuować',
    boardCreationError: 'Błąd podczas tworzenia tablicy',
    columns: 'Kolumny',
  },
  dashboard: {
    timer: {
      oneMinuteLeft: 'Pozostała jedna minuta do końca odliczania',
      timeCompleted: 'Hej! Czas się skończył',
      title: 'Start/Stop timera',
      helpTip:
        'Dostosuj minuty i sekundy używając przycisków + i -, lub strzałek w górę i w dół na klawiaturze. Maksymalnie 1 godzina.',
      invalid:
        'Proszę wprowadzić prawidłowe wartości minut/sekund. Dozwolony zakres to od 1 sekundy do 60 minut.',
      tooltip: 'Timer odliczania',
      shortText: 'Timer',
    },
    share: {
      toolTip: 'Udostępnij tablicę innym',
      shortText: 'Udostępn',
    },
    theme: {
      shortText: 'Motyw',
    },
    mask: {
      maskTooltip: 'Ukryj wiadomości',
      maskShortText: 'Ukryj',
      unmaskTooltip: 'Pokaż wiadomości',
      unmaskShortText: 'Pokaż',
    },
    lock: {
      lockTooltip: 'Zablokuj tablicę',
      lockShortText: 'Blok',
      unlockTooltip: 'Odblokuj tablicę',
      unlockShortText: 'Odblok',
      message: 'Nie można dodać ani zaktualizować. Tablica jest zablokowana przez właściciela.',
      discardChanges: 'Tablica zablokowana! Niezapisane wiadomości zostały odrzucone',
      unlockButton: 'Odblokuj',
    },
    spotlight: {
      noCardsToFocus: 'Brak kart do wyświetlenia',
      tooltip: 'Skup się na kartach',
      shortText: 'Fokus',
    },
    filter: {
      tooltip: 'Pokaż karty z polubieniami/komentarzami na górze',
      shortText: 'Sortuj',
      likesMiniTooltip: 'Najbardziej polubione karty na górze',
      commentsMiniTooltip: 'Najbardziej komentowane karty na górze',
    },
    print: {
      tooltip: 'Drukuj',
      shortText: 'Druk',
      withNamesTooltip: 'Drukuj (dołącz nazwy)',
      withCommentsTooltip: 'Drukuj (dołącz komentarze)',
      withAllTooltip: 'Drukuj (dołącz nazwy i komentarze)',
    },
    language: {
      tooltip: 'Zmień język',
      shortText: 'Język',
    },
    delete: {
      title: 'Potwierdź usunięcie',
      text: 'Po usunięciu danych nie można ich odzyskać. Czy na pewno chcesz kontynuować?',
      tooltip: 'Usuń tę tablicę',
      continueDelete: 'Tak',
      cancelDelete: 'Nie',
      shortText: 'Usuń',
    },
    columns: {
      col01: 'Co poszło dobrze',
      col02: 'Wyzwania',
      col03: 'Działania do podjęcia',
      col04: 'Docenienia',
      col05: 'Ulepszenia',
      cannotDisable: 'Nie można wyłączyć kolumny z przypisanymi kartami',
      update: 'Aktualizuj',
      discardNewMessages: 'Twój szkic został odrzucony, ponieważ kolumna została wyłączona.',
    },
    printFooter: 'Stworzone za pomocą',
    offline: 'Wygląda na to, że jesteś offline.',
    autoDeleteSchedule: 'Ta tablica zostanie automatycznie usunięta dnia {date}.',
    welcome: {
      title: 'Witaj!',
      maskInfo:
        'Wiadomości są domyślnie maskowane. Tylko autor może widzieć swoje wiadomości, dopóki nie cofniesz maskowania.',
      maskOnLabel: 'Maskowanie kart jest WŁĄCZONE',
      maskOffLabel: 'Maskowanie kart jest WYŁĄCZONE',
      ok: 'Rozumiem!',
    },
    notFound: {
      title: 'Nie znaleziono tablicy',
      text: 'Tablica, której szukasz, została usunięta automatycznie lub ręcznie przez jej właściciela.',
      createNewBoard: 'Utwórz nową tablicę',
      supportText: 'Wesprzyj nas gwiazdką na GitHubie! ⭐',
    },
    help: {
      shortText: 'Pomoc',
    },
    offlineLikes: {
      text: 'Głosy / polubienia offline',
      showPanelTooltip: 'Pokaż panel polubień offline',
      hidePanelTooltip: 'Ukryj panel polubień offline',
    },
    settings: {
      tooltip: 'Więcej opcji...',
      shortText: 'Opcje',
    },
    download: {
      jsonTooltip: 'Pobierz jako JSON',
    },
  },
  transferOwnership: {
    tooltip: 'Przekaż własność tablicy',
    promotedNotification:
      'Jesteś teraz właścicielem tablicy. Możesz zarządzać ustawieniami i uczestnikami.',
    demotedNotification: 'Własność została przekazana. Nie jesteś już właścicielem tablicy.',
    title: 'Przekaż własność',
    selectLabel: 'Wybierz nowego właściciela',
    selectPlaceholder: 'Wybierz użytkownika',
    cancel: 'Anuluj',
    confirm: 'Przekaż',
    shortText: 'Przekaż',
    reclaim: {
      tooltip: 'Odzyskaj własność tablicy',
      title: 'Odzyskaj własność',
      text: 'Ta czynność sprawi, że ponownie staniesz się właścicielem tablicy.',
      confirm: 'Odzyskaj',
      shortText: 'Odzysk',
    },
  },
}
