<template>
    <div v-loading="loading">
        <CodemirrorPro v-model="content" mode="nginx" />
        <el-button type="primary" @click="submit()" class="mt-2.5">
            {{ $t('nginx.saveAndReload') }}
        </el-button>
    </div>
</template>
<script lang="ts" setup>
import { getWebsiteConfig, updateNginxFile } from '@/api/modules/website';
import { computed, onMounted, ref } from 'vue';
import { File } from '@/api/interface/file';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import CodemirrorPro from '@/components/codemirror-pro/index.vue';

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
    getWebsiteConfig(id.value, 'openresty')
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
    updateNginxFile({
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
