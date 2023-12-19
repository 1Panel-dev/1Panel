<template>
    <div v-loading="loading">
        <codemirror
            :autofocus="true"
            :placeholder="$t('commons.msg.noneData')"
            :indent-with-tab="true"
            :tabSize="4"
            style="height: calc(100vh - 352px)"
            :lineWrapping="true"
            :matchBrackets="true"
            theme="cobalt"
            :styleActiveLine="true"
            :extensions="extensions"
            v-model="content"
        />
        <el-button type="primary" @click="openUpdate()" class="mt-2.5">
            {{ $t('nginx.saveAndReload') }}
        </el-button>
        <ConfirmDialog ref="confirmDialogRef" @confirm="submit()"></ConfirmDialog>
    </div>
</template>
<script lang="ts" setup>
import { Codemirror } from 'vue-codemirror';
import { oneDark } from '@codemirror/theme-one-dark';
import { GetWebsiteConfig, UpdatePHPFile } from '@/api/modules/website';
import { computed, onMounted, ref } from 'vue';
import { File } from '@/api/interface/file';
import i18n from '@/lang';
import { StreamLanguage } from '@codemirror/language';
import { nginx } from '@codemirror/legacy-modes/mode/nginx';
import { MsgSuccess } from '@/utils/message';

const extensions = [StreamLanguage.define(nginx), oneDark];

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
    type: {
        type: String,
        default: 'fpm',
    },
    installId: {
        type: Number,
        default: 0,
    },
});

const id = computed(() => {
    return props.id;
});

const data = ref<File.File>();
const loading = ref(false);
const content = ref('');
const confirmDialogRef = ref();

const get = () => {
    loading.value = true;
    GetWebsiteConfig(id.value, props.type)
        .then((res) => {
            data.value = res.data;
            content.value = data.value.content;
        })
        .finally(() => {
            loading.value = false;
        });
};

const openUpdate = async () => {
    confirmDialogRef.value!.acceptParams({
        header: i18n.global.t('database.confChange'),
        operationInfo: i18n.global.t('database.restartNowHelper'),
        submitInputInfo: i18n.global.t('database.restartNow'),
    });
};

const submit = async () => {
    loading.value = true;
    UpdatePHPFile({
        id: id.value,
        content: content.value,
        type: props.type,
    })
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
        })
        .finally(() => {
            loading.value = false;
        });
};

onMounted(() => {
    get();
});
</script>
