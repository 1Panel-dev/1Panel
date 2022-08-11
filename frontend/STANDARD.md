# 项目规范文档 📚

## 一、项目文件、组件命名规范

> **完全采用 Vue 官方推荐的风格指南，请务必查看 💢**
>
> **Link：** https://v3.cn.vuejs.org/style-guide

## 二、代码格式化工具（Prettier）

### 1、下载安装 prettier：

```text
npm install prettier
```

### 2、安装 Vscode 插件（Prettier）：

![Prettier](https://iamge-1259297738.cos.ap-chengdu.myqcloud.com/md/Prettier.png)

### 3、配置 Prettier：

```javascript
// @see: https://www.prettier.cn

module.exports = {
    // 超过最大值换行
    printWidth: 130,
    // 缩进字节数
    tabWidth: 2,
    // 使用制表符而不是空格缩进行
    useTabs: true,
    // 结尾不用分号(true有，false没有)
    semi: true,
    // 使用单引号(true单双引号，false双引号)
    singleQuote: false,
    // 更改引用对象属性的时间 可选值"<as-needed|consistent|preserve>"
    quoteProps: 'as-needed',
    // 在对象，数组括号与文字之间加空格 "{ foo: bar }"
    bracketSpacing: true,
    // 多行时尽可能打印尾随逗号。（例如，单行数组永远不会出现逗号结尾。） 可选值"<none|es5|all>"，默认none
    trailingComma: 'none',
    // 在JSX中使用单引号而不是双引号
    jsxSingleQuote: false,
    //  (x) => {} 箭头函数参数只有一个时是否要有小括号。avoid：省略括号 ,always：不省略括号
    arrowParens: 'avoid',
    // 如果文件顶部已经有一个 doclock，这个选项将新建一行注释，并打上@format标记。
    insertPragma: false,
    // 指定要使用的解析器，不需要写文件开头的 @prettier
    requirePragma: false,
    // 默认值。因为使用了一些折行敏感型的渲染器（如GitHub comment）而按照markdown文本样式进行折行
    proseWrap: 'preserve',
    // 在html中空格是否是敏感的 "css" - 遵守CSS显示属性的默认值， "strict" - 空格被认为是敏感的 ，"ignore" - 空格被认为是不敏感的
    htmlWhitespaceSensitivity: 'css',
    // 换行符使用 lf 结尾是 可选值"<auto|lf|crlf|cr>"
    endOfLine: 'auto',
    // 这两个选项可用于格式化以给定字符偏移量（分别包括和不包括）开始和结束的代码
    rangeStart: 0,
    rangeEnd: Infinity,
    // Vue文件脚本和样式标签缩进
    vueIndentScriptAndStyle: false,
};
```

## 三、代码规范工具（ESLint）

### 1、安装 ESLint 相关插件：

```text
npm install eslint eslint-config-prettier eslint-plugin-prettier eslint-plugin-vue @typescript-eslint/eslint-plugin @typescript-eslint/parser -D
```

|               依赖               |                               作用描述                               |
| :------------------------------: | :------------------------------------------------------------------: |
|              eslint              |                            ESLint 核心库                             |
|      eslint-config-prettier      |               关掉所有和 Prettier 冲突的 ESLint 的配置               |
|      eslint-plugin-prettier      |         将 Prettier 的 rules 以插件的形式加入到 ESLint 里面          |
|        eslint-plugin-vue         |                      为 Vue 使用 ESlint 的插件                       |
| @typescript-eslint/eslint-plugin |      ESLint 插件，包含了各类定义好的检测 TypeScript 代码的规范       |
|    @typescript-eslint/parser     | ESLint 的解析器，用于解析 TypeScript，从而检查和规范 TypeScript 代码 |

### 2、安装 Vscode 插件（ESLint、TSLint）：

-   **ESLint：**

![ESLint](https://iamge-1259297738.cos.ap-chengdu.myqcloud.com/md/ESLint.png)

### 3、配置 ESLint：

```javascript
// @see: http://eslint.cn

module.exports = {
    root: true,
    env: {
        browser: true,
        node: true,
        es6: true,
    },
    /* 指定如何解析语法 */
    parser: 'vue-eslint-parser',
    /* 优先级低于 parse 的语法解析配置 */
    parserOptions: {
        parser: '@typescript-eslint/parser',
        ecmaVersion: 2020,
        sourceType: 'module',
        jsxPragma: 'React',
        ecmaFeatures: {
            jsx: true,
        },
    },
    /* 继承某些已有的规则 */
    extends: [
        'plugin:vue/vue3-recommended',
        'plugin:@typescript-eslint/recommended',
        'prettier',
        'plugin:prettier/recommended',
    ],
    /*
     * "off" 或 0    ==>  关闭规则
     * "warn" 或 1   ==>  打开的规则作为警告（不影响代码执行）
     * "error" 或 2  ==>  规则作为一个错误（代码不能执行，界面报错）
     */
    rules: {
        // eslint (http://eslint.cn/docs/rules)
        'no-var': 'error', // 要求使用 let 或 const 而不是 var
        'no-multiple-empty-lines': ['error', { max: 1 }], // 不允许多个空行
        'no-use-before-define': 'off', // 禁止在 函数/类/变量 定义之前使用它们
        'prefer-const': 'off', // 此规则旨在标记使用 let 关键字声明但在初始分配后从未重新分配的变量，要求使用 const
        'no-irregular-whitespace': 'off', // 禁止不规则的空白

        // typeScript (https://typescript-eslint.io/rules)
        '@typescript-eslint/no-unused-vars': 'error', // 禁止定义未使用的变量
        '@typescript-eslint/no-inferrable-types': 'off', // 可以轻松推断的显式类型可能会增加不必要的冗长
        '@typescript-eslint/no-namespace': 'off', // 禁止使用自定义 TypeScript 模块和命名空间。
        '@typescript-eslint/no-explicit-any': 'off', // 禁止使用 any 类型
        '@typescript-eslint/ban-ts-ignore': 'off', // 禁止使用 @ts-ignore
        '@typescript-eslint/ban-types': 'off', // 禁止使用特定类型
        '@typescript-eslint/explicit-function-return-type': 'off', // 不允许对初始化为数字、字符串或布尔值的变量或参数进行显式类型声明
        '@typescript-eslint/no-var-requires': 'off', // 不允许在 import 语句中使用 require 语句
        '@typescript-eslint/no-empty-function': 'off', // 禁止空函数
        '@typescript-eslint/no-use-before-define': 'off', // 禁止在变量定义之前使用它们
        '@typescript-eslint/ban-ts-comment': 'off', // 禁止 @ts-<directive> 使用注释或要求在指令后进行描述
        '@typescript-eslint/no-non-null-assertion': 'off', // 不允许使用后缀运算符的非空断言(!)
        '@typescript-eslint/explicit-module-boundary-types': 'off', // 要求导出函数和类的公共类方法的显式返回和参数类型

        // vue (https://eslint.vuejs.org/rules)
        'vue/script-setup-uses-vars': 'error', // 防止<script setup>使用的变量<template>被标记为未使用，此规则仅在启用该no-unused-vars规则时有效。
        'vue/v-slot-style': 'error', // 强制执行 v-slot 指令样式
        'vue/no-mutating-props': 'off', // 不允许组件 prop的改变（明天找原因）
        'vue/custom-event-name-casing': 'off', // 为自定义事件名称强制使用特定大小写
        'vue/attributes-order': 'off', // vue api使用顺序，强制执行属性顺序
        'vue/one-component-per-file': 'off', // 强制每个组件都应该在自己的文件中
        'vue/html-closing-bracket-newline': 'off', // 在标签的右括号之前要求或禁止换行
        'vue/max-attributes-per-line': 'off', // 强制每行的最大属性数
        'vue/multiline-html-element-content-newline': 'off', // 在多行元素的内容之前和之后需要换行符
        'vue/singleline-html-element-content-newline': 'off', // 在单行元素的内容之前和之后需要换行符
        'vue/attribute-hyphenation': 'off', // 对模板中的自定义组件强制执行属性命名样式
        'vue/require-default-prop': 'off', // 此规则要求为每个 prop 为必填时，必须提供默认值
        'vue/multi-word-component-names': 'off', // 要求组件名称始终为 “-” 链接的单词
    },
};
```

## 四、样式规范工具（StyleLint）

### 1、安装 StyleLint 相关插件：

```text
npm i stylelint stylelint-config-html stylelint-config-recommended-scss stylelint-config-recommended-vue stylelint-config-standard stylelint-config-standard-scss stylelint-config-recess-order postcss postcss-html stylelint-config-prettier -D
```

|               依赖                |                                                                     作用描述                                                                     |
| :-------------------------------: | :----------------------------------------------------------------------------------------------------------------------------------------------: |
|             stylelint             |                                                                 stylelint 核心库                                                                 |
|       stylelint-config-html       |                                  Stylelint 的可共享 HTML（和类似 HTML）配置，捆绑 postcss-html 并对其进行配置。                                  |
| stylelint-config-recommended-scss |                                         扩展 stylelint-config-recommended 共享配置，并为 SCSS 配置其规则                                         |
| stylelint-config-recommended-vue  |                                         扩展 stylelint-config-recommended 共享配置，并为 Vue 配置其规则                                          |
|     stylelint-config-standard     | 打开额外的规则来执行在规范和一些 CSS 样式指南中发现的通用约定，包括：惯用 CSS 原则，谷歌的 CSS 样式指南，Airbnb 的样式指南，和 @mdo 的代码指南。 |
|  stylelint-config-standard-scss   |                                          扩展 stylelint-config-standard 共享配置，并为 SCSS 配置其规则                                           |
|              postcss              |                                                              postcss-html 的依赖包                                                               |
|           postcss-html            |                                                   用于解析 HTML（和类似 HTML）的 PostCSS 语法                                                    |
|   stylelint-config-recess-order   |                                                               属性的排序（插件包）                                                               |
|     stylelint-config-prettier     |                                                   关闭所有不必要的或可能与 Prettier 冲突的规则                                                   |

### 2、安装 Vscode 插件（Stylelint）：

![Stylelint](https://iamge-1259297738.cos.ap-chengdu.myqcloud.com/md/Stylelint.png)

### 3、在目录的 .vscode 文件中新建 settings.json：

```json
{
    "editor.formatOnSave": true,
    "stylelint.enable": true,
    "editor.codeActionsOnSave": {
        "source.fixAll.stylelint": true
    },
    "stylelint.validate": ["css", "less", "postcss", "scss", "vue", "sass", "html"],
    "files.eol": "\n"
}
```

> 😎 也可以在 vscode 中全局配置上述 json 代码 😎

### 4、配置 stylelint.config.js

```javascript
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
```

## 五、EditorConfig 配置

### 1、简介

> **EditorConfig** 帮助开发人员在 **不同的编辑器** 和 **IDE** 之间定义和维护一致的编码样式。

### 2、安装 VsCode 插件（EditorConfig ）：

![editorConfig](https://iamge-1259297738.cos.ap-chengdu.myqcloud.com/img/20220510142005.png)

### 3、配置 EditorConfig：

```javascript
# http://editorconfig.org
root = true

[*] # 表示所有文件适用
charset = utf-8 # 设置文件字符集为 utf-8
end_of_line = lf # 控制换行类型(lf | cr | crlf)
insert_final_newline = true # 始终在文件末尾插入一个新行
indent_style = tab # 缩进风格（tab | space）
indent_size = 2 # 缩进大小
max_line_length = 130 # 最大行长度

[*.md] # 表示仅 md 文件适用以下规则
max_line_length = off # 关闭最大行长度限制
trim_trailing_whitespace = false # 关闭末尾空格修剪
```

## 六、Git 流程规范配置

|              依赖               |                                    作用描述                                    |
| :-----------------------------: | :----------------------------------------------------------------------------: |
|              husky              |           操作 **git** 钩子的工具（在 **git xx** 之前执行某些命令）            |
|           lint-staged           |  在提交之前进行 **eslint** 校验，并使用 **prettier** 格式化本地暂存区的代码，  |
|         @commitlint/cli         |             校验 **git commit** 信息是否符合规范，保证团队的一致性             |
| @commitlint/config-conventional |                             **Anglar** 的提交规范                              |
|           commitizen            | 基于 **Node.js** 的 **git commit** 命令行工具，生成标准化的 **commit message** |
|             cz-git              |    一款工程性更强，轻量级，高度自定义，标准输出格式的 **commitize** 适配器     |

### 1、husky（操作 git 钩子的工具）：

> **安装：**

```text
npm install husky -D
```

> **使用（为了添加.husky 文件夹）：**

```text
# 编辑 package.json > prepare 脚本并运行一次

npm set-script prepare "husky install"
npm run prepare
```

### 2、 lint-staged（本地暂存代码检查工具）

> **安装：**

```text
npm install lint-staged --D
```

> **添加 ESlint Hook（在.husky 文件夹下添加 pre-commit 文件）：**
>
> **作用：通过钩子函数，判断提交的代码是否符合规范，并使用 prettier 格式化代码**

```text
npx husky add .husky/pre-commit "npm run lint:lint-staged"
```

> 新增 **lint-staged.config.js** 文件：

```text
module.exports = {
	"*.{js,jsx,ts,tsx}": ["eslint --fix", "prettier --write"],
	"{!(package)*.json,*.code-snippets,.!(browserslist)*rc}": ["prettier --write--parser json"],
	"package.json": ["prettier --write"],
	"*.vue": ["eslint --fix", "prettier --write", "stylelint --fix"],
	"*.{scss,less,styl,html}": ["stylelint --fix", "prettier --write"],
	"*.md": ["prettier --write"]
};
```

### 3、commitlint（commit 信息校验工具，不符合则报错）

> **安装：**

```text
npm i @commitlint/cli @commitlint/config-conventional -D
```

> **配置命令（在.husky 文件夹下添加 commit-msg 文件）：**

```text
npx husky add .husky/commit-msg 'npx --no-install commitlint --edit "$1"'
```

### 4、commitizen（基于 Node.js 的 git commit 命令行工具，生成标准化的 message）

```text
// 安装 commitizen，如此一来可以快速使用 cz 或 git cz 命令进行启动。
npm install commitizen -D
```

### 5、cz-git

> **指定提交文字规范，一款工程性更强，高度自定义，标准输出格式的 commitizen 适配器**

```text
npm install cz-git -D
```

> **配置 package.json：**

```text
"config": {
  "commitizen": {
    "path": "node_modules/cz-git"
  }
}
```

> **新建 commitlint.config.js 文件：**

```javascript
// @see: https://cz-git.qbenben.com/zh/guide
/** @type {import('cz-git').UserConfig} */

module.exports = {
    ignores: [(commit) => commit.includes('init')],
    extends: ['@commitlint/config-conventional'],
    rules: {
        // @see: https://commitlint.js.org/#/reference-rules
        'body-leading-blank': [2, 'always'],
        'footer-leading-blank': [1, 'always'],
        'header-max-length': [2, 'always', 108],
        'subject-empty': [2, 'never'],
        'type-empty': [2, 'never'],
        'subject-case': [0],
        'type-enum': [
            2,
            'always',
            [
                'feat',
                'fix',
                'docs',
                'style',
                'refactor',
                'perf',
                'test',
                'build',
                'ci',
                'chore',
                'revert',
                'wip',
                'workflow',
                'types',
                'release',
            ],
        ],
    },
    prompt: {
        messages: {
            type: "Select the type of change that you're committing:",
            scope: 'Denote the SCOPE of this change (optional):',
            customScope: 'Denote the SCOPE of this change:',
            subject: 'Write a SHORT, IMPERATIVE tense description of the change:\n',
            body: 'Provide a LONGER description of the change (optional). Use "|" to break new line:\n',
            breaking: 'List any BREAKING CHANGES (optional). Use "|" to break new line:\n',
            footerPrefixsSelect: 'Select the ISSUES type of changeList by this change (optional):',
            customFooterPrefixs: 'Input ISSUES prefix:',
            footer: 'List any ISSUES by this change. E.g.: #31, #34:\n',
            confirmCommit: 'Are you sure you want to proceed with the commit above?',
            // 中文版
            // type: "选择你要提交的类型 :",
            // scope: "选择一个提交范围（可选）:",
            // customScope: "请输入自定义的提交范围 :",
            // subject: "填写简短精炼的变更描述 :\n",
            // body: '填写更加详细的变更描述（可选）。使用 "|" 换行 :\n',
            // breaking: '列举非兼容性重大的变更（可选）。使用 "|" 换行 :\n',
            // footerPrefixsSelect: "选择关联issue前缀（可选）:",
            // customFooterPrefixs: "输入自定义issue前缀 :",
            // footer: "列举关联issue (可选) 例如: #31, #I3244 :\n",
            // confirmCommit: "是否提交或修改commit ?"
        },
        types: [
            {
                value: 'feat',
                name: 'feat:     🚀  A new feature',
                emoji: '🚀',
            },
            {
                value: 'fix',
                name: 'fix:      🧩  A bug fix',
                emoji: '🧩',
            },
            {
                value: 'docs',
                name: 'docs:     📚  Documentation only changes',
                emoji: '📚',
            },
            {
                value: 'style',
                name: 'style:    🎨  Changes that do not affect the meaning of the code',
                emoji: '🎨',
            },
            {
                value: 'refactor',
                name: 'refactor: ♻️   A code change that neither fixes a bug nor adds a feature',
                emoji: '♻️',
            },
            {
                value: 'perf',
                name: 'perf:     ⚡️  A code change that improves performance',
                emoji: '⚡️',
            },
            {
                value: 'test',
                name: 'test:     ✅  Adding missing tests or correcting existing tests',
                emoji: '✅',
            },
            {
                value: 'build',
                name: 'build:    📦️   Changes that affect the build system or external dependencies',
                emoji: '📦️',
            },
            {
                value: 'ci',
                name: 'ci:       🎡  Changes to our CI configuration files and scripts',
                emoji: '🎡',
            },
            {
                value: 'chore',
                name: "chore:    🔨  Other changes that don't modify src or test files",
                emoji: '🔨',
            },
            {
                value: 'revert',
                name: 'revert:   ⏪️  Reverts a previous commit',
                emoji: '⏪️',
            },
            // 中文版
            // { value: "特性", name: "特性:   🚀  新增功能", emoji: "🚀" },
            // { value: "修复", name: "修复:   🧩  修复缺陷", emoji: "🧩" },
            // { value: "文档", name: "文档:   📚  文档变更", emoji: "📚" },
            // { value: "格式", name: "格式:   🎨  代码格式（不影响功能，例如空格、分号等格式修正）", emoji: "🎨" },
            // { value: "重构", name: "重构:   ♻️  代码重构（不包括 bug 修复、功能新增）", emoji: "♻️" },
            // { value: "性能", name: "性能:   ⚡️  性能优化", emoji: "⚡️" },
            // { value: "测试", name: "测试:   ✅  添加疏漏测试或已有测试改动", emoji: "✅" },
            // { value: "构建", name: "构建:   📦️  构建流程、外部依赖变更（如升级 npm 包、修改 webpack 配置等）", emoji: "📦️" },
            // { value: "集成", name: "集成:   🎡  修改 CI 配置、脚本", emoji: "🎡" },
            // { value: "回退", name: "回退:   ⏪️  回滚 commit", emoji: "⏪️" },
            // { value: "其他", name: "其他:   🔨  对构建过程或辅助工具和库的更改（不影响源文件、测试用例）", emoji: "🔨" }
        ],
        useEmoji: true,
        themeColorCode: '',
        scopes: [],
        allowCustomScopes: true,
        allowEmptyScopes: true,
        customScopesAlign: 'bottom',
        customScopesAlias: 'custom',
        emptyScopesAlias: 'empty',
        upperCaseSubject: false,
        allowBreakingChanges: ['feat', 'fix'],
        breaklineNumber: 100,
        breaklineChar: '|',
        skipQuestions: [],
        issuePrefixs: [{ value: 'closed', name: 'closed:   ISSUES has been processed' }],
        customIssuePrefixsAlign: 'top',
        emptyIssuePrefixsAlias: 'skip',
        customIssuePrefixsAlias: 'custom',
        allowCustomIssuePrefixs: true,
        allowEmptyIssuePrefixs: true,
        confirmColorize: true,
        maxHeaderLength: Infinity,
        maxSubjectLength: Infinity,
        minSubjectLength: 0,
        scopeOverrides: undefined,
        defaultBody: '',
        defaultIssues: '',
        defaultScope: '',
        defaultSubject: '',
    },
};
```
