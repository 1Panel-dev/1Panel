// @see: https://stylelint.io

module.exports = {
    /* 继承某些已有的规则 */
    extends: [
        'stylelint-config-standard', // 配置stylelint拓展插件
        'stylelint-config-html/vue', // 配置 vue 中 template 样式格式化
        'stylelint-config-standard-scss', // 配置stylelint scss插件
        'stylelint-config-recommended-vue/scss', // 配置 vue 中 scss 样式格式化
        'stylelint-config-recess-order', // 配置stylelint css属性书写顺序插件,
        'stylelint-config-prettier', // 配置stylelint和prettier兼容
    ],
    overrides: [
        // 扫描 .vue/html 文件中的<style>标签内的样式
        {
            files: ['**/*.{vue,html}'],
            customSyntax: 'postcss-html',
        },
    ],
    /**
     * null  => 关闭该规则
     */
    rules: {
        'no-descending-specificity': null, // 禁止在具有较高优先级的选择器后出现被其覆盖的较低优先级的选择器
        'function-url-quotes': 'always', // 要求或禁止 URL 的引号 "always(必须加上引号)"|"never(没有引号)"
        'string-quotes': 'double', // 指定字符串使用单引号或双引号
        'unit-case': null, // 指定单位的大小写 "lower(全小写)"|"upper(全大写)"
        'color-hex-case': 'lower', // 指定 16 进制颜色的大小写 "lower(全小写)"|"upper(全大写)"
        'color-hex-length': 'long', // 指定 16 进制颜色的简写或扩写 "short(16进制简写)"|"long(16进制扩写)"
        'rule-empty-line-before': 'never', // 要求或禁止在规则之前的空行 "always(规则之前必须始终有一个空行)"|"never(规则前绝不能有空行)"|"always-multi-line(多行规则之前必须始终有一个空行)"|"never-multi-line(多行规则之前绝不能有空行。)"
        'font-family-no-missing-generic-family-keyword': null, // 禁止在字体族名称列表中缺少通用字体族关键字
        'block-opening-brace-space-before': 'always', // 要求在块的开大括号之前必须有一个空格或不能有空白符 "always(大括号前必须始终有一个空格)"|"never(左大括号之前绝不能有空格)"|"always-single-line(在单行块中的左大括号之前必须始终有一个空格)"|"never-single-line(在单行块中的左大括号之前绝不能有空格)"|"always-multi-line(在多行块中，左大括号之前必须始终有一个空格)"|"never-multi-line(多行块中的左大括号之前绝不能有空格)"
        'property-no-unknown': null, // 禁止未知的属性(true 为不允许)
        'no-empty-source': null, // 禁止空源码
        'declaration-block-trailing-semicolon': null, // 要求或不允许在声明块中使用尾随分号 string："always(必须始终有一个尾随分号)"|"never(不得有尾随分号)"
        'selector-class-pattern': null, // 强制选择器类名的格式
        'scss/at-import-partial-extension': null, // 解决不能引入scss文件
        'value-no-vendor-prefix': null, // 关闭 vendor-prefix(为了解决多行省略 -webkit-box)
        'selector-pseudo-class-no-unknown': [
            true,
            {
                ignorePseudoClasses: ['global', 'v-deep', 'deep'],
            },
        ],
    },
};
