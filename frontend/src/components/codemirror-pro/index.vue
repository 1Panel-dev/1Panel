<template>
    <div :style="customStyle">
        <div ref="editorRef" class="editor-container"></div>
    </div>
</template>

<script lang="ts" setup>
import { CSSProperties } from 'vue';
import { basicSetup, EditorView } from 'codemirror';
import { EditorState } from '@codemirror/state';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { StreamLanguage } from '@codemirror/language';
import { nginx } from '@codemirror/legacy-modes/mode/nginx';
import { yaml } from '@codemirror/legacy-modes/mode/yaml';
import { dockerFile } from '@codemirror/legacy-modes/mode/dockerfile';
import { placeholder } from '@codemirror/view';
import { json } from '@codemirror/lang-json';

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
    heightDiff: {
        type: Number,
        default: 200,
    },
    minHeight: {
        type: Number,
        default: 400,
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

const initCodeMirror = () => {
    const defaultTheme = EditorView.theme({
        '&.cm-editor': {
            minHeight: props.minHeight + 'px',
            height: 'calc(100vh - ' + props.heightDiff + 'px)',
        },
    });

    const extensions = [
        defaultTheme,
        oneDark,
        basicSetup,
        EditorView.updateListener.of((v: any) => {
            if (v.docChanged) {
                emit('update:modelValue', v.state.doc.toString());
            }
        }),
        placeholder(props.placeholder),
        EditorView.editable.of(!props.disabled),
    ];
    switch (props.mode) {
        case 'dockerfile':
            extensions.push(StreamLanguage.define(dockerFile));
            break;
        case 'javascript':
            extensions.push(javascript());
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
