module.exports = {
  env: {
    browser: true,
    es2021: true,
    node: true
  },
  extends: [
    'eslint:recommended',
    "plugin:@typescript-eslint/recommended",
    "plugin:@typescript-eslint/recommended-requiring-type-checking",
    "plugin:react/recommended",
    "plugin:eslint-comments/recommended",
    "plugin:import/errors",
    "plugin:import/warnings",
    "plugin:import/typescript",
    "plugin:prettier/recommended"
  ],
  parser: '@typescript-eslint/parser',
  parserOptions: {
    ecmaFeatures: {
      jsx: true
    },
    ecmaVersion: 12,
    sourceType: 'module',
    project: './tsconfig.json'
  },
  plugins: [
    'react',
    '@typescript-eslint'
  ],
  rules: {
    'no-console': process.env.NODE_ENV === 'production' ? 'error' : 'off',
    'no-debugger': process.env.NODE_ENV === 'production' ? 'error' : 'off',
    eqeqeq: 'error',
    quotes: ['error', 'single', { allowTemplateLiterals: true, avoidEscape: true }],
    'prettier/prettier': ['error', { singleQuote: true, printWidth: 120 }],
    "@typescript-eslint/explicit-module-boundary-types": "off",
    "eslint-comments/disable-enable-pair": [
      "error",
      { "allowWholeFile": true }
    ],
    "eslint-comments/no-unused-disable": "error",
    "import/order": [
      "error",
      { "newlines-between": "always", "alphabetize": { "order": "asc" } }
    ],
    "sort-imports": [
      "error",
      { "ignoreDeclarationSort": true, "ignoreCase": true }
    ]
  },
}