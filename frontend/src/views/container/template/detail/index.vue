<template>
    <DrawerPro v-model="detailVisible" :header="$t('commons.button.view')" size="large">
        <codemirror
            :autofocus="true"
            :placeholder="$t('commons.msg.noneData')"
            :indent-with-tab="true"
            :tabSize="4"
            style="width: 100%; height: calc(100vh - 160px)"
            :lineWrapping="true"
            :matchBrackets="true"
            theme="cobalt"
            :styleActiveLine="true"
            :extensions="extensions"
            v-model="detailInfo"
            :disabled="true"
        />
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="detailVisible = false">{{ $t('commons.button.cancel') }}</el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { ref } from 'vue';
import { Codemirror } from 'vue-codemirror';
const extensions = [javascript(), oneDark];

const detailVisible = ref(false);
const detailInfo = ref();

interface DialogProps {
    content: string;
}
const acceptParams = (params: DialogProps): void => {
    detailInfo.value = params.content;
    detailVisible.value = true;
};

defineExpose({
    acceptParams,
});
</script>
