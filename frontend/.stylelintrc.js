// @see: https://stylelint.io

module.exports = {
    extends: [
        'stylelint-config-standard', 
        'stylelint-config-html/vue', 
        'stylelint-config-standard-scss', 
        'stylelint-config-recommended-vue/scss', 
        'stylelint-config-recess-order', 
        'stylelint-config-prettier', 
    ],
    overrides: [
        {
            files: ['**/*.{vue,html}'],
            customSyntax: 'postcss-html',
        },
    ],
    rules: {
        'no-descending-specificity': null,
        'function-url-quotes': 'always',
        'string-quotes': 'double', 
        'unit-case': null, 
        'color-hex-case': 'lower',
        'color-hex-length': 'long',
        'rule-empty-line-before': 'never',
        'font-family-no-missing-generic-family-keyword': null,
        'block-opening-brace-space-before': 'always',
        'property-no-unknown': null,
        'no-empty-source': null,
        'declaration-block-trailing-semicolon': null,
        'selector-class-pattern': null,
        'scss/at-import-partial-extension': null,
        'value-no-vendor-prefix': null,
        'selector-pseudo-class-no-unknown': [
            true,
            {
                ignorePseudoClasses: ['global', 'v-deep', 'deep'],
            },
        ],
    },
};
