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
            <el-button type="primary" @click="submit()" :loading="loading">
                {{ $t('commons.button.save') }}
            </el-button>
        </div>
    </div>
</template>
<script lang="ts" setup>
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { GetWebsiteNginx } from '@/api/modules/website';
import { computed, onMounted, ref } from 'vue';
import { File } from '@/api/interface/file';
import { SaveFileContent } from '@/api/modules/files';
import { ElMessage } from 'element-plus';
import i18n from '@/lang';

const extensions = [javascript(), oneDark];

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

onMounted(() => {
    get();
});
</script>
