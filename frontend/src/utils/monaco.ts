import jsonWorker from 'monaco-editor/esm/vs/language/json/json.worker?worker';
import cssWorker from 'monaco-editor/esm/vs/language/css/css.worker?worker';
import htmlWorker from 'monaco-editor/esm/vs/language/html/html.worker?worker';
import tsWorker from 'monaco-editor/esm/vs/language/typescript/ts.worker?worker';
import EditorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker';

let initialized = false;
let languageSupportPromise: Promise<void> | null = null;

type MonacoEditorApi = typeof import('monaco-editor/esm/vs/editor/editor.api');
type PythonLanguageModule = typeof import('monaco-editor/esm/vs/basic-languages/python/python.js');

export const buildPatchedPythonLanguage = (language: PythonLanguageModule['language']) => ({
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
        import('monaco-editor/esm/vs/editor/editor.api'),
        import('monaco-editor/esm/vs/basic-languages/python/python.js'),
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
            import('monaco-editor/esm/vs/base/browser/ui/codicons/codiconStyles.js'),
            import('monaco-editor/esm/vs/editor/contrib/folding/browser/folding.js'),
            import('monaco-editor/esm/vs/editor/contrib/contextmenu/browser/contextmenu.js'),
            import('monaco-editor/esm/vs/editor/contrib/clipboard/browser/clipboard.js'),
            import('monaco-editor/esm/vs/editor/contrib/comment/browser/comment.js'),
            import('monaco-editor/esm/vs/editor/contrib/dropOrPasteInto/browser/copyPasteContribution.js'),
            import('monaco-editor/esm/vs/editor/contrib/find/browser/findController.js'),
            import('monaco-editor/esm/vs/editor/contrib/multicursor/browser/multicursor.js'),
            import('monaco-editor/esm/vs/editor/standalone/browser/quickAccess/standaloneCommandsQuickAccess.js'),
            import('monaco-editor/esm/vs/basic-languages/monaco.contribution'),
            import('monaco-editor/esm/vs/language/json/monaco.contribution'),
            import('monaco-editor/esm/vs/language/css/monaco.contribution'),
            import('monaco-editor/esm/vs/language/html/monaco.contribution'),
            import('monaco-editor/esm/vs/language/typescript/monaco.contribution'),
        ])
            .then(() => patchPythonLanguageSupportSafely())
            .then(() => undefined);
    }

    await languageSupportPromise;
}
