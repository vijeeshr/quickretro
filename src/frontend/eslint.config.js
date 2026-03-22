import js from '@eslint/js'
import vue from 'eslint-plugin-vue'
import tseslint from '@typescript-eslint/eslint-plugin'
import tsParser from '@typescript-eslint/parser'
import vueParser from 'vue-eslint-parser'
import prettier from 'eslint-plugin-prettier'
import configPrettier from 'eslint-config-prettier'
import globals from 'globals'

export default [
  // Base ESLint recommended rules
  js.configs.recommended,

  // TypeScript recommended rules
  // ...tseslint.configs.recommended,

  // Vue 3 recommended rules (Flat Config format)
  ...vue.configs['flat/recommended'],

  {
    // Global ignores (replaces .eslintignore)
    ignores: ['dist/**', 'node_modules/**', 'public/**'],
  },

  {
    files: ['**/*.ts', '**/*.vue'],
    languageOptions: {
      parser: vueParser,
      globals: {
        ...globals.browser,
        ...globals.node,
        ...globals.es2021,
      },
      parserOptions: {
        parser: tsParser,
        ecmaVersion: 'latest',
        sourceType: 'module',
      },
    },
    plugins: {
      '@typescript-eslint': tseslint,
      prettier,
    },
    // Custom rules
    rules: {
      ...tseslint.configs.recommended.rules,

      // Prettier integration
      'prettier/prettier': [
        'error',
        {
          endOfLine: 'lf', // Enforce LF line endings for cross-platform consistency
        },
      ],

      // Optional tweaks (good defaults)
      // 'vue/no-unused-vars': 'error',
      'vue/multi-word-component-names': 'off',
      '@typescript-eslint/no-unused-vars': ['warn'],
    },
  },

  // Disable formatting conflicts
  // Prettier integration (must be last to override others)
  configPrettier,
]
