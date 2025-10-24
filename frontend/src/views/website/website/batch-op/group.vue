<template>
    <DrawerPro v-model="open" :header="$t('commons.table.group')" size="30%" @close="handleClose">
        <el-form ref="websiteForm" label-position="top" :model="form" :rules="rules" v-loading="loading">
            <GroupSelect v-model="form.groupID" :prop="'groupID'" :groupType="'website'"></GroupSelect>
        </el-form>
        <template #footer>
            <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
            <el-button type="primary" @click="submit()" :disabled="loading">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import GroupSelect from '@/views/website/website/components/group/index.vue';

import { batchSetGroup } from '@/api/modules/website';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { FormInstance } from 'element-plus';

const open = ref(false);
const loading = ref(false);
const websiteForm = ref<FormInstance>();
const form = reactive({
    ids: [] as number[],
    groupID: 0,
});
const rules = ref({
    groupID: [Rules.requiredSelect],
});

const handleClose = () => {
    open.value = false;
};

const acceptParams = async (ids: []) => {
    form.ids = ids;
    form.groupID = 0;
    open.value = true;
};

const submit = async () => {
    loading.value = true;
    const valid = await websiteForm.value.validate();
    if (!valid) {
        loading.value = false;
        return;
    }
    batchSetGroup(form)
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            open.value = false;
        })
        .finally(() => {
            loading.value = false;
        });
};

defineExpose({
    acceptParams,
});
</script>
