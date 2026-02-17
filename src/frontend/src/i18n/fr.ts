export default {
  langName: 'Français',
  common: {
    anonymous: 'Anonyme',
    minutes: 'Minutes',
    seconds: 'Secondes',
    start: 'Démarrer',
    stop: 'Arrêter',
    copy: 'Copier',
    board: 'Tableau',
    toolTips: {
      darkTheme: 'Activer le mode sombre',
      lightTheme: 'Activer le mode clair',
    },
    contentOverloadError: 'Contenu dépasse la limite.',
    contentStrippingError: 'Contenu trop long. Texte excédentaire supprimé.',
    invalidColumnSelection: 'Veuillez sélectionner des colonnes',
    typing: '{name} est en train d’écrire',
    share: {
      linkCopied: 'Lien copié !',
      linkCopyError: 'Échec de copie. Copiez manuellement.',
    },
    customColumnSetup: {
      shareLabel: 'Partager la configuration des colonnes',
      shareHelp:
        'Copiez le lien ci-dessous pour partager votre configuration de colonnes personnalisée avec d’autres, ou ajoutez-le à vos favoris pour plus tard.',
      applied: 'Configuration de colonnes personnalisée appliquée',
    },
  },
  join: {
    label: 'Rejoindre en invité',
    namePlaceholder: 'Saisissez votre nom ici !',
    nameRequired: 'Veuillez saisir votre nom',
    button: 'Rejoindre',
  },
  createBoard: {
    label: 'Créer un tableau',
    namePlaceholder: 'Nom du tableau ici !',
    nameRequired: 'Veuillez saisir le nom du tableau',
    teamNamePlaceholder: "Nom de l'équipe ici !",
    button: 'Créer',
    buttonProgress: 'Création en cours..',
    captchaInfo: 'Veuillez compléter le CAPTCHA pour continuer',
    boardCreationError: 'Erreur lors de la création du tableau',
    columns: 'Colonnes',
  },
  dashboard: {
    timer: {
      oneMinuteLeft: 'Une minute restante',
      timeCompleted: 'Temps écoulé !',
      title: 'Démarrer/Arrêter le chrono',
      helpTip: 'Ajustez les minutes/secondes avec +/- ou flèches. Maximum 1 heure.',
      invalid: 'Valeurs invalides. Plage autorisée : 1 seconde à 60 minutes.',
      tooltip: 'Minuteur',
    },
    share: {
      title: "Copier et partager l'URL",
      toolTip: 'Partager le tableau',
    },
    mask: {
      maskTooltip: 'Masquer messages',
      unmaskTooltip: 'Afficher messages',
    },
    lock: {
      lockTooltip: 'Verrouiller tableau',
      unlockTooltip: 'Déverrouiller tableau',
      message: 'Tableau verrouillé.',
      discardChanges: 'Tableau verrouillé ! Messages non enregistrés supprimés',
    },
    spotlight: {
      noCardsToFocus: 'Aucune carte à focaliser',
      tooltip: 'Mettre en avant',
    },
    print: {
      tooltip: 'Imprimer',
    },
    language: {
      tooltip: 'Changer de langue',
    },
    delete: {
      title: 'Confirmer la suppression',
      text: 'Les données ne peuvent pas être récupérées après leur suppression. Êtes-vous sûr de vouloir continuer?',
      tooltip: 'Supprimer ce tableau',
      continueDelete: 'Oui',
      cancelDelete: 'Non',
    },
    columns: {
      col01: 'Ce qui a bien fonctionné',
      col02: 'Défis',
      col03: 'Actions',
      col04: 'Reconnaissance',
      col05: 'Améliorations',
      cannotDisable: 'Impossible de désactiver la colonne car elle contient des cartes',
      update: 'Mettre à jour',
      discardNewMessages: 'Votre brouillon a été supprimé car la colonne a été désactivée.',
    },
    printFooter: 'Créé avec',
    offline: 'Hors ligne.',
    notExists:
      'Le tableau a été soit supprimé automatiquement, soit manuellement par son créateur.',
    autoDeleteScheduleBase: 'Ce tableau sera automatiquement nettoyé le {date}',
    autoDeleteScheduleAddon: ', vous n’avez donc pas à vous soucier de le supprimer manuellement.',
  },
}
