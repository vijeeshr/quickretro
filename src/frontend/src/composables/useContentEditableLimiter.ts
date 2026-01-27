import { computed } from 'vue'
import { assertMessageContentValidation, calculateContentBudget, canAssertMessageContentValidation, debounce, throttleRAF, MessageContentValidationResult } from '../utils'
import { useI18n } from 'vue-i18n'

interface UseLimiterOptions {
  user: () => string
  nickname: () => string
  board: () => string
  category: () => string
  isComment?: boolean
  onInvalid: (msg: string) => void
}

export function useContentEditableLimiter(opts: UseLimiterOptions) {
  const { user, nickname, board, category, isComment = false, onInvalid } = opts
  const { t } = useI18n()
  
  // Precomputed byte budget: Only re-calculates if props change, not when the user types
  const contentByteBudget = computed(() =>
    calculateContentBudget(user(), nickname(), board(), category(), isComment)
  )

  // Wait 500ms after the user stops hitting the limit to fire the toast.
  const debouncedEmitInvalid = debounce((msg: string) => {
    onInvalid(msg)
  }, 500)

  // Handle the physical trimming at 60fps approx
  const validate = throttleRAF((event: Event) => {
    if (!canAssertMessageContentValidation()) return

    // Pass the pre-calculated budget
    const validationResult: MessageContentValidationResult = assertMessageContentValidation(event, contentByteBudget.value)
    if (validationResult.isValid) return

    const errorMessage: string = validationResult.isTrimmed ? t('common.contentStrippingError') : t('common.contentOverloadError')

    debouncedEmitInvalid(errorMessage)
  })

  return {
    onInput: validate
  }
}