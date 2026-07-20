# Translation Contribution Guide

Thank you for helping translate 1Panel! A complete locale touches the frontend,
the Core service, the Agent service, application-store metadata, and the project
README files. This guide lists the repository changes and validation required
before opening a pull request.

> **Reference pull requests**
>
> - [WIP: Dev v2 spanish (#10352)](https://github.com/1Panel-dev/1Panel/pull/10352)
>   is a full-locale example for Spanish (`es-ES`).
> - [docs: add Persian README (#13132)](https://github.com/1Panel-dev/1Panel/pull/13132)
>   is a localized-README example.

---

## Scope

1Panel has several independent translation surfaces:

| Surface | Technology | Repository format |
|---|---|---|
| Frontend | Vue 3, `vue-i18n`, Element Plus | TypeScript (`.ts`) |
| Core service | Go i18n | YAML (`.yaml`) |
| Agent service | Go i18n | YAML (`.yaml`) |
| Application store | API models and frontend types | Go and TypeScript |
| Project introduction | GitHub README files | Markdown (`.md`) |

Adding only a frontend module does not provide complete locale support. Follow
all applicable steps below and state explicitly in the pull request if a surface
is intentionally out of scope.

## Step 0 – Define locale identifiers

Choose the canonical runtime locale first. Use a
[BCP 47](https://www.ietf.org/rfc/bcp/bcp47.txt) language tag and preserve its
casing wherever the runtime locale is used. File names and third-party locale
packs do not always use the same casing.

| Language | Runtime locale | Frontend module | Element Plus pack | Backend YAML | App-store key | README |
|---|---|---|---|---|---|---|
| Simplified Chinese | `zh` | `zh.ts` | `zh-cn` | `zh.yaml` | `zh` | `README.zh-Hans.md` |
| Traditional Chinese | `zh-Hant` | `zh-Hant.ts` | `zh-tw` | `zh-Hant.yaml` | `zh-hant` | `README.zh-Hant.md` |
| English | `en` | `en.ts` | `en` | `en.yaml` | `en` | `README.md` |
| Brazilian Portuguese | `pt-BR` | `pt-br.ts` | `pt-br` | `pt-BR.yaml` | `pt-br` | `README.pt-br.md` |
| Spanish (Spain) | `es-ES` | `es-es.ts` | `es` | `es-ES.yaml` | `es-es` | `README.es-es.md` |
| Lao | `lo` | `lo.ts` | `lo` | `lo.yaml` | `lo` | `README.lo.md` |

Use lowercase for simple runtime codes (`ja`, `ko`, `lo`) and standard BCP 47
casing for variants (`pt-BR`, `es-ES`, `zh-Hant`). Do not infer a file name from
the runtime locale: add an explicit mapping when the casing or code differs.

Before editing, search for hard-coded locale collections so a newly introduced
selector, whitelist, or compatibility map is not missed:

```bash
rg "zh-Hant|pt-BR|es-ES" frontend core agent
```

## Step 1 – Add the frontend translation module

Copy the English module and translate its values:

```text
frontend/src/lang/modules/en.ts -> frontend/src/lang/modules/<module-name>.ts
```

Keep the following unchanged unless the English source changed structurally:

- object keys, nesting, and TypeScript syntax;
- interpolation placeholders such as `{0}`, `${name}`, and `{{ .detail }}`;
- HTML tags, Markdown, URLs, product names, commands, and configuration keys;
- the `getFuLocaleMessage` merge at the end of the file.

Pass the canonical runtime locale to `getFuLocaleMessage`, even when the module
file name uses different casing:

```ts
import { getFuLocaleMessage } from '@/lang/fu';

const message = {
    commons: {
        // Translated values.
    },
};

export default {
    ...getFuLocaleMessage('<runtime-locale>'),
    ...message,
};
```

The English module evolves continuously. Copy it from the same branch as your
change instead of relying on a fixed key count or line count. Machine
translation can be used for a first pass, but every string must be reviewed in
its UI context.

## Step 2 – Register the frontend loader and FU messages

Add the canonical runtime locale and its module file to `LOCALE_LOADERS` in
`frontend/src/lang/index.ts`:

```ts
const LOCALE_LOADERS: Record<string, LocaleLoader> = {
    // Existing entries.
    '<runtime-locale>': () => import('./modules/<module-name>'),
};
```

Then add the locale to `fuLocales` in `frontend/src/lang/fu.ts`. These messages
belong to the shared FU table and steps components and are not supplied by the
main module:

```ts
const fuLocales: Record<string, FuLocaleMessage> = {
    // Existing entries.
    '<runtime-locale>': {
        fu: {
            table: {
                more: '...',
                custom_table_rows: '...',
            },
            steps: {
                cancel: '...',
                prev: '...',
                next: '...',
                finish: '...',
            },
        },
    },
};
```

## Step 3 – Register the Element Plus locale

Element Plus provides its own translations for date pickers, pagination,
popovers, and other components. Import the matching pack in
`frontend/src/App.vue`:

```ts
import localePack from 'element-plus/es/locale/lang/<element-plus-code>';
```

Map the canonical runtime locale in the `i18nLocale` computed property:

```ts
const i18nLocale = computed(() => {
    // Existing mappings.
    if (globalStore.language === '<runtime-locale>') return localePack;
    return zhCn;
});
```

Confirm the pack exists in
`frontend/node_modules/element-plus/es/locale/lang/`. If Element Plus does not
provide the language, document the chosen fallback in the pull request rather
than importing a nonexistent module.

## Step 4 – Expose the locale in every frontend entry point

### Login page

`frontend/src/views/login/components/login-form.vue` currently renders two
language menus for different login layouts. Add the new locale to **both menu
blocks**, then add its native display name to `languageLabelMap`:

```ts
const languageLabelMap: Record<string, string> = {
    // Existing entries.
    '<runtime-locale>': '<Native language name>',
};
```

Verify both login layouts manually. Updating only `languageLabelMap` changes the
selected label but does not add a selectable menu item.

### Panel settings

Add the locale to `languageOptions` in
`frontend/src/views/setting/panel/index.vue`:

```ts
const languageOptions = ref([
    // Existing entries.
    { value: '<runtime-locale>', label: '<Native language name>' },
]);
```

### Public share page

Add the locale to `supportedLocales` in `frontend/src/views/share/index.vue`.
This list controls browser-language recognition and the `Accept-Language`
header used by public share requests.

```ts
const supportedLocales = [
    // Existing entries.
    '<runtime-locale>',
];
```

### Locale-sensitive fallbacks

Review locale comparisons that select language-specific API fields. For
example, operation records currently provide Chinese and English detail fields
in `frontend/src/views/log/operation/index.vue`. Chinese runtime locales should
use `detailZH`; every other locale should retain a visible `detailEN` fallback.
Prefer one shared fallback branch instead of adding every non-Chinese locale to
another hard-coded list.

## Step 5 – Allow the locale in the login API

The backend validates the language sent by the login request. Add the canonical
runtime locale to the `Language` field's `oneof` validation tag in
`core/app/dto/auth.go`:

```go
Language string `json:"language" validate:"oneof=zh en ... <runtime-locale>"`
```

Without this change, the new language can appear on the login page but the
login request will be rejected as an invalid parameter.

Run `gofmt` on the Go file after editing it.

## Step 6 – Add and register Core translations

Copy the current English catalog and translate only its values:

```text
core/i18n/lang/en.yaml -> core/i18n/lang/<runtime-locale>.yaml
```

Preserve every YAML key and placeholder:

```yaml
ErrInvalidParams: "Translated text: {{ .detail }}"
ErrRecordExist: "Translated text"
```

Register the catalog in `core/i18n/i18n.go`:

```go
var langFiles = map[string]string{
    // Existing entries.
    "<runtime-locale>": "lang/<runtime-locale>.yaml",
}
```

## Step 7 – Add and register Agent translations

Core and Agent are independent Go modules and have independent catalogs. Copy
and translate the Agent catalog separately:

```text
agent/i18n/lang/en.yaml -> agent/i18n/lang/<runtime-locale>.yaml
```

Register it in `agent/i18n/i18n.go`:

```go
var langFiles = map[string]string{
    // Existing entries.
    "<runtime-locale>": "lang/<runtime-locale>.yaml",
}
```

Do not copy the Core YAML file into Agent: their key sets are different.

## Step 8 – Extend application-store locale models

Application-store names and descriptions use their own locale fields. Add the
new field to both repository-side schemas.

In `agent/app/dto/app.go`, extend `Locale`:

```go
type Locale struct {
    // Existing fields.
    NewLanguage string `json:"<app-store-key>"`
}
```

In `frontend/src/api/interface/app.ts`, extend `Locale` with the same JSON key:

```ts
interface Locale {
    // Existing fields.
    '<app-store-key>': string;
}
```

Add the app-store key to the `description` example in
`core/cmd/server/app/app_config.yml` as well. This embedded template is written
when the server command scaffolds a new local application configuration.

Review `frontend/src/utils/app-store.ts` as well. Simple codes that lowercase to
the same value normally need no special handling; regional or script variants
may need an explicit runtime-locale-to-app-store-key mapping.

These schema changes make the repository capable of reading the locale. The
translated application names and descriptions are maintained in the external
application-store data source and must be coordinated separately. Until that
data exists, the application store falls back to English.

## Step 9 – Add a localized README

Copy the root README to the documentation directory:

```text
README.md -> docs/README.<readme-code>.md
```

Translate prose while preserving Markdown structure, HTML, images, links,
commands, badges, and code samples. Check every relative link from its new
location under `docs/`; a link that works from the repository root may need a
different relative path in the localized file.

Then keep the language navigation synchronized:

1. Add the new README link to the language badge row in `README.md`.
2. Add the same link to every existing `docs/README.*.md` language badge row.
3. Add the locale to the table in this guide's
   [Locale code reference](#locale-code-reference).

Use the native language name for the badge label. The README locale suffix may
follow an existing documentation convention (`zh-Hans`, `pt-br`, `es-es`) and
does not have to match the runtime locale casing.

## Step 10 – Format code and regenerate API documentation

Format every changed Go source file:

```bash
gofmt -w core/app/dto/auth.go core/i18n/i18n.go
gofmt -w agent/app/dto/app.go agent/i18n/i18n.go
```

The login-language enum and application-store locale model are represented in
generated Swagger files. Install the repository-compatible `swag` command, then
run the generator test from the Core module:

```bash
cd core
go test ./cmd/server/docs -run TestGenerateSwaggerDoc
```

This updates generated files under `core/cmd/server/docs/`. Do not edit the
generated Swagger output by hand. If the generator is unavailable, call that
out explicitly in the pull request instead of committing stale manual changes.

## Step 11 – Validate the locale

### Automated checks

Run checks from the repository root unless a command changes directory:

```bash
# Frontend type checking.
cd frontend
npm run type-check

# Lint only the frontend files changed by the locale contribution.
./node_modules/.bin/eslint \
    src/lang/modules/<module-name>.ts \
    src/lang/index.ts \
    src/lang/fu.ts \
    src/App.vue \
    src/views/login/components/login-form.vue \
    src/views/setting/panel/index.vue \
    src/views/share/index.vue \
    src/views/log/operation/index.vue \
    src/api/interface/app.ts \
    src/utils/app-store.ts

# Compile and test the directly affected Go packages.
cd ../core
go test ./i18n ./app/dto

cd ../agent
go test ./i18n ./app/dto

# Check whitespace errors in all changed files.
cd ..
git diff --check
```

Also compare each translated catalog with its English source:

- the frontend module has the same translation keys and nesting;
- each Core and Agent YAML file has exactly the same keys as its corresponding
  English file;
- placeholders, formatting tokens, HTML tags, URLs, and commands are preserved;
- YAML files parse successfully, with no duplicate keys;
- all README links and badges resolve to the intended files.

### Manual checks

- [ ] Both login layouts show the native language name and can select it.
- [ ] Login succeeds with the locale in the request; validation does not reject it.
- [ ] **Settings -> Panel -> Language** can switch to the locale and retains it after reload.
- [ ] Element Plus date, pagination, select, and confirmation components use the expected language.
- [ ] A representative Core API error is translated.
- [ ] A representative Agent task/error message is translated.
- [ ] A public share page recognizes the browser locale and sends the expected `Accept-Language` value.
- [ ] Operation-log details remain visible through the correct Chinese or English fallback.
- [ ] Application-store metadata uses the locale when available and falls back to English when absent.
- [ ] The localized README renders correctly and every README language badge links to it.
- [ ] Layout, punctuation, truncation, and text direction are correct at common screen sizes.

For right-to-left languages, verify the login page, navigation, forms, tables,
dialogs, and mixed-direction values such as IP addresses. RTL layout support may
require a separate frontend change beyond translating strings.

## External translation resources

Some runtime assets are not maintained in this repository:

- application-store names and descriptions come from the external
  application-store data source;
- the language shell resources downloaded as `language/lang.tar.gz` are
  published through the 1Panel resource service.

Repository locale support does not automatically update either resource. If the
new language needs them, coordinate publication with the relevant maintainer and
record the status in the pull request. Do not add generated or downloaded
archives to this repository unless a maintainer requests it.

## Complete pull-request checklist

- [ ] Canonical runtime, module, Element Plus, app-store, and README codes are documented.
- [ ] `frontend/src/lang/modules/<module-name>.ts` is complete and reviewed.
- [ ] `frontend/src/lang/index.ts` contains the locale loader.
- [ ] `frontend/src/lang/fu.ts` contains the FU messages.
- [ ] `frontend/src/App.vue` maps an existing Element Plus locale pack or documents a fallback.
- [ ] Both language menus and `languageLabelMap` are updated in `frontend/src/views/login/components/login-form.vue`.
- [ ] `languageOptions` is updated in `frontend/src/views/setting/panel/index.vue`.
- [ ] `supportedLocales` is updated in `frontend/src/views/share/index.vue`.
- [ ] Locale-sensitive API-field fallbacks, including operation-log details, remain visible.
- [ ] The login validator is updated in `core/app/dto/auth.go`.
- [ ] `core/i18n/lang/<runtime-locale>.yaml` exists and is registered in `core/i18n/i18n.go`.
- [ ] `agent/i18n/lang/<runtime-locale>.yaml` exists and is registered in `agent/i18n/i18n.go`.
- [ ] Application-store locale fields are updated in `agent/app/dto/app.go`, `frontend/src/api/interface/app.ts`, and `core/cmd/server/app/app_config.yml`.
- [ ] Regional or script variants are normalized in `frontend/src/utils/app-store.ts` when needed.
- [ ] `docs/README.<readme-code>.md` exists and every README language badge is synchronized.
- [ ] Changed Go files are formatted and Swagger output is regenerated.
- [ ] Frontend, Core, Agent, catalog-parity, placeholder, and Markdown checks pass.
- [ ] All applicable manual checks pass.
- [ ] External application-store and language-resource work is completed or clearly tracked.

## Locale code reference

| Language | Runtime locale |
|---|---|
| Simplified Chinese | `zh` |
| Traditional Chinese | `zh-Hant` |
| English | `en` |
| Japanese | `ja` |
| Korean | `ko` |
| Russian | `ru` |
| Malay | `ms` |
| Turkish | `tr` |
| Brazilian Portuguese | `pt-BR` |
| Spanish (Spain) | `es-ES` |
| Persian | `fa` |
| Lao | `lo` |

## Pull-request title convention

```text
feat(i18n): add <language name> (<runtime-locale>) locale support
```

Example: `feat(i18n): add German (de-DE) locale support`
