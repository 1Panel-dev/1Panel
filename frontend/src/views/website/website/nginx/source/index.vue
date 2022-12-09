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
        <div style="margin-top: 10px">
            <el-button @click="getDefaultConfig()" :loading="loading">
                {{ $t('app.defaultConfig') }}
            </el-button>
            <el-button type="primary" @click="submit()" :loading="loading">
                {{ $t('commons.button.save') }}
            </el-button>
        </div>
        <el-row>
            <el-col :span="4">
                <el-alert
                    v-if="useOld"
                    style="margin-top: 10px"
                    :title="$t('app.defaultConfigHelper')"
                    type="info"
                    :closable="false"
                ></el-alert>
            </el-col>
        </el-row>
    </div>
</template>
<script lang="ts" setup>
import { GetNginx, UpdateNginxConfigFile } from '@/api/modules/nginx';
import { onMounted, ref } from 'vue';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { ElMessage } from 'element-plus';
import i18n from '@/lang';
import { GetAppDefaultConfig } from '@/api/modules/app';

const extensions = [javascript(), oneDark];

let data = ref();
let content = ref('');
let loading = ref(false);
let useOld = ref(false);

const submit = () => {
    loading.value = true;
    UpdateNginxConfigFile({
        filePath: data.value.path,
        content: content.value,
        backup: useOld.value,
    })
        .then(() => {
            ElMessage.success(i18n.global.t('commons.msg.updateSuccess'));
            getNginx();
        })
        .finally(() => {
            loading.value = false;
        });
};

const getNginx = async () => {
    const res = await GetNginx();
    data.value = res.data;
    content.value = data.value.content;
    useOld.value = false;
};

const getDefaultConfig = async () => {
    loading.value = true;
    const res = await GetAppDefaultConfig('nginx');
    content.value = res.data;
    useOld.value = true;
    loading.value = false;
};

onMounted(() => {
    getNginx();
});
</script>
