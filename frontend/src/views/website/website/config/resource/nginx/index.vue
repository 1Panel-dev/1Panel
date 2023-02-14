<template>
    <div>
        <codemirror
            :autofocus="true"
            placeholder="None data"
            :indent-with-tab="true"
            :tabSize="4"
            style="margin-top: 10px; max-height: 700px"
            :lineWrapping="true"
            :matchBrackets="true"
            theme="cobalt"
            :styleActiveLine="true"
            :extensions="extensions"
            v-model="content"
            :readOnly="true"
        />
        <div style="margin-top: 10px">
            <el-button type="primary" @click="submit()" :loading="loading">
                {{ $t('nginx.saveAndReload') }}
            </el-button>
        </div>
    </div>
</template>
<script lang="ts" setup>
import { Codemirror } from 'vue-codemirror';
import { oneDark } from '@codemirror/theme-one-dark';
import { GetWebsiteNginx, UpdateNginxFile } from '@/api/modules/website';
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
});

const id = computed(() => {
    return props.id;
});

let data = ref<File.File>();
let loading = ref(false);
let content = ref('');

const get = () => {
    loading.value = true;
    GetWebsiteNginx(id.value)
        .then((res) => {
            data.value = res.data;
            content.value = data.value.content;
        })
        .finally(() => {
            loading.value = false;
        });
};

const submit = () => {
    loading.value = true;
    UpdateNginxFile({
        id: id.value,
        content: content.value,
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
