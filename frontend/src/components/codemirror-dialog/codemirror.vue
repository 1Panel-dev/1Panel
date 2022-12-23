<template>
    <el-dialog v-model="codeVisiable" :destroy-on-close="true" :close-on-click-modal="false" width="70%">
        <template #header>
            <div class="card-header">
                <span>{{ header }}</span>
            </div>
        </template>
        <codemirror
            ref="mymirror"
            :autofocus="true"
            placeholder="None data"
            :indent-with-tab="true"
            :tabSize="4"
            style="max-height: 500px; min-height: 200px; width: 100%"
            :lineWrapping="true"
            :matchBrackets="true"
            theme="cobalt"
            :styleActiveLine="true"
            :extensions="extensions"
            v-model="detailInfo"
            :readOnly="true"
        />
    </el-dialog>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';

const mymirror = ref();

const extensions = [javascript(), oneDark];
const header = ref();
const detailInfo = ref();
const codeVisiable = ref(false);

interface DialogProps {
    header: string;
    detailInfo: string;
}

const acceptParams = (props: DialogProps): void => {
    header.value = props.header;
    detailInfo.value = props.detailInfo;
    codeVisiable.value = true;
};

defineExpose({
    acceptParams,
});
</script>
