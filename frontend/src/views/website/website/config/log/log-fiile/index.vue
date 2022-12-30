<template>
    <div v-loading="loading">
        <el-form-item prop="enable" :label="$t('website.enable')">
            <el-switch v-model="data.enable" @change="updateEnable"></el-switch>
        </el-form-item>
        <codemirror
            style="max-height: 500px; width: 100%; min-height: 200px"
            :autofocus="true"
            :placeholder="$t('website.noLog')"
            :indent-with-tab="true"
            :tabSize="4"
            :lineWrapping="true"
            :matchBrackets="true"
            theme="cobalt"
            :styleActiveLine="true"
            :extensions="extensions"
            v-model="data.content"
            :readOnly="true"
        />
    </div>
</template>
<script lang="ts" setup>
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { computed, onMounted, ref } from 'vue';
import { OpWebsiteLog } from '@/api/modules/website';

const extensions = [javascript(), oneDark];
const props = defineProps({
    logType: {
        type: String,
        default: '',
    },
    id: {
        type: Number,
        default: 0,
    },
});
const logType = computed(() => {
    return props.logType;
});
const id = computed(() => {
    return props.id;
});
let loading = ref(false);
let data = ref({
    enable: false,
    content: '',
});

const getContent = () => {
    const req = {
        id: id.value,
        operate: 'get',
        logType: logType.value,
    };
    loading.value = true;
    OpWebsiteLog(req)
        .then((res) => {
            data.value = res.data;
        })
        .finally(() => {
            loading.value = false;
        });
};

const updateEnable = () => {
    const operate = data.value.enable ? 'enable' : 'disable';
    const req = {
        id: id.value,
        operate: operate,
        logType: logType.value,
    };
    loading.value = true;
    OpWebsiteLog(req)
        .then(() => {
            getContent();
        })
        .finally(() => {
            loading.value = false;
        });
};

onMounted(() => {
    getContent();
});
</script>
