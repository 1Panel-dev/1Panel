<template>
    <div>
        <codemirror
            :autofocus="true"
            placeholder="None data"
            :indent-with-tab="true"
            :tabSize="4"
            style="margin-top: 10px; max-height: 500px"
            :lineWrapping="true"
            :matchBrackets="true"
            theme="cobalt"
            :styleActiveLine="true"
            :extensions="extensions"
            v-model="content"
            :readOnly="true"
        />
        <div style="float: right; margin-top: 10px">
            <el-button type="primary" @click="submit()" :loading="loading">
                {{ $t('commons.button.save') }}
            </el-button>
        </div>
    </div>
</template>
<script lang="ts" setup>
import { GetNginx } from '@/api/modules/nginx';
import { onMounted, ref } from 'vue';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { SaveFileContent } from '@/api/modules/files';
import { ElMessage } from 'element-plus';
import i18n from '@/lang';

const extensions = [javascript(), oneDark];

let data = ref();
let content = ref('');
let loading = ref(false);

const submit = () => {
    loading.value = true;
    SaveFileContent({
        path: data.value.path,
        content: content.value,
    })
        .then(() => {
            ElMessage.success(i18n.global.t('commons.msg.updateSuccess'));
        })
        .finally(() => {
            loading.value = false;
        });
};

const getNginx = async () => {
    const res = await GetNginx();
    data.value = res.data;
    content.value = data.value.content;
};

onMounted(() => {
    getNginx();
});
</script>
