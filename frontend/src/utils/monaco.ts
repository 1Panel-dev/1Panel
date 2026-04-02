import jsonWorker from 'monaco-editor/esm/vs/language/json/json.worker?worker';
import cssWorker from 'monaco-editor/esm/vs/language/css/css.worker?worker';
import htmlWorker from 'monaco-editor/esm/vs/language/html/html.worker?worker';
import tsWorker from 'monaco-editor/esm/vs/language/typescript/ts.worker?worker';
import EditorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker';

let initialized = false;

export function setupMonacoEnvironment() {
    if (initialized) {
        return;
    }
    initialized = true;

    self.MonacoEnvironment = {
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
