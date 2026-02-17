import { computed } from 'vue'
import {
  assertMessageContentValidation,
  calculateContentBudget,
  debounce,
  throttleRAF,
  MessageContentValidationResult,
} from '../utils'
import { useI18n } from 'vue-i18n'
import { CONTENT_EDITABLE_INVALID_DEBOUNCE_MS } from '../utils/appConfig'

interface UseLimiterOptions {
  nickname: () => string
  category: () => string
  anon: () => boolean
  isComment?: boolean
  onInvalid: (msg: string) => void
}

export function useContentEditableLimiter(opts: UseLimiterOptions) {
  const { nickname, category, anon, isComment = false, onInvalid } = opts
  const { t } = useI18n()

  // Precomputed byte budget: Only re-calculates if props change, not when the user types
  const contentByteBudget = computed(() =>
    calculateContentBudget(nickname(), category(), anon(), isComment)
  )

  // Wait *ms after the user stops hitting the limit to fire the toast
  const debouncedEmitInvalid = debounce((msg: string) => {
    onInvalid(msg)
  }, CONTENT_EDITABLE_INVALID_DEBOUNCE_MS)

  // Handle the physical trimming at 60fps approx
  const validate = throttleRAF((event: Event) => {
    // Pass the pre-calculated budget
    const validationResult: MessageContentValidationResult = assertMessageContentValidation(
      event,
      contentByteBudget.value
    )
    if (validationResult.isValid) return

    const errorMessage: string = validationResult.isTrimmed
      ? t('common.contentStrippingError')
      : t('common.contentOverloadError')

    debouncedEmitInvalid(errorMessage)
  })

  return {
    onInput: validate,
  }
}
