export default {
  langName: 'Español',
  common: {
    anonymous: 'Anónimo',
    minutes: 'Minutos',
    seconds: 'Segundos',
    start: 'Iniciar',
    stop: 'Detener',
    copy: 'Copiar',
    board: 'Tablero',
    toolTips: {
      darkTheme: 'Activar tema oscuro',
      lightTheme: 'Activar tema claro',
    },
    contentOverloadError: 'Contenido excede el límite permitido.',
    contentStrippingError:
      'Contenido excede el límite permitido. El texto adicional ha sido eliminado.',
    invalidColumnSelection: 'Por favor selecciona columna(s)',
    typing: '{name} está escribiendo',
    share: {
      linkCopied: '¡Enlace copiado!',
      linkCopyError: 'Error al copiar. Copia manualmente.',
    },
    customColumnSetup: {
      shareLabel: 'Compartir configuración de columnas',
      shareHelp:
        'Copia el enlace de abajo para compartir tu configuración de columnas personalizada con otros, o guárdalo en favoritos para usarlo más tarde.',
      applied: 'Configuración de columnas personalizada aplicada',
    },
  },
  join: {
    label: 'Unirse como invitado',
    namePlaceholder: '¡Escribe tu nombre aquí!',
    nameRequired: 'Por favor ingresa tu nombre',
    button: 'Unirse',
  },
  createBoard: {
    label: 'Crear tablero',
    namePlaceholder: '¡Escribe el nombre del tablero aquí!',
    nameRequired: 'Por favor ingresa el nombre del tablero',
    teamNamePlaceholder: '¡Escribe el nombre del equipo aquí!',
    button: 'Crear',
    buttonProgress: 'Creando..',
    captchaInfo: 'Por favor, complete el CAPTCHA para continuar',
    boardCreationError: 'Error al crear el tablero',
    columns: 'Columnas',
  },
  dashboard: {
    timer: {
      oneMinuteLeft: 'Queda un minuto',
      timeCompleted: '¡Se ha acabado el tiempo!',
      title: 'Iniciar/Detener temporizador',
      helpTip:
        'Ajusta minutos y segundos con los controles + - o flechas del teclado. Máximo: 1 hora.',
      invalid: 'Valores inválidos. Rango permitido: 1 segundo a 60 minutos.',
      tooltip: 'Temporizador regresivo',
    },
    share: {
      title: 'Copia y comparte esta URL con participantes',
      toolTip: 'Compartir tablero',
    },
    mask: {
      maskTooltip: 'Ocultar mensajes',
      unmaskTooltip: 'Mostrar mensajes',
    },
    lock: {
      lockTooltip: 'Bloquear tablero',
      unlockTooltip: 'Desbloquear tablero',
      message: 'Tablero bloqueado por el propietario.',
      discardChanges: '¡Tablero bloqueado! Los mensajes no guardados se han descartado',
    },
    spotlight: {
      noCardsToFocus: 'No hay tarjetas para enfocar',
      tooltip: 'Enfocar tarjetas',
    },
    print: {
      tooltip: 'Imprimir',
    },
    language: {
      tooltip: 'Cambiar idioma',
    },
    delete: {
      title: 'Confirmar eliminación',
      text: 'Los datos no se pueden recuperar después de ser eliminados. ¿Seguro que desea continuar?',
      tooltip: 'Eliminar este tablero',
      continueDelete: 'Sí',
      cancelDelete: 'No',
    },
    columns: {
      col01: 'Lo que salió bien',
      col02: 'Desafíos',
      col03: 'Acciones',
      col04: 'Agradecimientos',
      col05: 'Mejoras',
      cannotDisable: 'No se puede desactivar la(s) columna(s) que tienen tarjetas',
      update: 'Actualizar',
      discardNewMessages: 'Tu borrador se ha descartado porque la columna fue deshabilitada.',
    },
    printFooter: 'Creado con',
    offline: 'Sin conexión.',
    notExists: 'El tablero fue eliminado automáticamente o lo eliminó manualmente su creador.',
    autoDeleteScheduleBase: 'Este tablero se limpiará automáticamente el {date}',
    autoDeleteScheduleAddon: ', así que no necesitas preocuparte por eliminarlo manualmente.',
  },
}
