import type { languages } from 'monaco-editor/editor';
import jsonWorker from 'monaco-editor/languages/features/json/json.worker?worker';
import cssWorker from 'monaco-editor/languages/features/css/css.worker?worker';
import htmlWorker from 'monaco-editor/languages/features/html/html.worker?worker';
import tsWorker from 'monaco-editor/languages/features/typescript/ts.worker?worker';
import EditorWorker from 'monaco-editor/editor/editor.worker?worker';

let initialized = false;
let languageSupportPromise: Promise<void> | null = null;

type MonacoEditorApi = typeof import('monaco-editor/editor');
type PythonLanguageModule = {
    conf: languages.LanguageConfiguration;
    language: languages.IMonarchLanguage;
};

export const buildPatchedPythonLanguage = (language: PythonLanguageModule['language']): languages.IMonarchLanguage => ({
    ...language,
    tokenizer: {
        ...language.tokenizer,
        strings: [
            [/[rR]?[fF]'''|[fF][rR]'''/, 'string.escape', '@fLongStringBody'],
            [/[rR]?[fF]"""|[fF][rR]"""/, 'string.escape', '@fLongDblStringBody'],
            ...language.tokenizer.strings,
        ],
        fLongStringBody: [
            [/'''/, 'string.escape', '@popall'],
            [/\{[^\}':!=]+/, 'identifier', '@fStringDetail'],
            [/\\./, 'string'],
            [/[^\\'\{\}]+/, 'string'],
            [/'/, 'string'],
            [/\\$/, 'string'],
        ],
        fLongDblStringBody: [
            [/"""/, 'string.escape', '@popall'],
            [/\{[^\}':!=]+/, 'identifier', '@fStringDetail'],
            [/\\./, 'string'],
            [/[^\\"\{\}]+/, 'string'],
            [/"/, 'string'],
            [/\\$/, 'string'],
        ],
    },
});

const patchPythonLanguageSupport = async () => {
    const [monaco, python] = await Promise.all([
        import('monaco-editor/editor'),
        import('monaco-editor/languages/definitions/python/python.js'),
    ]);
    const monacoApi = monaco as MonacoEditorApi;
    const pythonModule = python as PythonLanguageModule;

    monacoApi.languages.setLanguageConfiguration('python', pythonModule.conf);
    monacoApi.languages.setMonarchTokensProvider('python', buildPatchedPythonLanguage(pythonModule.language));
};

const patchPythonLanguageSupportSafely = async () => {
    try {
        await patchPythonLanguageSupport();
    } catch (error) {
        console.warn('Failed to patch Monaco Python language support:', error);
    }
};

export function setupMonacoEnvironment() {
    if (initialized) {
        return;
    }
    initialized = true;

    (
        self as typeof self & { MonacoEnvironment?: { getWorker: (_: string, label: string) => Worker } }
    ).MonacoEnvironment = {
        getWorker(_: string, label: string) {
            if (label === 'json') {
                return new jsonWorker();
            }
            if (label === 'css' || label === 'scss' || label === 'less') {
                return new cssWorker();
            }
            if (label === 'html' || label === 'handlebars' || label === 'razor') {
                return new htmlWorker();
            }
            if (['typescript', 'javascript'].includes(label)) {
                return new tsWorker();
            }
            return new EditorWorker();
        },
    };
}

export async function loadMonacoLanguageSupport() {
    if (!languageSupportPromise) {
        languageSupportPromise = Promise.all([
            import('monaco-editor/features/codicon/register'),
            import('monaco-editor/features/folding/register'),
            import('monaco-editor/features/contextmenu/register'),
            import('monaco-editor/features/clipboard/register'),
            import('monaco-editor/features/comment/register'),
            import('monaco-editor/features/dropOrPasteInto/register'),
            import('monaco-editor/features/find/register'),
            import('monaco-editor/features/multicursor/register'),
            import('monaco-editor/features/quickCommand/register'),
            import('monaco-editor/languages/definitions/register.all'),
            import('monaco-editor/languages/features/json/register'),
            import('monaco-editor/languages/features/css/register'),
            import('monaco-editor/languages/features/html/register'),
            import('monaco-editor/languages/features/typescript/register'),
        ])
            .then(() => patchPythonLanguageSupportSafely())
            .then(() => undefined);
    }

    await languageSupportPromise;
}
