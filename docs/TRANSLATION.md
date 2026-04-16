# Translation Contribution Guide

Thank you for your interest in helping translate 1Panel into a new language! This guide walks through every file you need to create or modify to add full language support across the frontend and backend.

> **Reference PR:** [feat(i18n): add de-DE locale support (#12051)](https://github.com/1Panel-dev/1Panel/pull/12051) — a complete, real-world example of adding a brand-new locale (German).

---

## Overview

1Panel's i18n system spans two layers:

| Layer | Technology | Translation format |
|---|---|---|
| **Frontend** (Vue 3) | `vue-i18n` + Element Plus | TypeScript (`.ts`) |
| **Backend** (Go) | `go-i18n` / `nicksnyder/go-i18n` | YAML (`.yaml`) |

Adding a new language requires changes in **both** layers, as well as registering the locale in a few selector components.

---

## Step 1 – Frontend translation file

Copy the reference English file and translate all values.

```
frontend/src/lang/modules/en.ts  →  frontend/src/lang/modules/<locale>.ts
```

Replace every English string with its translation. Keep all keys, nested objects, and the `getFuLocaleMessage` call at the bottom unchanged:

```ts
// frontend/src/lang/modules/<locale>.ts
import { getFuLocaleMessage } from '@/lang/fu';

const message = {
    commons: {
        // ... translated strings
    },
    // ...
};

export default {
    ...getFuLocaleMessage('<locale>'),
    ...message,
};
```

> **Tip:** The file is ~4 400 lines. Using a translation tool for a first pass is fine, but please review the output for accuracy and context.

---

## Step 2 – Register the locale loader

Open `frontend/src/lang/index.ts` and add your locale to `LOCALE_LOADERS`:

```ts
// frontend/src/lang/index.ts
const LOCALE_LOADERS: Record<string, LocaleLoader> = {
    // existing entries …
    '<locale>': () => import('./modules/<locale>'),
};
```

---

## Step 3 – FU component translations

`frontend/src/lang/fu.ts` contains translations for custom FU table/steps components. Add a new entry for your locale:

```ts
// frontend/src/lang/fu.ts
const fuLocales: Record<string, FuLocaleMessage> = {
    // existing entries …
    '<locale>': {
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

---

## Step 4 – Element Plus component locale

Element Plus ships its own locale strings (used in date pickers, pagination, etc.). Import the matching locale pack and wire it into `frontend/src/App.vue`:

```ts
// frontend/src/App.vue  — import section
import <varName> from 'element-plus/es/locale/lang/<ep-locale-code>';
```

Then add a branch in the `i18nLocale` computed property:

```ts
const i18nLocale = computed(() => {
    // existing branches …
    if (globalStore.language === '<locale>') return <varName>;
    return zhCn;  // fallback unchanged
});
```

You can find all available Element Plus locale codes in `node_modules/element-plus/es/locale/lang/`.

---

## Step 5 – Login page language selector

Open `frontend/src/views/login/components/login-form.vue` and add your locale label to `languageLabelMap`:

```ts
const languageLabelMap: Record<string, string> = {
    // existing entries …
    '<locale>': '<Native language name>',
};
```

---

## Step 6 – Panel settings language selector

Open `frontend/src/views/setting/panel/index.vue` and add an option to `languageOptions`:

```ts
const languageOptions = ref([
    // existing entries …
    { value: '<locale>', label: '<Native language name>' },
]);
```

---

## Step 7 – Backend translation files (core & agent)

The backend has **two** independent Go modules, each with its own YAML translation file. Copy the English reference and translate:

```
core/i18n/lang/en.yaml   →  core/i18n/lang/<locale>.yaml   (~264 lines)
agent/i18n/lang/en.yaml  →  agent/i18n/lang/<locale>.yaml  (~592 lines)
```

YAML format example:

```yaml
ErrInvalidParams: "Invalid request parameters: {{ .detail }}"
ErrRecordExist: "Record already exists"
# …
```

---

## Step 8 – Register backend locale in Go

Add your locale key and file path in both i18n registries:

**`core/i18n/i18n.go`**
```go
var langFiles = map[string]string{
    // existing entries …
    "<locale>": "lang/<locale>.yaml",
}
```

**`agent/i18n/i18n.go`**
```go
var langFiles = map[string]string{
    // existing entries …
    "<locale>": "lang/<locale>.yaml",
}
```

---

## Checklist

Before opening a PR, verify the following:

- [ ] `frontend/src/lang/modules/<locale>.ts` created and all strings translated
- [ ] Locale registered in `frontend/src/lang/index.ts` (`LOCALE_LOADERS`)
- [ ] Locale entry added to `frontend/src/lang/fu.ts`
- [ ] Element Plus locale imported and mapped in `frontend/src/App.vue`
- [ ] Locale label added to `languageLabelMap` in `frontend/src/views/login/components/login-form.vue`
- [ ] Locale option added to `languageOptions` in `frontend/src/views/setting/panel/index.vue`
- [ ] `core/i18n/lang/<locale>.yaml` created
- [ ] `agent/i18n/lang/<locale>.yaml` created
- [ ] Locale registered in `core/i18n/i18n.go` (`langFiles`)
- [ ] Locale registered in `agent/i18n/i18n.go` (`langFiles`)
- [ ] Manually verified: new locale appears in the login language dropdown
- [ ] Manually verified: new locale appears in **Settings → Panel → Language**
- [ ] Manually verified: UI renders correctly after switching to the new locale

---

## Locale code conventions

Use [BCP 47](https://www.ietf.org/rfc/bcp/bcp47.txt) locale codes. Examples used in this project:

| Language | Locale code |
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

Use lowercase for simple codes (`ja`, `ko`) and the standard BCP 47 casing for regional variants (`pt-BR`, `es-ES`, `zh-Hant`).

**Frontend TypeScript file names** use all-lowercase with hyphens preserved exactly as they appear in the existing `frontend/src/lang/modules/` directory (e.g. `pt-br.ts`, `es-es.ts`, `zh-hant.ts`).

**Backend YAML file names** mirror the locale code exactly, preserving the original casing (e.g. `pt-BR.yaml`, `es-ES.yaml`, `zh-Hant.yaml`).

---

## PR title convention

```
feat(i18n): add <language name> (<locale>) locale support
```

Example: `feat(i18n): add German (de-DE) locale support`
