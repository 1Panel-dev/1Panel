<template>
    <DrawerPro v-model="drawerVisible" :header="$t('setting.license')" :resource="licenseName" @close="handleClose">
        <el-form label-position="top" :model="form" @submit.prevent v-loading="loading">
            <el-form-item :label="$t('license.add')" prop="nodeID">
                <el-select filterable multiple v-model="form.nodeIDs" clearable class="w-full">
                    <div v-for="item in unboundOptions" :key="item.id">
                        <el-option :label="item.name" :value="item.id" />
                    </div>
                </el-select>
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button :disabled="loading" type="primary" @click="onBind()">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DrawerPro>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { changeBind, listNodeOptions } from '@/api/modules/setting';
import { Setting } from '@/api/interface/setting';

interface DialogProps {
    licenseID: number;
    licenseName: string;
    freeNodes: Array<Setting.NodeItem>;
}
const drawerVisible = ref();
const loading = ref();
const licenseName = ref();
const unboundOptions = ref([]);

const form = reactive({
    licenseID: null,
    nodeIDs: [],
});

const acceptParams = (params: DialogProps): void => {
    licenseName.value = params.licenseName;
    form.licenseID = params.licenseID;
    form.nodeIDs = [];
    for (const item of params.freeNodes) {
        if (!item.isXpack) {
            form.nodeIDs.push(item.id);
        }
    }
    unboundOptions.value = [];
    loadNodes();
    drawerVisible.value = true;
};

const onBind = async () => {
    loading.value = true;
    await changeBind(form.licenseID, form.nodeIDs)
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            window.location.reload();
        })
        .catch(() => {
            loading.value = false;
        });
};

const loadNodes = async () => {
    await listNodeOptions('free').then((res) => {
        let nodeOptions = res.data || [];
        for (const item of nodeOptions) {
            if (!item.isBound) {
                unboundOptions.value.push(item);
                continue;
            }
            for (const item2 of form.nodeIDs) {
                if (item.id === item2) {
                    unboundOptions.value.push(item);
                }
            }
        }
    });
};

const handleClose = () => {
    drawerVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
