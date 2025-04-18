<template>
    <DrawerPro v-model="open" :header="$t('app.ignoreList')" @close="handleClose" size="small">
        <template #content>
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
                                    <el-tag v-if="app.version != ''">{{ app.version }}</el-tag>
                                    <el-tag v-else>{{ $t('commons.table.all') + $t('app.version') }}</el-tag>
                                </div>
                            </el-col>
                            <el-col :span="6">
                                <el-button type="primary" link @click="cancelIgnore(app.ID)">
                                    {{ $t('app.cancelIgnore') }}
                                </el-button>
                            </el-col>
                        </el-row>
                    </el-card>
                </el-col>
            </el-row>
        </template>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
            </span>
        </template>
    </DrawerPro>
</template>
<script lang="ts" setup>
import { cancelAppIgnore, getIgnoredApp } from '@/api/modules/app';
import { ref } from 'vue';
import { MsgSuccess } from '@/utils/message';
import i18n from '@/lang';
import bus from '@/global/bus';

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
        const res = await getIgnoredApp();
        apps.value = res.data;
    } catch (error) {}
};

const cancelIgnore = async (id: number) => {
    loading.value = true;
    await cancelAppIgnore({ id: id })
        .then(() => {
            bus.emit('upgrade', true);
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
