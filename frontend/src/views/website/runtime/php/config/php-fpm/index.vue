<template>
    <div v-loading="loading">
        <CodemirrorPro v-model="content"></CodemirrorPro>
        <el-button type="primary" @click="openUpdate()" class="mt-2.5">
            {{ $t('nginx.saveAndReload') }}
        </el-button>
    </div>
</template>
<script lang="ts" setup>
import { GetPHPConfigFile, UpdatePHPFile } from '@/api/modules/runtime';
import { computed, onMounted, ref } from 'vue';
import { File } from '@/api/interface/file';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
    type: {
        type: String,
        default: 'fpm',
    },
});

const id = computed(() => {
    return props.id;
});

const data = ref<File.File>();
const loading = ref(false);
const content = ref('');

const get = () => {
    loading.value = true;
    GetPHPConfigFile({ id: id.value, type: props.type })
        .then((res) => {
            data.value = res.data;
            content.value = data.value.content;
        })
        .finally(() => {
            loading.value = false;
        });
};

const openUpdate = async () => {
    const action = await ElMessageBox.confirm(
        i18n.global.t('database.restartNowHelper'),
        i18n.global.t('database.confChange'),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        },
    );
    if (action === 'confirm') {
        loading.value = true;
        submit();
    }
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
