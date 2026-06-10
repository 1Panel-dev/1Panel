const tsParser = require('@typescript-eslint/parser');
const tsPlugin = require('@typescript-eslint/eslint-plugin');
const vueParser = require('vue-eslint-parser');
const vuePlugin = require('eslint-plugin-vue');
const prettierPlugin = require('eslint-plugin-prettier');
const eslintConfigPrettier = require('eslint-config-prettier');

const globals = {
    window: 'readonly',
    document: 'readonly',
    navigator: 'readonly',
    console: 'readonly',
    localStorage: 'readonly',
    sessionStorage: 'readonly',
    setTimeout: 'readonly',
    clearTimeout: 'readonly',
    setInterval: 'readonly',
    clearInterval: 'readonly',
    module: 'readonly',
    require: 'readonly',
    process: 'readonly',
    __dirname: 'readonly',
};

const commonRules = {
    'no-var': 'error',
    'no-multiple-empty-lines': ['error', { max: 1 }],
    'no-use-before-define': 'off',
    'prefer-const': 'off',
    'no-irregular-whitespace': 'off',
    'prettier/prettier': 'error',
};

const tsRules = {
    ...tsPlugin.configs.recommended.rules,
    '@typescript-eslint/no-unused-vars': ['error', { caughtErrors: 'none' }],
    '@typescript-eslint/no-inferrable-types': 'off',
    '@typescript-eslint/no-namespace': 'off',
    '@typescript-eslint/no-explicit-any': 'off',
    '@typescript-eslint/ban-ts-ignore': 'off',
    '@typescript-eslint/ban-types': 'off',
    '@typescript-eslint/explicit-function-return-type': 'off',
    '@typescript-eslint/no-var-requires': 'off',
    '@typescript-eslint/no-empty-function': 'off',
    '@typescript-eslint/no-empty-object-type': 'off',
    '@typescript-eslint/no-use-before-define': 'off',
    '@typescript-eslint/no-unused-expressions': 'off',
    '@typescript-eslint/no-unsafe-function-type': 'off',
    '@typescript-eslint/ban-ts-comment': 'off',
    '@typescript-eslint/no-non-null-assertion': 'off',
    '@typescript-eslint/explicit-module-boundary-types': 'off',
};

const vueRules = {
    'vue/no-v-html': 'off',
    'vue/v-slot-style': 'error',
    'vue/no-mutating-props': 'off',
    'vue/custom-event-name-casing': 'off',
    'vue/attributes-order': 'off',
    'vue/one-component-per-file': 'off',
    'vue/html-closing-bracket-newline': 'off',
    'vue/max-attributes-per-line': 'off',
    'vue/multiline-html-element-content-newline': 'off',
    'vue/singleline-html-element-content-newline': 'off',
    'vue/attribute-hyphenation': 'off',
    'vue/require-default-prop': 'off',
    'vue/multi-word-component-names': 'off',
    'vue/no-dupe-keys': 'off',
};

module.exports = [
    {
        ignores: [
            '**/*.sh',
            '**/node_modules/**',
            '**/*.md',
            '**/*.woff',
            '**/*.ttf',
            '**/.vscode/**',
            '**/.idea/**',
            '**/dist/**',
            'public/**',
            'docs/**',
            '.husky/**',
            '.local/**',
            'bin/**',
            '.eslintrc.js',
            '.prettierrc.js',
            'src/mock/**',
            'src/assets/**',
        ],
    },
    ...vuePlugin.configs['flat/recommended'],
    {
        files: ['src/**/*.{js,ts,vue}'],
        languageOptions: {
            ecmaVersion: 2020,
            sourceType: 'module',
            globals,
        },
        plugins: {
            '@typescript-eslint': tsPlugin,
            prettier: prettierPlugin,
        },
        rules: {
            ...commonRules,
            ...vueRules,
        },
    },
    {
        files: ['src/**/*.ts'],
        languageOptions: {
            parser: tsParser,
            parserOptions: {
                ecmaVersion: 2020,
                sourceType: 'module',
                jsxPragma: 'React',
                ecmaFeatures: {
                    jsx: true,
                },
            },
        },
        rules: tsRules,
    },
    {
        files: ['src/**/*.vue'],
        languageOptions: {
            parser: vueParser,
            parserOptions: {
                parser: tsParser,
                ecmaVersion: 2020,
                sourceType: 'module',
                jsxPragma: 'React',
                ecmaFeatures: {
                    jsx: true,
                },
            },
        },
        rules: {
            ...tsRules,
            ...vueRules,
        },
    },
    eslintConfigPrettier,
];
