<template>
    <el-drawer
        v-model="detailVisible"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        size="50%"
    >
        <template #header>
            <DrawerHeader :header="$t('commons.button.view')" :back="handleClose" />
        </template>
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
    </el-drawer>
</template>

<script lang="ts" setup>
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { ref } from 'vue';
import { Codemirror } from 'vue-codemirror';
import DrawerHeader from '@/components/drawer-header/index.vue';
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

const handleClose = () => {
    detailVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
