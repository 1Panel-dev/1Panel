<template>
    <el-drawer :close-on-click-modal="false" :close-on-press-escape="false" v-model="open" size="30%">
        <template #header>
            <Header :header="$t('app.ignoreList')" :back="handleClose"></Header>
        </template>
        <el-row :gutter="5">
            <el-col v-for="(app, index) in apps" :key="index">
                <el-card class="app-margin">
                    <el-row :gutter="20">
                        <el-col :span="6">
                            <el-avatar shape="square" :size="60" :src="'data:image/png;base64,' + app.icon" />
                        </el-col>
                        <el-col :span="12">
                            <span>{{ app.name }}</span>
                            <div class="app-margin">
                                <el-tag>{{ app.version }}</el-tag>
                            </div>
                        </el-col>
                        <el-col :span="6">
                            <el-button type="primary" link @click="cancelIgnore(app.detailID)">
                                {{ $t('app.cancelIgnore') }}
                            </el-button>
                        </el-col>
                    </el-row>
                </el-card>
            </el-col>
        </el-row>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
            </span>
        </template>
    </el-drawer>
</template>
<script lang="ts" setup>
import { GetIgnoredApp, IgnoreUpgrade } from '@/api/modules/app';
import { ref } from 'vue';
import Header from '@/components/drawer-header/index.vue';
import { MsgSuccess } from '@/utils/message';
import i18n from '@/lang';

const open = ref(false);
const loading = ref(false);
const apps = ref();
const em = defineEmits(['close']);

const handleClose = () => {
    open.value = false;
    em('close', open);
};

const acceptParams = () => {
    open.value = true;
    getApps();
};

const getApps = async () => {
    try {
        const res = await GetIgnoredApp();
        apps.value = res.data;
    } catch (error) {}
};

const cancelIgnore = async (id: number) => {
    loading.value = true;
    await IgnoreUpgrade({ detailID: id, operate: 'cancel' })
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .finally(() => {
            getApps();
            loading.value = false;
        });
};

defineExpose({
    acceptParams,
});
</script>

<style scoped>
.app-margin {
    margin-top: 10px;
}
</style>
