<template>
    <div :style="customStyle">
        <div ref="editorRef" class="editor-container"></div>
    </div>
</template>

<script lang="ts" setup>
import { CSSProperties } from 'vue';
import { basicSetup, EditorView } from 'codemirror';
import { EditorState } from '@codemirror/state';
import { oneDark } from '@codemirror/theme-one-dark';
import { StreamLanguage } from '@codemirror/language';
import { nginx } from './nginx';
import { yaml } from '@codemirror/legacy-modes/mode/yaml';
import { shell } from '@codemirror/legacy-modes/mode/shell';
import { dockerFile } from '@codemirror/legacy-modes/mode/dockerfile';
import { javascript } from '@codemirror/legacy-modes/mode/javascript';
import { KeyBinding, placeholder } from '@codemirror/view';
import { json } from '@codemirror/lang-json';
import { keymap } from '@codemirror/view';
import { defaultKeymap, indentWithTab } from '@codemirror/commands';

defineOptions({ name: 'CodemirrorPro' });

const props = defineProps({
    disabled: {
        type: Boolean,
        default: false,
    },
    modelValue: {
        type: String,
        default: '',
    },
    mode: {
        type: String,
        default: 'javascript',
    },
    placeholder: {
        type: String,
        default: '',
    },
    height: {
        type: Number,
        default: 0,
    },
    heightDiff: {
        type: Number,
        default: 200,
    },
    minHeight: {
        type: Number,
        default: 400,
    },
    lineWrapping: {
        type: Boolean,
        default: false,
    },
});

const emit = defineEmits(['update:modelValue']);
const editorRef = ref();
const editorView = ref();
const content = computed(() => {
    return props.modelValue;
});

const customStyle = computed<CSSProperties>(() => ({
    width: '100%',
}));

const toggleLineComment = (mode: string) => {
    return (view: EditorView) => {
        const commentChar =
            mode === 'yaml' || mode === 'shell' || mode === 'nginx' || mode === 'dockerfile' ? '#' : '//';

        const { state } = view;

        const transaction = state.changeByRange((range) => {
            let line = state.doc.lineAt(range.from);
            let text = line.text;
            let changes;

            if (text.trim().startsWith(commentChar)) {
                const pos = line.from + text.indexOf(commentChar);
                changes = { from: pos, to: pos + commentChar.length, insert: '' };
            } else {
                changes = { from: line.from, insert: commentChar };
            }

            return {
                changes,
                range,
            };
        });

        view.dispatch(transaction);
        return true;
    };
};

const customKeymap: KeyBinding[] = [
    {
        key: 'Alt-/',
        run: toggleLineComment(props.mode),
        preventDefault: true,
    },
    {
        key: 'Mod-/',
        run: toggleLineComment(props.mode),
        preventDefault: true,
    },
];

const initCodeMirror = () => {
    const defaultTheme = EditorView.theme({
        '&.cm-editor': {
            minHeight: props.minHeight + 'px',
            height: loadHeight(),
        },
    });

    const extensions = [
        defaultTheme,
        oneDark,
        basicSetup,
        keymap.of([...defaultKeymap, indentWithTab, ...customKeymap]),
        EditorView.updateListener.of((v: any) => {
            if (v.docChanged) {
                emit('update:modelValue', v.state.doc.toString());
            }
        }),
        placeholder(props.placeholder),
        EditorView.editable.of(!props.disabled),
    ];

    if (props.lineWrapping) {
        extensions.push(EditorView.lineWrapping);
    }

    switch (props.mode) {
        case 'dockerfile':
            extensions.push(StreamLanguage.define(dockerFile));
            break;
        case 'javascript':
            extensions.push(StreamLanguage.define(javascript));
            break;
        case 'nginx':
            extensions.push(StreamLanguage.define(nginx));
            break;
        case 'yaml':
            extensions.push(StreamLanguage.define(yaml));
            break;
        case 'json':
            extensions.push(json());
            break;
        case 'shell':
            extensions.push(StreamLanguage.define(shell));
            break;
    }

    let startState = EditorState.create({
        doc: content.value,
        extensions: extensions,
    });

    editorView.value = new EditorView({
        state: startState,
        parent: editorRef.value,
    });
};

const loadHeight = () => {
    if (props.height || props.heightDiff) {
        return props.height ? props.height + 'px' : 'calc(100vh - ' + props.heightDiff + 'px)';
    }
};

watch(
    () => content.value,
    (newValue) => {
        if (editorView.value) {
            if (newValue === editorView.value.state.doc.toString()) {
                return;
            }
            editorView.value.dispatch({
                changes: {
                    from: 0,
                    to: editorView.value.state.doc.length,
                    insert: newValue,
                },
                scrollIntoView: false,
            });
        } else {
            initCodeMirror();
        }
    },
    { immediate: true },
);

onMounted(() => {
    initCodeMirror();
});

onUnmounted(() => {
    editorView.value?.destroy();
});
</script>
