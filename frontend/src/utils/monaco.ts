import jsonWorker from 'monaco-editor/esm/vs/language/json/json.worker?worker';
import cssWorker from 'monaco-editor/esm/vs/language/css/css.worker?worker';
import htmlWorker from 'monaco-editor/esm/vs/language/html/html.worker?worker';
import tsWorker from 'monaco-editor/esm/vs/language/typescript/ts.worker?worker';
import EditorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker';

let initialized = false;
let languageSupportPromise: Promise<void> | null = null;

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
            import('monaco-editor/esm/vs/basic-languages/monaco.contribution'),
            import('monaco-editor/esm/vs/language/json/monaco.contribution'),
            import('monaco-editor/esm/vs/language/css/monaco.contribution'),
            import('monaco-editor/esm/vs/language/html/monaco.contribution'),
            import('monaco-editor/esm/vs/language/typescript/monaco.contribution'),
        ]).then(() => undefined);
    }

    await languageSupportPromise;
}
